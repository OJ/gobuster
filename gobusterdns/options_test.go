package gobusterdns

import "testing"

func TestNewOptions(t *testing.T) {
	t.Parallel()

	o := NewOptionsDNS()
	if o.WildcardIps.Set == nil {
		t.Fatal("WildcardIps not initialized")
	}
}
