package bitcask

import (
	// Import errors package for Is/As if needed
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

// Helper function to create and close an engine instance for simple tests
// Returns the engine, the temp directory path, and a cleanup function
func setupTestEngine(t *testing.T) (*BitCaskStorageEngine, string) {
	t.Helper() // Mark this as a test helper function
	tempDir := t.TempDir()
	db, err := NewBitCaskStorageEngine(tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize test engine in %s: %v", tempDir, err)
	}

	// Ensure Close is called when the test using this helper finishes
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing test engine in %s: %v", tempDir, err)
		}
	})

	return db, tempDir
}

func TestBitCaskStorageEngine_SetGet(t *testing.T) {
	// Test cases for basic Set and Get operations
	tests := []struct {
		name         string
		key          string
		value        string
		expectGet    string // Expected value on subsequent Get
		setShouldErr bool   // Whether Set itself should error (e.g., invalid key?) - not applicable here
		getShouldErr bool   // Whether Get should error (e.g., key not found)
		errMsg       string // Expected error message substring if getShouldErr is true
	}{
		{name: "simple set/get", key: "key1", value: "value1", expectGet: "value1", getShouldErr: false},
		{name: "set/get empty value", key: "key2", value: "", expectGet: "", getShouldErr: false},
		{name: "set/get unicode", key: "你好", value: "世界", expectGet: "世界", getShouldErr: false},
		{name: "get non-existent", key: "non_existent_key", value: "", expectGet: "", getShouldErr: true, errMsg: "key not found"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, _ := setupTestEngine(t) // Use helper to setup/teardown

			// Perform Set if a value is provided for setting
			if tt.value != "" || tt.key == "key2" { // Handle setting empty value case
				err := db.Set(tt.key, tt.value)
				if (err != nil) != tt.setShouldErr {
					t.Fatalf("Set(%q, %q) error = %v, wantSetErr %v", tt.key, tt.value, err, tt.setShouldErr)
				}
				if err != nil {
					return // Don't proceed to Get if Set failed as expected
				}
			} else if tt.key == "non_existent_key" {
				// Don't set anything for the non-existent key test case
			}

			// Perform Get
			gotValue, err := db.Get(tt.key)

			// Check for expected errors on Get
			if err != nil {
				if !tt.getShouldErr {
					t.Errorf("Get(%q) unexpected error = %v", tt.key, err)
				} else if !strings.Contains(err.Error(), tt.errMsg) { // Check if error message contains expected text
					t.Errorf("Get(%q) error = %q, want error containing %q", tt.key, err.Error(), tt.errMsg)
				}
			} else { // err == nil
				if tt.getShouldErr {
					t.Errorf("Get(%q) expected error containing %q, but got nil", tt.key, tt.errMsg)
				}
			}

			// Check the retrieved value only if no error was expected
			if !tt.getShouldErr && gotValue != tt.expectGet {
				t.Errorf("Get(%q) got value %q, want %q", tt.key, gotValue, tt.expectGet)
			}
		})
	}
}

func TestBitCaskStorageEngine_Overwrite(t *testing.T) {
	t.Run("overwrite existing key", func(t *testing.T) {
		db, _ := setupTestEngine(t)

		key := "overwrite_me"
		initialValue := "value_v1"
		newValue := "value_v2"

		// Set initial value
		if err := db.Set(key, initialValue); err != nil {
			t.Fatalf("Initial Set failed: %v", err)
		}
		got1, err1 := db.Get(key)
		if err1 != nil || got1 != initialValue {
			t.Fatalf("Get after initial Set failed: err=%v, val=%q, want=%q", err1, got1, initialValue)
		}

		// Set new value (overwrite)
		if err := db.Set(key, newValue); err != nil {
			t.Fatalf("Overwrite Set failed: %v", err)
		}

		// Get again and check new value
		got2, err2 := db.Get(key)
		if err2 != nil {
			t.Errorf("Get after overwrite Set returned error: %v", err2)
		}
		if got2 != newValue {
			t.Errorf("Get after overwrite: got %q, want %q", got2, newValue)
		}
	})
}

func TestBitCaskStorageEngine_Delete(t *testing.T) {
	tests := []struct {
		name         string
		keyToSet     string // Key to set initially (if any)
		valueToSet   string
		keyToDelete  string
		delShouldErr bool   // Currently Delete doesn't return errors for non-existent keys
		getShouldErr bool   // Whether Get after Delete should error
		errMsg       string // Expected error message for Get after Delete
	}{
		{name: "delete existing key", keyToSet: "del_key1", valueToSet: "del_val1", keyToDelete: "del_key1", delShouldErr: false, getShouldErr: true, errMsg: "key not found"},
		{name: "delete non-existent key", keyToSet: "", valueToSet: "", keyToDelete: "del_key_never_set", delShouldErr: false, getShouldErr: true, errMsg: "key not found"},
		{name: "delete key then check another", keyToSet: "another_key", valueToSet: "another_val", keyToDelete: "del_key1", delShouldErr: false, getShouldErr: true, errMsg: "key not found"}, // Check getting the deleted key
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, _ := setupTestEngine(t)

			// Set initial key if specified
			if tt.keyToSet != "" {
				if err := db.Set(tt.keyToSet, tt.valueToSet); err != nil {
					t.Fatalf("Initial Set(%q) failed: %v", tt.keyToSet, err)
				}
				// Optional: Verify set worked
				// _, err := db.Get(tt.keyToSet)
				// if err != nil { t.Fatalf("Get after initial Set failed: %v", err) }
			}

			// Perform Delete
			err := db.Delete(tt.keyToDelete)
			if (err != nil) != tt.delShouldErr {
				t.Fatalf("Delete(%q) error = %v, wantDelErr %v", tt.keyToDelete, err, tt.delShouldErr)
			}
			if err != nil {
				return // Don't proceed if Delete failed unexpectedly
			}

			// Perform Get on the deleted key
			_, errGet := db.Get(tt.keyToDelete)

			if errGet != nil {
				if !tt.getShouldErr {
					t.Errorf("Get(%q) after Delete unexpected error = %v", tt.keyToDelete, errGet)
				} else if !strings.Contains(errGet.Error(), tt.errMsg) {
					t.Errorf("Get(%q) after Delete error = %q, want error containing %q", tt.keyToDelete, errGet.Error(), tt.errMsg)
				}
			} else { // errGet == nil
				if tt.getShouldErr {
					t.Errorf("Get(%q) after Delete expected error containing %q, but got nil", tt.keyToDelete, tt.errMsg)
				}
			}

			// If we deleted a different key, check that the other key still exists
			if tt.name == "delete key then check another" {
				val, errOther := db.Get(tt.keyToSet)
				if errOther != nil {
					t.Errorf("Get(%q) after deleting another key returned error: %v", tt.keyToSet, errOther)
				}
				if val != tt.valueToSet {
					t.Errorf("Get(%q) after deleting another key: got %q, want %q", tt.keyToSet, val, tt.valueToSet)
				}
			}
		})
	}
}

// TestPersistence requires a directory that survives between sub-tests.
// We use MkdirTemp and t.Cleanup for this.
func TestBitCaskStorageEngine_Persistence(t *testing.T) {
	// Create a temp dir manually that persists across t.Run calls
	persistDir, err := os.MkdirTemp("", "bitcask-persist-test-*")
	if err != nil {
		t.Fatalf("Failed to create persistent temp dir: %v", err)
	}
	// Use t.Cleanup to remove the directory *after* all sub-tests in this function complete
	t.Cleanup(func() {
		os.RemoveAll(persistDir)
		t.Logf("Cleaned up persistent test directory: %s", persistDir)
	})
	t.Logf("Using persistent test directory: %s", persistDir)

	key := "persist_key"
	value := "persist_value"

	// --- Phase 1: Set data and close ---
	t.Run("SetAndClose", func(t *testing.T) {
		db1, err := NewBitCaskStorageEngine(persistDir)
		if err != nil {
			t.Fatalf("Failed to initialize engine (Phase 1): %v", err)
		}
		defer func() { // Defer close within the subtest
			if err := db1.Close(); err != nil {
				t.Errorf("Error closing engine (Phase 1): %v", err)
			}
		}()

		if err := db1.Set(key, value); err != nil {
			t.Fatalf("Set failed (Phase 1): %v", err)
		}
		// Optional: immediate get check
		// got, _ := db1.Get(key)
		// if got != value { t.Fatalf("Immediate Get failed (Phase 1)") }
	})

	// --- Phase 2: Re-open and Get data ---
	t.Run("ReopenAndGet", func(t *testing.T) {
		// Allow some time for file system operations/closing if needed (usually not necessary)
		// time.Sleep(10 * time.Millisecond)

		db2, err := NewBitCaskStorageEngine(persistDir) // Use the SAME directory
		if err != nil {
			t.Fatalf("Failed to initialize engine (Phase 2): %v", err)
		}
		defer func() { // Defer close within the subtest
			if err := db2.Close(); err != nil {
				t.Errorf("Error closing engine (Phase 2): %v", err)
			}
		}()

		gotValue, err := db2.Get(key)
		if err != nil {
			t.Fatalf("Get failed (Phase 2): %v", err)
		}
		if gotValue != value {
			t.Errorf("Get (Phase 2): got %q, want %q", gotValue, value)
		}
	})
}

// Test Delete Persistence (similar setup to TestPersistence)
func TestBitCaskStorageEngine_DeletePersistence(t *testing.T) {
	persistDir, err := os.MkdirTemp("", "bitcask-del-persist-test-*")
	if err != nil {
		t.Fatalf("Failed to create persistent temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(persistDir) })
	t.Logf("Using delete persistent test directory: %s", persistDir)

	key1, val1 := "del_persist1", "value1"
	key2, val2 := "del_persist2", "value2"

	// Phase 1: Set two keys, delete one, close
	t.Run("SetDeleteClose", func(t *testing.T) {
		db1, err := NewBitCaskStorageEngine(persistDir)
		if err != nil {
			t.Fatalf("Phase 1 init failed: %v", err)
		}
		defer func() { db1.Close() }() // Simplified defer

		if err := db1.Set(key1, val1); err != nil {
			t.Fatalf("Phase 1 Set key1 failed: %v", err)
		}
		if err := db1.Set(key2, val2); err != nil {
			t.Fatalf("Phase 1 Set key2 failed: %v", err)
		}
		if err := db1.Delete(key1); err != nil {
			t.Fatalf("Phase 1 Delete key1 failed: %v", err)
		}
	})

	// Phase 2: Re-open, check deleted key is gone, other key remains
	t.Run("ReopenCheck", func(t *testing.T) {
		db2, err := NewBitCaskStorageEngine(persistDir)
		if err != nil {
			t.Fatalf("Phase 2 init failed: %v", err)
		}
		defer func() { db2.Close() }() // Simplified defer

		// Check deleted key
		_, errGet1 := db2.Get(key1)
		if errGet1 == nil {
			t.Errorf("Get(%q) after delete/reopen succeeded, expected 'not found' error", key1)
		} else if !strings.Contains(errGet1.Error(), "key not found") {
			t.Errorf("Get(%q) after delete/reopen wrong error: %v", key1, errGet1)
		}

		// Check remaining key
		gotVal2, errGet2 := db2.Get(key2)
		if errGet2 != nil {
			t.Errorf("Get(%q) after delete/reopen failed: %v", key2, errGet2)
		}
		if gotVal2 != val2 {
			t.Errorf("Get(%q) after delete/reopen got %q, want %q", key2, gotVal2, val2)
		}
	})
}

// TestFileLocking ensures only one process can write. Needs careful setup.
func TestBitCaskStorageEngine_FileLocking(t *testing.T) {
	// Create a shared directory for this test
	lockDir, err := os.MkdirTemp("", "bitcask-lock-test-*")
	if err != nil {
		t.Fatalf("Failed to create locking test dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(lockDir) })

	// --- Instance 1: Acquire lock ---
	db1, err1 := NewBitCaskStorageEngine(lockDir)
	if err1 != nil {
		t.Fatalf("Failed to initialize first engine instance: %v", err1)
	}
	// Defer Close for db1 *after* the second check
	defer func() {
		if err := db1.Close(); err != nil {
			t.Errorf("Error closing first engine instance: %v", err)
		}
	}()

	// --- Instance 2: Try to acquire lock on the SAME directory ---
	db2, err2 := NewBitCaskStorageEngine(lockDir)

	// --- Assertions ---
	if err2 == nil {
		// If err2 is nil, it means the second instance acquired the lock, which is wrong.
		// We need to close the second instance if it was wrongly created.
		if db2 != nil {
			db2.Close() // Attempt to clean up
		}
		t.Fatalf("Second NewBitCaskStorageEngine succeeded, expected file lock error")
	}

	// Check if the error message indicates a locking issue
	expectedErrorSubstring := "locked by another process"
	if !strings.Contains(err2.Error(), expectedErrorSubstring) {
		t.Errorf("Second NewBitCaskStorageEngine error = %q, want error containing %q", err2.Error(), expectedErrorSubstring)
	}

	if db2 != nil {
		t.Errorf("Second engine instance db2 should be nil on error, but wasn't")
	}
}

// Test basic concurrency safety using the RWMutex
func TestBitCaskStorageEngine_Concurrency(t *testing.T) {
	db, _ := setupTestEngine(t) // Setup engine with automatic cleanup

	numGoroutines := 50
	numOpsPerGoroutine := 100
	var wg sync.WaitGroup

	// Concurrent Sets
	wg.Add(numGoroutines)
	for i := range numGoroutines {
		go func(gID int) {
			defer wg.Done()
			for j := range numOpsPerGoroutine {
				key := fmt.Sprintf("conc_key_%d_%d", gID, j)
				val := fmt.Sprintf("conc_val_%d_%d", gID, j)
				err := db.Set(key, val)
				if err != nil {
					t.Errorf("Concurrent Set failed for key %s: %v", key, err)
					// Use t.Error not t.Fatal in goroutines
				}
			}
		}(i)
	}

	// Concurrent Gets (can run alongside Sets due to RWMutex)
	// Start gets slightly after sets to increase chance of overlap
	time.Sleep(5 * time.Millisecond)
	wg.Add(numGoroutines)
	for i := range numGoroutines {
		go func(gID int) {
			defer wg.Done()
			for j := range numOpsPerGoroutine / 2 { // Do fewer Gets to avoid excessive file opening
				// Try getting keys potentially set by other goroutines
				getKey := fmt.Sprintf("conc_key_%d_%d", gID/2, j) // Example key pattern
				_, err := db.Get(getKey)
				// Getting might error if the key wasn't set *yet*,
				// or if it was deleted concurrently (if deletes were added).
				// We mainly care that the Gets don't cause crashes or corrupt state.
				if err != nil && !strings.Contains(err.Error(), "key not found") {
					// Ignore "key not found" as it's expected race condition
					t.Errorf("Concurrent Get failed unexpectedly for key %s: %v", getKey, err)
				}
			}
		}(i)
	}

	wg.Wait() // Wait for all Sets and Gets to complete

	// Verification: Check if some keys set earlier are still present and correct
	// Check a subset to avoid excessive test time
	for j := 0; j < numOpsPerGoroutine; j += 10 {
		gID := numGoroutines / 2 // Check keys from a middle goroutine
		key := fmt.Sprintf("conc_key_%d_%d", gID, j)
		expectedVal := fmt.Sprintf("conc_val_%d_%d", gID, j)
		gotVal, err := db.Get(key)
		if err != nil {
			t.Errorf("Verification Get failed for key %s: %v", key, err)
		} else if gotVal != expectedVal {
			t.Errorf("Verification Get for key %s: got %q, want %q", key, gotVal, expectedVal)
		}
	}
}
