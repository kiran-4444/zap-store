package inmem

import "testing"

func TestInMemStorageEngineSet(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		value      string
		wantErr    bool
		wantErrMsg string
		wantMapVal string
	}{
		{name: "valid", key: "key", value: "value", wantErr: false, wantErrMsg: "", wantMapVal: "value"},
		{name: "empty_key", key: "", value: "value2", wantErr: true, wantErrMsg: "key cannot be empty"},
		{name: "empty_value", key: "key2", value: "", wantErr: false, wantErrMsg: "", wantMapVal: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var inMemStorageEngine = NewInMemStorageEngine()
			err := inMemStorageEngine.Set(tt.key, tt.value)

			if err != nil {
				if !tt.wantErr {
					t.Errorf("Set(%q, %q) error = %v, wantErr %v", tt.key, tt.value, err, tt.wantErr)
					return
				}
			}

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
			}

			if got, _ := inMemStorageEngine.Get(tt.key); got != tt.wantMapVal {
				t.Errorf("Set(%q, %q) map[%q] = %q, want %q", tt.key, tt.value, tt.key, got, tt.wantMapVal)
				return
			}
		})
	}

}

func TestInMemStorageEngineGet(t *testing.T) {
	var inMemStorageEngine = NewInMemStorageEngine()
	inMemStorageEngine.Set("foo", "bar")

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
			got, err := inMemStorageEngine.Get(tt.key)

			if err != nil {
				if !tt.wantErr {
					t.Errorf("Set(%q, %q) error = %v, wantErr %v", tt.key, tt.value, err, tt.wantErr)
					return
				}
			}

			if err == nil {
				if tt.wantErr {
					t.Errorf("Set(%q, %q) error = %v, wantErr %v", tt.key, tt.value, err, tt.wantErr)
					return
				}
			}

			if got != tt.wantMapVal {
				t.Errorf("Set(%q, %q) map[%q] = %q, want %q", tt.key, tt.value, tt.key, got, tt.wantMapVal)
				return
			}
		})
	}
}

func TestInMemStorageEngineDel(t *testing.T) {
	var inMemStorageEngine = NewInMemStorageEngine()
	inMemStorageEngine.Set("foo", "bar")

	tests := []struct {
		name       string
		key        string
		wantErr    bool
		wantErrMsg string
		wantMapVal string
	}{
		{name: "existent_key", key: "foo", wantErr: false, wantErrMsg: "", wantMapVal: ""},
		{name: "non_existent_key", key: "baz", wantErr: false, wantErrMsg: "", wantMapVal: ""},
	}

	for _, tt := range tests {
		inMemStorageEngine.Delete(tt.key)

		if got, _ := inMemStorageEngine.Get(tt.key); got != "" {
			t.Errorf("Delete() = %v, want %v", got, tt.wantMapVal)
		}
	}
}
