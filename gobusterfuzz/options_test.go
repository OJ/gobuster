package gobusterfuzz

import "testing"

func TestNewOptions(t *testing.T) {
	t.Parallel()

	o := NewOptionsFuzz()
	if o.StatusCodesParsed.Set == nil {
		t.Fatal("StatusCodesParsed not initialized")
	}

	if o.ExtensionsParsed.Set == nil {
		t.Fatal("ExtensionsParsed not initialized")
	}
}
