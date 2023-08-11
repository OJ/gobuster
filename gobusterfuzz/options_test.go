package gobusterfuzz

import "testing"

func TestNewOptions(t *testing.T) {
	t.Parallel()

	o := NewOptionsFuzz()
	if o.ExcludedStatusCodesParsed.Set == nil {
		t.Fatal("StatusCodesParsed not initialized")
	}
}
