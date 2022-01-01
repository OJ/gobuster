package gobusterdir

import (
	"testing"

	"github.com/OJ/gobuster/v3/libgobuster"
)

const nothing = `[+] Url:        
[+] Method:     
[+] Threads:    0
[+] Wordlist:   
[+] Timeout:    0s`

func TestGobusterDir_GetConfigString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		d       *GobusterDir
		want    string
		wantErr bool
	}{
		{"ok", &GobusterDir{
			options:    &OptionsDir{},
			globalopts: &libgobuster.Options{},
			http:       &libgobuster.HTTPClient{},
		}, nothing, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.d.GetConfigString()
			if (err != nil) != tt.wantErr {
				t.Errorf("GobusterDir.GetConfigString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("GobusterDir.GetConfigString() = %q, want %q", got, tt.want)
			}
		})
	}
}
