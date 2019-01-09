package gobusterdir

import (
	"reflect"
	"testing"

	"github.com/OJ/gobuster/v3/libgobuster"
)

func TestParseExtensions(t *testing.T) {
	t.Parallel()

	var tt = []struct {
		testName           string
		Extensions         string
		expectedExtensions libgobuster.StringSet
		expectedError      string
	}{
		{"Valid extensions", "php,asp,txt", libgobuster.StringSet{Set: map[string]bool{"php": true, "asp": true, "txt": true}}, ""},
		{"Spaces", "php, asp , txt", libgobuster.StringSet{Set: map[string]bool{"php": true, "asp": true, "txt": true}}, ""},
		{"Double extensions", "php,asp,txt,php,asp,txt", libgobuster.StringSet{Set: map[string]bool{"php": true, "asp": true, "txt": true}}, ""},
		{"Leading dot", ".php,asp,.txt", libgobuster.StringSet{Set: map[string]bool{"php": true, "asp": true, "txt": true}}, ""},
		{"Empty string", "", libgobuster.NewStringSet(), "invalid extension string provided"},
	}

	for _, x := range tt {
		t.Run(x.testName, func(t *testing.T) {
			o := NewOptionsDir()
			o.Extensions = x.Extensions
			err := o.ParseExtensions()
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
		expectedCodes libgobuster.IntSet
		expectedError string
	}{
		{"Valid codes", "200,100,202", libgobuster.IntSet{Set: map[int]bool{100: true, 200: true, 202: true}}, ""},
		{"Spaces", "200, 100 , 202", libgobuster.IntSet{Set: map[int]bool{100: true, 200: true, 202: true}}, ""},
		{"Double codes", "200, 100, 202, 100", libgobuster.IntSet{Set: map[int]bool{100: true, 200: true, 202: true}}, ""},
		{"Invalid code", "200,AAA", libgobuster.NewIntSet(), "invalid status code given: AAA"},
		{"Invalid integer", "2000000000000000000000000000000", libgobuster.NewIntSet(), "invalid status code given: 2000000000000000000000000000000"},
		{"Empty string", "", libgobuster.NewIntSet(), "invalid status code string provided"},
	}

	for _, x := range tt {
		t.Run(x.testName, func(t *testing.T) {
			o := NewOptionsDir()
			o.StatusCodes = x.stringCodes
			err := o.ParseStatusCodes()
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
