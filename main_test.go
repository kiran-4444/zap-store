package main

import (
	"bufio"
	"bytes"
	"kv-store/internal/storage/inmem"
	"kv-store/internal/zapstore"
	"strings"
	"testing"
)

func TestRunLoop(t *testing.T) {
	tests := []struct {
		name       string
		input      string            // Simulated user input with \n
		wantOutput string            // Expected output to out
		wantMap    map[string]string // Expected ZapStore state
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
			wantOutput: "> > Deleted foo\n> key not found\n> Exiting...\n",
			wantMap:    map[string]string{},
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
			wantOutput: "> key cannot be empty\n> Exiting...\n",
			wantMap:    map[string]string{},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var storageEngine = inmem.NewInMemStorageEngine()
			kvs := zapstore.NewZapStore(storageEngine)
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

			// Check ZapStore state
			for key, wantVal := range tt.wantMap {
				if got, err := kvs.Get(key); err != nil || got != wantVal {
					t.Errorf("kvs.Get(%q) = %q, err = %v; want %q, nil", key, got, err, wantVal)
				}
			}

		})
	}
}
