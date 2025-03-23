package kvstore

import (
	"fmt"
	"kv-store/internal/storage/inmem"
	"math/rand"
	"strings"
	"testing"
)

func TestKVStoreInMemSet(t *testing.T) {
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
			kvs := NewKVStore(storageEngine)
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

func TestKVStoreGet(t *testing.T) {
	var storageEngine = inmem.NewInMemStorageEngine()
	kvs := NewKVStore(storageEngine)
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

func TestKVStoreDel(t *testing.T) {
	var storageEngine = inmem.NewInMemStorageEngine()
	kvs := NewKVStore(storageEngine)

	kvs.Set("foo", "bar")
	kvs.Del("foo")

	if got, _ := kvs.Get("foo"); got != "" {
		t.Errorf("Del() = %v, want %v", got, false)
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

func BenchmarkKVStoreInMemSet(b *testing.B) {
	var storageEngine = inmem.NewInMemStorageEngine()
	kvs := NewKVStore(storageEngine)

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

func BenchmarkKVStoreInMemGet(b *testing.B) {
	var storageEngine = inmem.NewInMemStorageEngine()
	kvs := NewKVStore(storageEngine)

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

func BenchmarkKVStoreInMemDel(b *testing.B) {
	var storageEngine = inmem.NewInMemStorageEngine()
	kvs := NewKVStore(storageEngine)

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
		if err := kvs.Del(key); err != nil {
			b.Fatalf("Del failed: %v", err)
		}
		// Re-insert to avoid running out of keys
		if err := kvs.Set(key, "value"); err != nil {
			b.Fatalf("Set failed: %v", err)
		}
	}
}

func BenchmarkKVStoreInMemMixed(b *testing.B) {
	var storageEngine = inmem.NewInMemStorageEngine()
	kvs := NewKVStore(storageEngine)

	// Pre-populate with 10,000 keys to simulate a realistic dataset
	keys := preKeys(10000)
	for _, key := range keys {
		if err := kvs.Set(key, "value"); err != nil {
			b.Fatalf("Set failed: %v", err)
		}
	}

	// Mixed workload: 50% Get, 40% Set, 10% Del

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
		default: // 10% Del (0.9 to 1.0)
			if err := kvs.Del(key); err != nil {
				b.Fatalf("Del failed: %v", err)
			}
			// Re-insert to avoid running out of keys
			if err := kvs.Set(key, "value"); err != nil {
				b.Fatalf("Set failed: %v", err)
			}
		}
	}
}

func BenchmarkKVStoreInMemConcurrent(b *testing.B) {
	var storageEngine = inmem.NewInMemStorageEngine()
	kvs := NewKVStore(storageEngine)

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
			default: // 10% Del
				if err := kvs.Del(key); err != nil {
					b.Fatalf("Del failed: %v", err)
				}
				// Re-insert to avoid running out of keys
				if err := kvs.Set(key, "value"); err != nil {
					b.Fatalf("Set failed: %v", err)
				}
			}
		}
	})
}
