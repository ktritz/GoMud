package plugins

import "testing"

func TestPluginReadIntoStructReturnsUnmarshalError(t *testing.T) {
	tmpDir := t.TempDir()
	oldWriteFolder := writeFolderPath
	writeFolderPath = tmpDir
	t.Cleanup(func() {
		writeFolderPath = oldWriteFolder
	})

	p := &Plugin{name: "plugin_test_invalid_yaml", version: "1.0"}

	if err := p.WriteBytes("state", []byte("not: [valid")); err != nil {
		t.Fatalf("WriteBytes() error = %v", err)
	}

	var out struct {
		Value string `yaml:"value"`
	}

	if err := p.ReadIntoStruct("state", &out); err == nil {
		t.Fatal("ReadIntoStruct() error = nil, want unmarshal error")
	}
}

func TestPluginReadIntoStructRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	oldWriteFolder := writeFolderPath
	writeFolderPath = tmpDir
	t.Cleanup(func() {
		writeFolderPath = oldWriteFolder
	})

	p := &Plugin{name: "plugin_test_roundtrip", version: "1.0"}

	in := struct {
		Value string `yaml:"value"`
		Count int    `yaml:"count"`
	}{
		Value: "ok",
		Count: 7,
	}

	if err := p.WriteStruct("state", in); err != nil {
		t.Fatalf("WriteStruct() error = %v", err)
	}

	var out struct {
		Value string `yaml:"value"`
		Count int    `yaml:"count"`
	}

	if err := p.ReadIntoStruct("state", &out); err != nil {
		t.Fatalf("ReadIntoStruct() error = %v", err)
	}

	if out != in {
		t.Fatalf("ReadIntoStruct() = %+v, want %+v", out, in)
	}
}

func TestLogInitError(t *testing.T) {
	if LogInitError("test-component", nil) {
		t.Fatal("LogInitError(nil) = true, want false")
	}

	if !LogInitError("test-component", assertError("boom")) {
		t.Fatal("LogInitError(err) = false, want true")
	}
}

type assertError string

func (e assertError) Error() string { return string(e) }
