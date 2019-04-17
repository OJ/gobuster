package helper

import (
	"reflect"
	"testing"

	"github.com/OJ/gobuster/v3/libgobuster"
)

func TestParseExtensions(t *testing.T) {
	t.Parallel()

	var tt = []struct {
		testName           string
		extensions         string
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
			ret, err := ParseExtensions(x.extensions)
			if x.expectedError != "" {
				if err.Error() != x.expectedError {
					t.Fatalf("Expected error %q but got %q", x.expectedError, err.Error())
				}
			} else if !reflect.DeepEqual(x.expectedExtensions, ret) {
				t.Fatalf("Expected %v but got %v", x.expectedExtensions, ret)
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
			ret, err := ParseStatusCodes(x.stringCodes)
			if x.expectedError != "" {
				if err.Error() != x.expectedError {
					t.Fatalf("Expected error %q but got %q", x.expectedError, err.Error())
				}
			} else if !reflect.DeepEqual(x.expectedCodes, ret) {
				t.Fatalf("Expected %v but got %v", x.expectedCodes, ret)
			}
		})
	}
}

func BenchmarkParseExtensions(b *testing.B) {
	var tt = []struct {
		testName           string
		extensions         string
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
		b.Run(x.testName, func(b2 *testing.B) {
			for y := 0; y < b2.N; y++ {
				_, _ = ParseExtensions(x.extensions)
			}
		})
	}
}

func BenchmarkParseStatusCodes(b *testing.B) {
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
		b.Run(x.testName, func(b2 *testing.B) {
			for y := 0; y < b2.N; y++ {
				_, _ = ParseStatusCodes(x.stringCodes)
			}
		})
	}
}
