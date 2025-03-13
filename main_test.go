package main

import (
	"bufio"
	"bytes"
	"strings"
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
		{name: "empty_key", key: "", value: "value2", wantErr: true, wantErrMsg: "KEY CANNOT BE EMPTY", wantMapVal: ""},
		{name: "empty_value", key: "key2", value: "", wantErr: true, wantErrMsg: "VALUE CANNOT BE EMPTY", wantMapVal: ""},
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
		{name: "non_existent_key", key: "baz", value: "", wantErr: true, wantErrMsg: "KEY NOT FOUND", wantMapVal: ""},
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

func TestKVStoreGetAll(t *testing.T) {
	kvs := NewKVStore()

	kvs.hashMap["foo"] = "bar"
	kvs.hashMap["baz"] = "qux"

	if got := kvs.GetAll(); got != "baz:qux\nfoo:bar\n" {
		t.Errorf("GetAll() = %v, want %v", got, "foo:bar\nbaz:qux\n")
		return
	}

}

func TestRunLoop(t *testing.T) {
	tests := []struct {
		name       string
		input      string            // Simulated user input with \n
		wantOutput string            // Expected output to out
		wantMap    map[string]string // Expected KVStore state
		wantErr    bool              // Expect an error?
	}{
		{
			name:       "set_and_get",
			input:      "set foo bar\nget foo\nexit\n",
			wantOutput: "> > bar\n> Exiting...\n",
			wantMap:    map[string]string{"foo": "bar"},
			wantErr:    false,
		},
		{
			name:       "delete_key",
			input:      "set foo bar\ndel foo\nget foo\nexit\n",
			wantOutput: "> > Deleted foo\n> KEY NOT FOUND\n> Exiting...\n",
			wantMap:    map[string]string{},
			wantErr:    false,
		},
		{
			name:       "getall_sorted",
			input:      "set baz qux\nset foo bar\ngetall\nexit\n",
			wantOutput: "> > > baz:qux\nfoo:bar\n> Exiting...\n",
			wantMap:    map[string]string{"baz": "qux", "foo": "bar"},
			wantErr:    false,
		},
		{
			name:       "invalid_command",
			input:      "bad cmd\nexit\n",
			wantOutput: "> Invalid command: bad\n> Exiting...\n",
			wantMap:    map[string]string{},
			wantErr:    false,
		},
		{
			name:       "invalid_command_getall",
			input:      "getallinvalid\nexit\n",
			wantOutput: "> Invalid command: getallinvalid\n> Exiting...\n",
			wantMap:    map[string]string{},
			wantErr:    false,
		},
		{
			name:       "io_error",
			input:      "set foo bar",
			wantOutput: "> ",
			wantMap:    map[string]string{},
			wantErr:    true,
		},
		{
			name:       "empty_input",
			input:      "\nset foo bar\nexit\n",
			wantOutput: "> > > Exiting...\n",
			wantMap:    map[string]string{"foo": "bar"},
			wantErr:    false,
		},
		{
			name:       "invalid_set",
			input:      "a b c\nexit\n",
			wantOutput: "> Invalid command: a\n> Exiting...\n",
			wantMap:    map[string]string{},
			wantErr:    false,
		},
		{
			name:       "set_empty_key",
			input:      "set  val\nexit\n",
			wantOutput: "> KEY CANNOT BE EMPTY\n> Exiting...\n",
			wantMap:    map[string]string{},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			kvs := NewKVStore()
			reader := bufio.NewReader(strings.NewReader(tt.input))
			var buf bytes.Buffer

			// Run
			err := runLoop(kvs, reader, &buf)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("runLoop() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check output
			if got := buf.String(); got != tt.wantOutput {
				t.Errorf("runLoop() output = %q, want %q", got, tt.wantOutput)
			}

			// Check KVStore state
			for key, wantVal := range tt.wantMap {
				if got, err := kvs.Get(key); err != nil || got != wantVal {
					t.Errorf("kvs.Get(%q) = %q, err = %v; want %q, nil", key, got, err, wantVal)
				}
			}
			// Verify missing keys
			for key := range kvs.hashMap {
				if _, ok := tt.wantMap[key]; !ok {
					t.Errorf("Unexpected key %q in hashMap", key)
				}
			}
		})
	}
}
