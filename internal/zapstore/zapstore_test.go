package zapstore

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"zap-store/internal/storage/bitcask"
	"zap-store/internal/storage/inmem"
)

func TestZapStoreInMemSet(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		value      string
		wantErr    bool
		wantErrMsg string
		wantMapVal string
	}{
		{name: "valid", key: "key", value: "value", wantErr: false, wantErrMsg: "", wantMapVal: "value"},
		{name: "empty_key", key: "", value: "value2", wantErr: true, wantErrMsg: "key cannot be empty", wantMapVal: ""},
		{name: "empty_value", key: "key2", value: "", wantErr: false, wantErrMsg: "", wantMapVal: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var storageEngine = inmem.NewInMemStorageEngine()
			kvs := NewZapStore(storageEngine)
			err := kvs.Set(tt.key, tt.value)

			// Error occured when it shouldn't
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Set(%q, %q) error = %v, wantErr %v", tt.key, tt.value, err, tt.wantErr)
					return
				}
			}
			// No error occured when it should
			if err == nil {
				if tt.wantErr {
					t.Errorf("Set(%q, %q) error = %v, wantErr %v", tt.key, tt.value, err, tt.wantErr)
					return
				}
			}

			if tt.wantErr {
				if err.Error() != tt.wantErrMsg {
					t.Errorf("Set(%q, %q) error = %q, want %q", tt.key, tt.value, err.Error(), tt.wantErrMsg)
				}
				return
			}

			if got, _ := kvs.Get(tt.key); got != tt.wantMapVal {
				t.Errorf("Set(%q, %q) map[%q] = %q, want %q", tt.key, tt.value, tt.key, got, tt.wantMapVal)
			}
		})
	}
}

func TestZapStoreGet(t *testing.T) {
	var storageEngine = inmem.NewInMemStorageEngine()
	kvs := NewZapStore(storageEngine)
	kvs.Set("foo", "bar")

	tests := []struct {
		name       string
		key        string
		value      string
		wantErr    bool
		wantErrMsg string
		wantMapVal string
	}{
		{name: "valid", key: "foo", value: "bar", wantErr: false, wantErrMsg: "", wantMapVal: "bar"},
		{name: "non_existent_key", key: "baz", value: "", wantErr: true, wantErrMsg: "key not found", wantMapVal: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := kvs.Get(tt.key)

			// Error occured when it shouldn't
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Set(%q, %q) error = %v, wantErr %v", tt.key, tt.value, err, tt.wantErr)
					return
				}
			}
			// No error occured when it should
			if err == nil {
				if tt.wantErr {
					t.Errorf("Set(%q, %q) error = %v, wantErr %v", tt.key, tt.value, err, tt.wantErr)
					return
				}
			}

			if tt.wantErr {
				if err.Error() != tt.wantErrMsg {
					t.Errorf("Set(%q, %q) error = %q, want %q", tt.key, tt.value, err.Error(), tt.wantErrMsg)
				}
				return
			}

			if got != tt.wantMapVal {
				t.Errorf("Set(%q, %q) map[%q] = %q, want %q", tt.key, tt.value, tt.key, got, tt.wantMapVal)
				return
			}
		})

	}
}

func TestZapStoreDel(t *testing.T) {
	var storageEngine = inmem.NewInMemStorageEngine()
	kvs := NewZapStore(storageEngine)

	kvs.Set("foo", "bar")
	kvs.Delete("foo")

	if got, _ := kvs.Get("foo"); got != "" {
		t.Errorf("Delete() = %v, want %v", got, false)
	}
}

// preKeys generates a slice of pre-allocated keys to reduce allocations during benchmarks.
func preKeys(count int) []string {
	keys := make([]string, count)
	for i := 0; i < count; i++ {
		keys[i] = fmt.Sprintf("key%d", i)
	}
	return keys
}

func BenchmarkZapStoreInMemSet(b *testing.B) {
	var storageEngine = inmem.NewInMemStorageEngine()
	kvs := NewZapStore(storageEngine)

	// Pre-generate 1000 keys to avoid allocations during the loop
	keys := preKeys(1000)
	value := "value" // Fixed value to avoid allocations

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Cycle through pre-generated keys
		key := keys[i%1000]
		if err := kvs.Set(key, value); err != nil {
			b.Fatalf("Set failed: %v", err)
		}
	}
}

func BenchmarkZapStoreInMemGet(b *testing.B) {
	var storageEngine = inmem.NewInMemStorageEngine()
	kvs := NewZapStore(storageEngine)

	// Pre-populate with one key-value pair for consistent Get
	if err := kvs.Set("key", "value"); err != nil {
		b.Fatalf("Set failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := kvs.Get("key"); err != nil {
			b.Fatalf("Get failed: %v", err)
		}
	}
}

func BenchmarkZapStoreInMemDel(b *testing.B) {
	var storageEngine = inmem.NewInMemStorageEngine()
	kvs := NewZapStore(storageEngine)

	// Pre-populate with 1000 keys to ensure we can delete them
	keys := preKeys(1000)
	for _, key := range keys {
		if err := kvs.Set(key, "value"); err != nil {
			b.Fatalf("Set failed: %v", err)
		}
	}

	for i := 0; b.Loop(); i++ {
		// Cycle through keys to delete
		key := keys[i%1000]
		if err := kvs.Delete(key); err != nil {
			b.Fatalf("Delete failed: %v", err)
		}
		// Re-insert to avoid running out of keys
		if err := kvs.Set(key, "value"); err != nil {
			b.Fatalf("Set failed: %v", err)
		}
	}
}

func BenchmarkZapStoreInMemMixed(b *testing.B) {
	var storageEngine = inmem.NewInMemStorageEngine()
	kvs := NewZapStore(storageEngine)

	// Pre-populate with 10,000 keys to simulate a realistic dataset
	keys := preKeys(10000)
	for _, key := range keys {
		if err := kvs.Set(key, "value"); err != nil {
			b.Fatalf("Set failed: %v", err)
		}
	}

	// Mixed workload: 50% Get, 40% Set, 10% Delete

	for i := 0; b.Loop(); i++ {
		r := rand.Float64() // Random number between 0 and 1
		key := keys[i%10000]
		switch {
		case r < 0.5: // 50% Get
			if _, err := kvs.Get(key); err != nil {
				b.Fatalf("Get failed: %v", err)
			}
		case r < 0.9: // 40% Set (0.5 to 0.9)
			if err := kvs.Set(key, "value"); err != nil {
				b.Fatalf("Set failed: %v", err)
			}
		default: // 10% Delete (0.9 to 1.0)
			if err := kvs.Delete(key); err != nil {
				b.Fatalf("Delete failed: %v", err)
			}
			// Re-insert to avoid running out of keys
			if err := kvs.Set(key, "value"); err != nil {
				b.Fatalf("Set failed: %v", err)
			}
		}
	}
}

func BenchmarkZapStoreInMemConcurrent(b *testing.B) {
	var storageEngine = inmem.NewInMemStorageEngine()
	kvs := NewZapStore(storageEngine)

	// Pre-populate with 10,000 keys
	keys := preKeys(10000)
	for _, key := range keys {
		if err := kvs.Set(key, "value"); err != nil {
			b.Fatalf("Set failed: %v", err)
		}
	}

	// Run in parallel to simulate concurrent access
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r := rand.Float64()
			key := keys[rand.Intn(10000)]
			switch {
			case r < 0.5: // 50% Get
				if _, err := kvs.Get(key); err != nil && !strings.Contains(err.Error(), "key not found") {
					b.Fatalf("Get failed: %v", err)
				}
			case r < 0.9: // 40% Set
				if err := kvs.Set(key, "value"); err != nil {
					b.Fatalf("Set failed: %v", err)
				}
			default: // 10% Delete
				if err := kvs.Delete(key); err != nil {
					b.Fatalf("Delete failed: %v", err)
				}
				// Re-insert to avoid running out of keys
				if err := kvs.Set(key, "value"); err != nil {
					b.Fatalf("Set failed: %v", err)
				}
			}
		}
	})
}

func BenchmarkZapStoreBitCaskSet(b *testing.B) {
	tempDir := b.TempDir()
	storageEngine, err := bitcask.NewBitCaskStorageEngine(tempDir)

	if err != nil {
		b.Fatalf("Failed to initialize BitCaskStorageEngine: %v", err)
	}

	kvs := NewZapStore(storageEngine)

	// Pre-generate 1000 keys to avoid allocations during the loop
	keys := preKeys(1000)
	value := "value" // Fixed value to avoid allocations

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Cycle through pre-generated keys
		key := keys[i%1000]
		if err := kvs.Set(key, value); err != nil {
			b.Fatalf("Set failed: %v", err)
		}
	}

	// Clean up
	if err := storageEngine.Close(); err != nil {
		b.Fatalf("Failed to close BitCaskStorageEngine: %v", err)
	}
}

func BenchmarkZapStoreBitCaskGet(b *testing.B) {
	tempDir := b.TempDir()
	storageEngine, err := bitcask.NewBitCaskStorageEngine(tempDir)
	if err != nil {
		b.Fatalf("Failed to initialize BitCaskStorageEngine: %v", err)
	}
	kvs := NewZapStore(storageEngine)

	// Pre-populate with one key-value pair for consistent Get
	if err := kvs.Set("key", "value"); err != nil {
		b.Fatalf("Set failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := kvs.Get("key"); err != nil {
			b.Fatalf("Get failed: %v", err)
		}
	}

	// Clean up
	if err := storageEngine.Close(); err != nil {
		b.Fatalf("Failed to close BitCaskStorageEngine: %v", err)
	}
}

func BenchmarkZapStoreBitCaskDel(b *testing.B) {
	tempDir := b.TempDir()
	storageEngine, err := bitcask.NewBitCaskStorageEngine(tempDir)
	if err != nil {
		b.Fatalf("Failed to initialize BitCaskStorageEngine: %v", err)
	}

	kvs := NewZapStore(storageEngine)

	// Pre-populate with 1000 keys to ensure we can delete them
	keys := preKeys(1000)
	for _, key := range keys {
		if err := kvs.Set(key, "value"); err != nil {
			b.Fatalf("Set failed: %v", err)
		}
	}

	for i := 0; b.Loop(); i++ {
		// Cycle through keys to delete
		key := keys[i%1000]
		if err := kvs.Delete(key); err != nil {
			b.Fatalf("Delete failed: %v", err)
		}
		// Re-insert to avoid running out of keys
		if err := kvs.Set(key, "value"); err != nil {
			b.Fatalf("Set failed: %v", err)
		}
	}

	// Clean up
	if err := storageEngine.Close(); err != nil {
		b.Fatalf("Failed to close BitCaskStorageEngine: %v", err)
	}
}

func BenchmarkZapStoreBitCaskMixed(b *testing.B) {
	tempDir := b.TempDir()
	storageEngine, err := bitcask.NewBitCaskStorageEngine(tempDir)
	if err != nil {
		b.Fatalf("Failed to initialize BitCaskStorageEngine: %v", err)
	}

	kvs := NewZapStore(storageEngine)

	// Pre-populate with 10,000 keys to simulate a realistic dataset
	keys := preKeys(10000)
	for _, key := range keys {
		if err := kvs.Set(key, "value"); err != nil {
			b.Fatalf("Set failed: %v", err)
		}
	}

	// Mixed workload: 50% Get, 40% Set, 10% Delete

	for i := 0; b.Loop(); i++ {
		r := rand.Float64() // Random number between 0 and 1
		key := keys[i%10000]
		switch {
		case r < 0.5: // 50% Get
			if _, err := kvs.Get(key); err != nil {
				b.Fatalf("Get failed: %v", err)
			}
		case r < 0.9: // 40% Set (0.5 to 0.9)
			if err := kvs.Set(key, "value"); err != nil {
				b.Fatalf("Set failed: %v", err)
			}
		default: // 10% Delete (0.9 to 1.0)
			if err := kvs.Delete(key); err != nil {
				b.Fatalf("Delete failed: %v", err)
			}
			// Re-insert to avoid running out of keys
			if err := kvs.Set(key, "value"); err != nil {
				b.Fatalf("Set failed: %v", err)
			}
		}
	}
}

func BenchmarkZapStoreBitCaskConcurrent(b *testing.B) {
	tempDir := b.TempDir()
	storageEngine, err := bitcask.NewBitCaskStorageEngine(tempDir)
	if err != nil {
		b.Fatalf("Failed to initialize BitCaskStorageEngine: %v", err)
	}

	kvs := NewZapStore(storageEngine)

	// Pre-populate with 10,000 keys
	keys := preKeys(10000)
	for _, key := range keys {
		if err := kvs.Set(key, "value"); err != nil {
			b.Fatalf("Set failed: %v", err)
		}
	}

	// Run in parallel to simulate concurrent access
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r := rand.Float64()
			key := keys[rand.Intn(10000)]
			switch {
			case r < 0.5: // 50% Get
				if _, err := kvs.Get(key); err != nil && !strings.Contains(err.Error(), "key not found") {
					b.Fatalf("Get failed: %v", err)
				}
			case r < 0.9: // 40% Set
				if err := kvs.Set(key, "value"); err != nil {
					b.Fatalf("Set failed: %v", err)
				}
			default: // 10% Delete
				if err := kvs.Delete(key); err != nil {
					b.Fatalf("Delete failed: %v", err)
				}
				// Re-insert to avoid running out of keys
				if err := kvs.Set(key, "value"); err != nil {
					b.Fatalf("Set failed: %v", err)
				}
			}
		}
	})
}
