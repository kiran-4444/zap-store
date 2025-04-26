package bitcask

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync" // Import sync package
	"time"

	"github.com/gofrs/flock" // Import a file locking library
)

type DataDirFileLogEntry struct {
	crc       uint32
	timeStamp int64
	keySize   int64
	valueSize int64
	key       string
	value     string
}

func newDataDirFileLogEntry(key string, value string) *DataDirFileLogEntry {
	timeStamp := time.Now().UnixNano() // Use higher precision timestamp
	keySize := len(key)
	valueSize := len(value)
	// CRC should cover metadata + key + value for better integrity
	// Simplified version: just value CRC
	checksum := crc32.ChecksumIEEE([]byte(value))

	return &DataDirFileLogEntry{
		crc:       checksum,
		timeStamp: timeStamp,
		keySize:   int64(keySize),
		valueSize: int64(valueSize),
		key:       key,
		value:     value,
	}
}

func (ddfle *DataDirFileLogEntry) toBytes() ([]byte, error) {
	// Fixed header size: crc(4) + ts(8) + ksz(8) + vsz(8) = 28 bytes
	headerSize := 28
	buf := bytes.NewBuffer(make([]byte, 0, headerSize+int(ddfle.keySize)+int(ddfle.valueSize)))

	// Use BigEndian consistently
	if err := binary.Write(buf, binary.BigEndian, ddfle.crc); err != nil {
		return nil, fmt.Errorf("failed to write crc: %w", err)
	}
	if err := binary.Write(buf, binary.BigEndian, ddfle.timeStamp); err != nil {
		return nil, fmt.Errorf("failed to write timestamp: %w", err)
	}
	if err := binary.Write(buf, binary.BigEndian, ddfle.keySize); err != nil {
		return nil, fmt.Errorf("failed to write keySize: %w", err)
	}
	if err := binary.Write(buf, binary.BigEndian, ddfle.valueSize); err != nil {
		return nil, fmt.Errorf("failed to write valueSize: %w", err)
	}
	if _, err := buf.Write([]byte(ddfle.key)); err != nil {
		return nil, fmt.Errorf("failed to write key: %w", err)
	}
	if _, err := buf.Write([]byte(ddfle.value)); err != nil {
		return nil, fmt.Errorf("failed to write value: %w", err)
	}

	return buf.Bytes(), nil
}

type Log struct {
	writerPosition int64
	file           *os.File
	fileId         int64
	filePath       string
}

func openLogFile(dataDir string, fileId int64) (*Log, error) {
	fileName := fmt.Sprintf("%016d.log", fileId)
	filePath := filepath.Join(dataDir, fileName)

	// Use O_APPEND for efficient writes, O_RDWR needed for potential future ReadAt on active file
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("cannot open log file %s: %w", filePath, err)
	}

	// Get current size to set initial writerPosition (important for append mode)
	stat, err := file.Stat()
	if err != nil {
		file.Close() // Clean up on error
		return nil, fmt.Errorf("cannot stat log file %s: %w", filePath, err)
	}

	return &Log{
		writerPosition: stat.Size(), // Start writing at the end
		file:           file,
		fileId:         fileId,
		filePath:       filePath,
	}, nil
}

// setLogEntry writes to the log file. Called ONLY when holding the engine's write lock.
func (l *Log) setLogEntry(dataDirFileLogEntry *DataDirFileLogEntry) (valueStartOffset int64, entryEndOffset int64, err error) {
	bytesToWrite, err := dataDirFileLogEntry.toBytes()
	if err != nil {
		return -1, -1, fmt.Errorf("failed to serialize entry: %w", err)
	}

	// We opened with O_APPEND, so writes automatically go to the end.
	// No need for Seek before Write. The OS handles atomicity of positioning+write for APPEND.
	bytesWritten, err := l.file.Write(bytesToWrite)
	if err != nil {
		// Attempt to get current file size to know where the partial write *might* have ended
		currentSize, statErr := l.file.Seek(0, io.SeekEnd)
		if statErr != nil {
			// If we can't even get the size, the state is very uncertain
			return -1, -1, fmt.Errorf("failed to write entry (write error: %w, failed to get size after error: %v)", err, statErr)
		}
		// Update writerPosition even on error, assuming OS append guarantees some ordering
		l.writerPosition = currentSize
		return -1, -1, fmt.Errorf("failed to write entry: %w", err)
	}

	// Calculate offsets *after* successful write
	entryEndOffset = l.writerPosition + int64(bytesWritten)
	// valueStartOffset = entryEndOffset - dataDirFileLogEntry.valueSize
	// More robustly: value starts after header and key
	// Fixed header size: crc(4) + ts(8) + ksz(8) + vsz(8) = 28 bytes
	valueStartOffset = (l.writerPosition + 28 + dataDirFileLogEntry.keySize)

	// Update writerPosition *after* successful write
	l.writerPosition = entryEndOffset

	return valueStartOffset, entryEndOffset, nil
}

// getLogEntry reads ONLY the value from a specific offset in a *potentially inactive* file.
// It opens the file read-only on demand. Called when holding the engine's read lock.
func getLogValue(dataDir string, fileId int64, valueOffset int64, valueSize int64) (string, error) {
	// Construct file path (must match naming scheme used in openLogFile)
	fileName := fmt.Sprintf("%016d.log", fileId)
	filePath := filepath.Join(dataDir, fileName)

	// Open read-only
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0) // No need for 0755 on read-only
	if err != nil {
		// Handle file not found specifically?
		return "", fmt.Errorf("failed to open log file %s for reading: %w", filePath, err)
	}
	defer file.Close() // Ensure file is closed

	value := make([]byte, valueSize)

	// Use ReadAt for efficiency and correctness at specific offsets
	bytesRead, err := file.ReadAt(value, valueOffset)
	if err != nil {
		// io.EOF might be okay if valueOffset + valueSize == fileSize, but ReadAt handles this.
		// Return error on unexpected EOF or other read errors.
		return "", fmt.Errorf("failed reading value from %s at offset %d: %w", filePath, valueOffset, err)
	}

	if int64(bytesRead) != valueSize {
		return "", fmt.Errorf("short read: expected %d bytes, got %d from %s at offset %d",
			valueSize, bytesRead, filePath, valueOffset)
	}

	return string(value), nil
}

// Close closes the underlying file handle. Called when holding the engine's write lock (e.g., during rotation or engine Close).
func (l *Log) Close() error {
	if l.file != nil {
		err := l.file.Close()
		l.file = nil // Prevent double close
		if err != nil {
			return fmt.Errorf("failed to close log file %s: %w", l.filePath, err)
		}
	}
	return nil
}

// --- KeyDir entry remains the same ---
type KeyDir struct {
	fileId        int64
	valueSize     int64
	valuePosition int64 // Position where the VALUE starts
	timeStamp     int64 // Use UnixNano for better resolution
}

// readEntry reads a full entry (header, key, value) from a given position. Used for KeyDir rebuild.
func readEntry(f *os.File, position int64) (*DataDirFileLogEntry, int64, error) {
	// Fixed header size
	// Fixed header size: crc(4) + ts(8) + ksz(8) + vsz(8) = 28 bytes
	headerSize := int64(28)

	// Seek to the start of the entry
	_, err := f.Seek(position, io.SeekStart)
	if err != nil {
		// Check for EOF specifically when seeking - indicates end of file reached cleanly
		if err == io.EOF {
			return nil, 0, io.EOF
		}
		return nil, 0, fmt.Errorf("seek failed at pos %d: %w", position, err)
	}

	header := make([]byte, headerSize)
	_, err = io.ReadFull(f, header)
	if err != nil {
		// Distinguish between clean EOF (tried to read header past end) and other errors
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return nil, 0, io.EOF // Treat unexpected EOF same as clean EOF here
		}
		return nil, 0, fmt.Errorf("failed reading header at pos %d: %w", position, err)
	}

	buf := bytes.NewReader(header)
	var entry DataDirFileLogEntry

	// Read header fields
	if err := binary.Read(buf, binary.BigEndian, &entry.crc); err != nil {
		return nil, 0, fmt.Errorf("failed decoding crc at pos %d: %w", position, err)
	}
	if err := binary.Read(buf, binary.BigEndian, &entry.timeStamp); err != nil {
		return nil, 0, fmt.Errorf("failed decoding timestamp at pos %d: %w", position, err)
	}
	if err := binary.Read(buf, binary.BigEndian, &entry.keySize); err != nil {
		return nil, 0, fmt.Errorf("failed decoding keySize at pos %d: %w", position, err)
	}
	if err := binary.Read(buf, binary.BigEndian, &entry.valueSize); err != nil {
		return nil, 0, fmt.Errorf("failed decoding valueSize at pos %d: %w", position, err)
	}

	// Basic sanity check
	if entry.keySize < 0 || entry.valueSize < 0 || entry.keySize > 1<<20 || entry.valueSize > 1<<20 {
		return nil, 0, fmt.Errorf("invalid entry size (ksz=%d, vsz=%d) at pos %d", entry.keySize, entry.valueSize, position)
	}

	keyBytes := make([]byte, entry.keySize)
	_, err = io.ReadFull(f, keyBytes)
	if err != nil {
		return nil, 0, fmt.Errorf("failed reading key (%d bytes) at pos %d: %w", entry.keySize, position+headerSize, err)
	}
	entry.key = string(keyBytes)

	valueBytes := make([]byte, entry.valueSize)
	_, err = io.ReadFull(f, valueBytes)
	if err != nil {
		return nil, 0, fmt.Errorf("failed reading value (%d bytes) at pos %d: %w", entry.valueSize, position+headerSize+entry.keySize, err)
	}
	entry.value = string(valueBytes)

	// TODO: Verify CRC checksum here if needed for recovery robustness

	entrySize := headerSize + entry.keySize + entry.valueSize
	return &entry, entrySize, nil
}

// getKeyDir rebuilds the KeyDir map from existing log files. Called during init.
func getKeyDir(dataDir string) (map[string]KeyDir, int64, error) {
	keyDir := make(map[string]KeyDir)
	var maxFileId int64 = 0 // Track the latest file ID found

	files, err := os.ReadDir(dataDir)
	if err != nil {
		// If the directory doesn't exist yet, that's okay for init, return empty map
		if os.IsNotExist(err) {
			return keyDir, maxFileId, nil
		}
		return nil, maxFileId, fmt.Errorf("failed to read data directory %s: %w", dataDir, err)
	}

	for _, dirEntry := range files {
		// Skip directories and non-log files (like lock files)
		if dirEntry.IsDir() || filepath.Ext(dirEntry.Name()) != ".log" {
			continue
		}

		// Parse file ID from name (adjust parsing based on chosen naming scheme)
		fileName := dirEntry.Name()
		baseName := fileName[:len(fileName)-len(filepath.Ext(fileName))] // Remove ".log"
		fileId, err := strconv.ParseInt(baseName, 10, 64)
		if err != nil {
			// Log warning about potentially invalid file names
			fmt.Fprintf(os.Stderr, "Warning: Skipping file with invalid name format: %s (%v)\n", fileName, err)
			continue
		}

		// Keep track of the highest file ID seen
		if fileId > maxFileId {
			maxFileId = fileId
		}

		filePath := filepath.Join(dataDir, fileName)
		file, err := os.Open(filePath) // Open read-only for scanning
		if err != nil {
			// Log warning, skip file if unreadable
			fmt.Fprintf(os.Stderr, "Warning: Skipping unreadable file %s: %v\n", filePath, err)
			continue
		}

		var position int64 = 0
		for {
			entry, entrySize, err := readEntry(file, position)
			if err == io.EOF {
				break // End of this file
			}
			if err != nil {
				// Log warning about corrupted entry/file, stop processing this file
				fmt.Fprintf(os.Stderr, "Warning: Error reading entry from %s at pos %d, stopping scan for this file: %v\n", filePath, position, err)
				break
			}

			if entry.value == "<DELETED>" {
				delete(keyDir, entry.key)
			} else {
				// Calculate value position
				// Fixed header size: crc(4) + ts(8) + ksz(8) + vsz(8) = 28 bytes
				valuePos := position + 28 + entry.keySize

				// Only store if this entry is newer than existing one
				// Use TimeStamp for comparison
				existingEntry, exists := keyDir[entry.key]
				if !exists || entry.timeStamp > existingEntry.timeStamp {
					keyDir[entry.key] = KeyDir{
						fileId:        fileId,
						valueSize:     entry.valueSize,
						valuePosition: valuePos,
						timeStamp:     entry.timeStamp,
					}
				}
			}
			position += entrySize

		}
		file.Close() // Close after scanning each file
	}

	return keyDir, maxFileId, nil
}

// Lock file name
const lockFileName = "bitcask.lock"

type BitCaskStorageEngine struct {
	keyDir    map[string]KeyDir
	activeLog *Log         // Pointer to the current active log file
	dataDir   string       // Store dataDir path
	mu        sync.RWMutex // Mutex for goroutine safety (intra-process)
	fLock     *flock.Flock // File lock for single writer (inter-process)
}

func NewBitCaskStorageEngine(dataDir string) (*BitCaskStorageEngine, error) {
	// 1. Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory %s: %w", dataDir, err)
	}

	// 2. Acquire Inter-Process Lock (Single Writer)
	lockPath := filepath.Join(dataDir, lockFileName)
	fLock := flock.New(lockPath)
	// Try to lock exclusively, non-blocking
	locked, err := fLock.TryLock()
	if err != nil {
		// Error acquiring lock (e.g., permissions)
		return nil, fmt.Errorf("failed to check or acquire file lock %s: %w", lockPath, err)
	}
	if !locked {
		// Lock is already held by another process
		return nil, fmt.Errorf("data directory %s is locked by another process (lock file: %s)", dataDir, lockPath)
	}
	// If successful, fLock is held. It MUST be released on Close.

	// 3. Load KeyDir from existing files
	keyDir, lastFileId, err := getKeyDir(dataDir)
	if err != nil {
		fLock.Unlock() // Release lock if KeyDir load fails
		return nil, fmt.Errorf("failed to load key directory: %w", err)
	}

	// 4. Open the next log file for writing
	// If no files existed, start with ID 1. Otherwise, start with lastFileId + 1.
	nextFileId := lastFileId + 1
	if nextFileId == 1 && len(keyDir) == 0 { // Handle the very first run case explicitly
		// Could start with 0 or 1, let's use 1
	} else if lastFileId == 0 { // If directory existed but was empty or had invalid files
		nextFileId = 1
	}

	activeLog, err := openLogFile(dataDir, nextFileId)
	if err != nil {
		fLock.Unlock() // Release lock if opening log fails
		return nil, fmt.Errorf("failed to open active log file: %w", err)
	}

	// 5. Create the engine instance
	engine := &BitCaskStorageEngine{
		keyDir:    keyDir,
		activeLog: activeLog,
		dataDir:   dataDir,
		fLock:     fLock,
		// mu is implicitly initialized
	}

	return engine, nil
}

func (bcse *BitCaskStorageEngine) Set(key string, value string) error {
	// Acquire exclusive lock for writing (goroutine safety)
	bcse.mu.Lock()
	defer bcse.mu.Unlock()

	// TODO: Add logic here to check if bcse.activeLog.writerPosition exceeds a threshold.
	// If so:
	// 1. Call bcse.activeLog.Close()
	// 2. Determine the next file ID
	// 3. Call newLogFile = openLogFile(bcse.dataDir, nextFileId)
	// 4. Update bcse.activeLog = newLogFile
	// Handle errors at each step.

	// Prepare the entry
	dataDirFileLogEntry := newDataDirFileLogEntry(key, value)

	// Write to the active log file
	valuePosition, _, err := bcse.activeLog.setLogEntry(dataDirFileLogEntry)
	if err != nil {
		// This is a critical error, might indicate disk issues
		return fmt.Errorf("failed to write log entry for key '%s': %w", key, err)
	}

	// Update the in-memory KeyDir
	bcse.keyDir[key] = KeyDir{
		fileId:        bcse.activeLog.fileId,
		valueSize:     dataDirFileLogEntry.valueSize,
		valuePosition: valuePosition, // Store the start position of the value
		timeStamp:     dataDirFileLogEntry.timeStamp,
	}

	return nil
}

func (bcse *BitCaskStorageEngine) Get(key string) (string, error) {
	// Acquire shared lock for reading (goroutine safety)
	bcse.mu.RLock()
	defer bcse.mu.RUnlock()

	// Look up key in the in-memory index
	keyData, ok := bcse.keyDir[key]
	if !ok {
		return "", fmt.Errorf("key not found: %s", key) // Consider defining a specific ErrNotFound
	}

	// Read the value from the appropriate log file using the stored position and size
	value, err := getLogValue(bcse.dataDir, keyData.fileId, keyData.valuePosition, keyData.valueSize)
	if err != nil {
		// Error reading from disk
		return "", fmt.Errorf("failed to retrieve value for key '%s': %w", key, err)
	}

	if value == "<DELETED>" {
		return "", fmt.Errorf("key not found")
	}

	return value, nil
}

func (bcse *BitCaskStorageEngine) Delete(key string) error {
	// Acquire exclusive lock (goroutine safety) - as Delete modifies KeyDir and writes a tombstone
	bcse.mu.Lock()
	defer bcse.mu.Unlock()

	// 1. Check if key exists
	_, ok := bcse.keyDir[key]
	if !ok {
		return nil // Deleting a non-existent key is often treated as success (idempotent)
		// Alternatively: return fmt.Errorf("key not found: %s", key)
	}

	// 2. Write a "tombstone" entry to the log
	// A common practice is to write the key with a special marker value or an empty value.
	// Let's use an empty value string "" as the tombstone marker.
	tombstoneEntry := newDataDirFileLogEntry(key, "<DELETED>") // <DELETED> marks deletion

	_, _, err := bcse.activeLog.setLogEntry(tombstoneEntry)
	if err != nil {
		return fmt.Errorf("failed to write tombstone entry for key '%s': %w", key, err)
	}

	// 3. Remove the key from the in-memory KeyDir
	delete(bcse.keyDir, key)

	return nil
}

// Close releases resources (file lock, active log file). Crucial!
func (bcse *BitCaskStorageEngine) Close() error {
	// Acquire exclusive lock to prevent operations during close
	bcse.mu.Lock()
	defer bcse.mu.Unlock()

	var firstError error

	// Close the active log file
	if bcse.activeLog != nil {
		if err := bcse.activeLog.Close(); err != nil {
			firstError = fmt.Errorf("failed closing active log %s: %w", bcse.activeLog.filePath, err)
		}
		bcse.activeLog = nil // Mark as closed
	}

	// Release the inter-process file lock
	if bcse.fLock != nil {
		if err := bcse.fLock.Unlock(); err != nil {
			err = fmt.Errorf("failed releasing file lock %s: %w", bcse.fLock.Path(), err)
			// Chain errors if log closing also failed
			if firstError != nil {
				firstError = fmt.Errorf("%v; additionally: %w", firstError, err)
			} else {
				firstError = err
			}
		}
		bcse.fLock = nil // Mark as unlocked
	}

	// Optional: Clear the keyDir to free memory if the engine instance won't be reused
	// bcse.keyDir = nil

	return firstError // Return the first error encountered, if any
}
