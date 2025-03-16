package kvstore

import (
	"testing"
)

func TestKVStoreSet(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		value      string
		wantErr    bool
		wantErrMsg string
		wantMapVal string
	}{
		{name: "valid", key: "key", value: "value", wantErr: false, wantErrMsg: "", wantMapVal: "value"},
		{name: "empty_key", key: "", value: "value2", wantErr: true, wantErrMsg: "Key cannot be empty", wantMapVal: ""},
		{name: "empty_value", key: "key2", value: "", wantErr: false, wantErrMsg: "", wantMapVal: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kvs := NewKVStore()
			err := kvs.Set(tt.key, tt.value)

			// Error occured when it shouldn't
			if err != nil {
				if tt.wantErr == false {
					t.Errorf("Set(%q, %q) error = %v, wantErr %v", tt.key, tt.value, err, tt.wantErr)
					return
				}
			}
			// No error occured when it should
			if err == nil {
				if tt.wantErr == true {
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

			if got := kvs.hashMap[tt.key]; got != tt.wantMapVal {
				t.Errorf("Set(%q, %q) map[%q] = %q, want %q", tt.key, tt.value, tt.key, got, tt.wantMapVal)
			}
		})
	}
}

func TestKVStoreGet(t *testing.T) {
	kvs := NewKVStore()
	kvs.hashMap["foo"] = "bar"

	tests := []struct {
		name       string
		key        string
		value      string
		wantErr    bool
		wantErrMsg string
		wantMapVal string
	}{
		{name: "valid", key: "foo", value: "bar", wantErr: false, wantErrMsg: "", wantMapVal: "bar"},
		{name: "non_existent_key", key: "baz", value: "", wantErr: true, wantErrMsg: "Key not found", wantMapVal: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := kvs.Get(tt.key)

			// Error occured when it shouldn't
			if err != nil {
				if tt.wantErr == false {
					t.Errorf("Set(%q, %q) error = %v, wantErr %v", tt.key, tt.value, err, tt.wantErr)
					return
				}
			}
			// No error occured when it should
			if err == nil {
				if tt.wantErr == true {
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
			}
		})

	}
}

func TestKVStoreDel(t *testing.T) {
	kvs := NewKVStore()

	kvs.hashMap["foo"] = "bar"
	kvs.Del("foo")

	if _, ok := kvs.hashMap["foo"]; ok {
		t.Errorf("Del() = %v, want %v", ok, false)
	}
}
