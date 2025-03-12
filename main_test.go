package main

import "testing"

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
		{name: "empty_key", key: "", value: "value2", wantErr: true, wantErrMsg: "KEY CANNOT BE EMPTY", wantMapVal: ""},
		{name: "empty_value", key: "key2", value: "", wantErr: true, wantErrMsg: "VALUE CANNOT BE EMPTY", wantMapVal: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kvs := NewKVStore()
			err := kvs.Set(tt.key, tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("Set(%q, %q) error = %v, wantErr %v", tt.key, tt.value, err, tt.wantErr)
				return
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
		{name: "non_existent_key", key: "baz", value: "", wantErr: true, wantErrMsg: "KEY NOT FOUND", wantMapVal: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := kvs.Get(tt.key)

			if (err != nil) != tt.wantErr {
				t.Errorf("Set(%q, %q) error = %v, wantErr %v", tt.key, tt.value, err, tt.wantErr)
				return
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

func TestKVStoreGetAll(t *testing.T) {
	kvs := NewKVStore()

	kvs.hashMap["foo"] = "bar"
	kvs.hashMap["baz"] = "qux"

	if got := kvs.GetAll(); got != "foo:bar\nbaz:qux\n" {
		t.Errorf("GetAll() = %v, want %v", got, "foo:bar\nbaz:qux\n")
		return
	}

}
