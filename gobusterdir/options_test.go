package gobusterdir

import "testing"

func TestNewOptions(t *testing.T) {
	t.Parallel()

	o := NewOptionsDir()
	if o.StatusCodesParsed.Set == nil {
		t.Fatal("StatusCodesParsed not initialized")
	}

	if o.ExtensionsParsed.Set == nil {
		t.Fatal("ExtensionsParsed not initialized")
	}
}
