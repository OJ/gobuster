package dir

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/OJ/gobuster/v3/lib"
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
			globalopts: &lib.Options{},
			http:       &lib.HTTPClient{},
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

func TestNewGobusterDir(t *testing.T) {
	type args struct {
		globalopts *lib.Options
		opts       *OptionsDir
	}
	tests := []struct {
		name    string
		args    args
		want    *GobusterDir
		wantErr bool
	}{
		{"no", args{nil, nil}, nil, true},
		{"no", args{&lib.Options{
			Threads:        0,
			Wordlist:       "",
			PatternFile:    "",
			Patterns:       []string{},
			OutputFilename: "test.txt",
			NoStatus:       false,
			NoProgress:     false,
			NoError:        false,
			Quiet:          false,
			Verbose:        false,
			Delay:          0,
		}, nil}, nil, true},
		{"yes", args{&lib.Options{
			Threads:        0,
			Wordlist:       "rockyou.txt",
			PatternFile:    "",
			Patterns:       []string{},
			OutputFilename: "test.txt",
			NoStatus:       false,
			NoProgress:     false,
			NoError:        false,
			Quiet:          false,
			Verbose:        false,
			Delay:          0,
		}, &OptionsDir{
			HTTPOptions:                lib.HTTPOptions{},
			Extensions:                 "",
			ExtensionsParsed:           lib.StringSet{},
			StatusCodes:                "",
			StatusCodesParsed:          lib.IntSet{},
			StatusCodesBlacklist:       "",
			StatusCodesBlacklistParsed: lib.IntSet{},
			UseSlash:                   false,
			HideLength:                 false,
			Expanded:                   false,
			NoStatus:                   false,
			DiscoverBackup:             false,
			ExcludeLength:              []int{},
		}}, &GobusterDir{globalopts: &lib.Options{Wordlist: "rockyou.txt"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewGobusterDir(tt.args.globalopts, tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewGobusterDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) && (got.globalopts.Wordlist != tt.want.globalopts.Wordlist) {
				fmt.Printf("%+v\n", got.globalopts)
				t.Errorf("NewGobusterDir() = %q, want %q", got.globalopts.Wordlist, tt.want.globalopts.Wordlist)
			}
		})
	}
}
