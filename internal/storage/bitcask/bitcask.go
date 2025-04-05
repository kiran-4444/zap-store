package bitcask

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"
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
	timeStamp := time.Now().Unix()
	keySize := len(key)
	valueSize := len(value)
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
	var buf bytes.Buffer

	if err := binary.Write(&buf, binary.BigEndian, ddfle.crc); err != nil {
		return nil, err
	}
	if err := binary.Write(&buf, binary.BigEndian, ddfle.timeStamp); err != nil {
		return nil, err
	}
	if err := binary.Write(&buf, binary.BigEndian, ddfle.keySize); err != nil {
		return nil, err
	}
	if err := binary.Write(&buf, binary.BigEndian, ddfle.valueSize); err != nil {
		return nil, err
	}
	if _, err := buf.Write([]byte(ddfle.key)); err != nil {
		return nil, err
	}
	if _, err := buf.Write([]byte(ddfle.value)); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type Log struct {
	writerPosition int64
	file           *os.File
	fileId         int64
	filePath       string
	dataDir        string
}

func newLog(dataDir string) (*Log, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory %s: %w", dataDir, err)
	}
	fileId := time.Now().Unix()
	filePath := path.Join(dataDir, fmt.Sprint(fileId))
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, 0755)

	if err != nil {
		return nil, fmt.Errorf("cannot open l entry %s: %w", filePath, err)
	}

	return &Log{
		writerPosition: 0,
		file:           file,
		fileId:         fileId,
		filePath:       filePath,
		dataDir:        dataDir,
	}, nil
}

func (l *Log) setLogEntry(dataDirFileLogEntry *DataDirFileLogEntry) (int64, error) {
	bytesToWrite, err := dataDirFileLogEntry.toBytes()
	if err != nil {
		return -1, err
	}

	l.file.Seek(l.writerPosition, os.SEEK_SET)
	l.file.Write(bytesToWrite)
	l.writerPosition += int64(len(bytesToWrite))

	valuePosition := l.writerPosition - dataDirFileLogEntry.valueSize

	return valuePosition, nil
}

func (l *Log) getLogEntry(fileId int64, offset int64, size int64) (string, error) {
	filePath := path.Join(l.dataDir, strconv.Itoa(int(fileId)))
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0755)

	if err != nil {
		return "", err
	}
	currentOffset, err := file.Seek(offset, os.SEEK_SET)
	if err != nil {
		return "", err
	}

	value := make([]byte, size)

	readLen, err := file.Read(value)
	if err != nil {
		return "", err
	}

	if int64(readLen) != size {
		return "", fmt.Errorf("size mismatch")
	}

	_, err = file.Seek(currentOffset, os.SEEK_SET)
	if err != nil {
		return "", err
	}

	return string(value), nil
}

type KeyDir struct {
	fileId        int64
	valueSize     int64
	valuePosition int64
	timeStamp     int64
}

// readEntry reads a single DataDirFileLogEntry from the file at the given position
func readEntry(f *os.File, position int64) (*DataDirFileLogEntry, error) {
	// Seek to the position
	_, err := f.Seek(position, os.SEEK_SET)
	if err != nil {
		return nil, err
	}

	// Read fixed-size fields (20 bytes: crc=4, timeStamp=8, keySize=8, valueSize=8)
	header := make([]byte, 28)
	_, err = io.ReadFull(f, header)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewReader(header)
	var entry DataDirFileLogEntry

	if err := binary.Read(buf, binary.BigEndian, &entry.crc); err != nil {
		return nil, err
	}
	if err := binary.Read(buf, binary.BigEndian, &entry.timeStamp); err != nil {
		return nil, err
	}
	if err := binary.Read(buf, binary.BigEndian, &entry.keySize); err != nil {
		return nil, err
	}
	if err := binary.Read(buf, binary.BigEndian, &entry.valueSize); err != nil {
		return nil, err
	}

	// Read key
	keyBytes := make([]byte, entry.keySize)
	_, err = io.ReadFull(f, keyBytes)
	if err != nil {
		return nil, err
	}
	entry.key = string(keyBytes)

	// Read value
	valueBytes := make([]byte, entry.valueSize)
	_, err = io.ReadFull(f, valueBytes)
	if err != nil {
		return nil, err
	}
	entry.value = string(valueBytes)

	return &entry, nil
}

func getKeyDir(dataDir string) (map[string]KeyDir, error) {
	keyDir := make(map[string]KeyDir)

	files, err := os.ReadDir(dataDir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		fileName := file.Name()
		fileId, err := strconv.ParseInt(fileName, 10, 64)
		if err != nil {
			return nil, err
		}

		filePath := filepath.Join(dataDir, fileName)
		file, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		var position int64
		for {
			entry, err := readEntry(file, position)
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}

			keyDir[entry.key] = KeyDir{
				fileId:        fileId,
				valueSize:     int64(entry.valueSize),
				valuePosition: position + int64(28) + 1, // Offset after fixed fields
				timeStamp:     entry.timeStamp,
			}

			position += int64(28) + int64(entry.keySize) + int64(entry.valueSize)
		}

	}

	return keyDir, nil
}

type BitCaskStorageEngine struct {
	keyDir map[string]KeyDir
	l      *Log
}

func NewBitCaskStorageEngine(dataDir string) (*BitCaskStorageEngine, error) {
	logEntry, err := newLog(dataDir)
	if err != nil {
		return nil, err
	}

	keyDir, err := getKeyDir(dataDir)

	if err != nil {
		return nil, err
	}
	return &BitCaskStorageEngine{
		keyDir: keyDir,
		l:      logEntry,
	}, nil
}

func (bcse *BitCaskStorageEngine) Set(key string, value string) error {
	dataDirFileLogEntry := newDataDirFileLogEntry(key, value)

	valuePosition, err := bcse.l.setLogEntry(dataDirFileLogEntry)
	if err != nil {
		return err
	}

	bcse.keyDir[key] = KeyDir{
		fileId:        bcse.l.fileId,
		valueSize:     dataDirFileLogEntry.valueSize,
		valuePosition: valuePosition,
		timeStamp:     dataDirFileLogEntry.timeStamp,
	}

	fmt.Println(bcse.keyDir)

	return nil
}

func (bcse *BitCaskStorageEngine) Get(key string) (string, error) {
	if _, ok := bcse.keyDir[key]; !ok {
		return "", fmt.Errorf("key not found")
	}

	value, err := bcse.l.getLogEntry(bcse.keyDir[key].fileId, bcse.keyDir[key].valuePosition, bcse.keyDir[key].valueSize)
	if err != nil {
		return "", err
	}

	return value, nil

}

func (bcse *BitCaskStorageEngine) Del(key string) error {
	return nil
}
