package libgobuster

import (
	"reflect"
	"testing"
)

func TestNewOptions(t *testing.T) {
	t.Parallel()

	o := NewOptions()
	if o.StatusCodesParsed.Set == nil {
		t.Fatal("StatusCodesParsed not initialized")
	}

	if o.ExtensionsParsed.Set == nil {
		t.Fatal("ExtensionsParsed not initialized")
	}
}

func TestParseExtensions(t *testing.T) {
	t.Parallel()

	var tt = []struct {
		testName           string
		Extensions         string
		expectedExtensions stringSet
		expectedError      string
	}{
		{"Valid extensions", "php,asp,txt", stringSet{Set: map[string]bool{"php": true, "asp": true, "txt": true}}, ""},
		{"Spaces", "php, asp , txt", stringSet{Set: map[string]bool{"php": true, "asp": true, "txt": true}}, ""},
		{"Double extensions", "php,asp,txt,php,asp,txt", stringSet{Set: map[string]bool{"php": true, "asp": true, "txt": true}}, ""},
		{"Leading dot", ".php,asp,.txt", stringSet{Set: map[string]bool{"php": true, "asp": true, "txt": true}}, ""},
		{"Empty string", "", newStringSet(), "invalid extension string provided"},
	}

	for _, x := range tt {
		t.Run(x.testName, func(t *testing.T) {
			o := NewOptions()
			o.Extensions = x.Extensions
			err := o.parseExtensions()
			if x.expectedError != "" {
				if err.Error() != x.expectedError {
					t.Fatalf("Expected error %q but got %q", x.expectedError, err.Error())
				}
			} else if !reflect.DeepEqual(x.expectedExtensions, o.ExtensionsParsed) {
				t.Fatalf("Expected %v but got %v", x.expectedExtensions, o.ExtensionsParsed)
			}
		})
	}
}

func TestParseStatusCodes(t *testing.T) {
	t.Parallel()

	var tt = []struct {
		testName      string
		stringCodes   string
		expectedCodes intSet
		expectedError string
	}{
		{"Valid codes", "200,100,202", intSet{Set: map[int]bool{100: true, 200: true, 202: true}}, ""},
		{"Spaces", "200, 100 , 202", intSet{Set: map[int]bool{100: true, 200: true, 202: true}}, ""},
		{"Double codes", "200, 100, 202, 100", intSet{Set: map[int]bool{100: true, 200: true, 202: true}}, ""},
		{"Invalid code", "200,AAA", newIntSet(), "invalid status code given: AAA"},
		{"Invalid integer", "2000000000000000000000000000000", newIntSet(), "invalid status code given: 2000000000000000000000000000000"},
		{"Empty string", "", newIntSet(), "invalid status code string provided"},
	}

	for _, x := range tt {
		t.Run(x.testName, func(t *testing.T) {
			o := NewOptions()
			o.StatusCodes = x.stringCodes
			err := o.parseStatusCodes()
			if x.expectedError != "" {
				if err.Error() != x.expectedError {
					t.Fatalf("Expected error %q but got %q", x.expectedError, err.Error())
				}
			} else if !reflect.DeepEqual(x.expectedCodes, o.StatusCodesParsed) {
				t.Fatalf("Expected %v but got %v", x.expectedCodes, o.StatusCodesParsed)
			}
		})
	}
}
