package libgobuster

import (
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
	"testing/iotest"
)

func TestNewSet(t *testing.T) {
	t.Parallel()
	if NewSet[string]().Set == nil {
		t.Fatal("NewSet[string] returned nil Set")
	}

	if NewSet[int]().Set == nil {
		t.Fatal("NewSet[int] returned nil Set")
	}
}

func TestSetAdd(t *testing.T) {
	t.Parallel()
	x := NewSet[string]()
	x.Add("test")
	if len(x.Set) != 1 {
		t.Fatalf("Unexpected string size. Should have 1 Got %v", len(x.Set))
	}

	y := NewSet[int]()
	y.Add(1)
	if len(y.Set) != 1 {
		t.Fatalf("Unexpected int size. Should have 1 Got %v", len(y.Set))
	}
}

func TestSetAddDouble(t *testing.T) {
	t.Parallel()
	x := NewSet[string]()
	x.Add("test")
	x.Add("test")
	if len(x.Set) != 1 {
		t.Fatalf("Unexpected string size. Should be 1 (unique) Got %v", len(x.Set))
	}

	y := NewSet[int]()
	y.Add(1)
	y.Add(1)
	if len(y.Set) != 1 {
		t.Fatalf("Unexpected int size. Should be 1 (unique) Got %v", len(y.Set))
	}
}

func TestSetAddRange(t *testing.T) {
	t.Parallel()
	x := NewSet[string]()
	x.AddRange([]string{"string1", "string2"})
	if len(x.Set) != 2 {
		t.Fatalf("Unexpected string size. Should have 2 Got %v", len(x.Set))
	}

	y := NewSet[int]()
	y.AddRange([]int{1, 2})
	if len(y.Set) != 2 {
		t.Fatalf("Unexpected int size. Should have 2 Got %v", len(y.Set))
	}
}

func TestSetAddRangeDouble(t *testing.T) {
	t.Parallel()
	x := NewSet[string]()
	x.AddRange([]string{"string1", "string2", "string1", "string2"})
	if len(x.Set) != 2 {
		t.Fatalf("Unexpected string size. Should be 2 (unique) Got %v", len(x.Set))
	}

	y := NewSet[int]()
	y.AddRange([]int{1, 2, 1, 2})
	if len(y.Set) != 2 {
		t.Fatalf("Unexpected int size. Should be 2 (unique) Got %v", len(y.Set))
	}
}

func TestSetContains(t *testing.T) {
	t.Parallel()
	x := NewSet[string]()
	v := []string{"string1", "string2", "1234", "5678"}
	x.AddRange(v)
	for _, i := range v {
		if !x.Contains(i) {
			t.Fatalf("Did not find value %s in array. %v", i, x.Set)
		}
	}

	y := NewSet[int]()
	v2 := []int{1, 2312, 123121, 999, -99}
	y.AddRange(v2)
	for _, i := range v2 {
		if !y.Contains(i) {
			t.Fatalf("Did not find value %d in array. %v", i, y.Set)
		}
	}
}

func TestSetContainsAny(t *testing.T) {
	t.Parallel()
	x := NewSet[string]()
	v := []string{"string1", "string2", "1234", "5678"}
	x.AddRange(v)
	if !x.ContainsAny(v) {
		t.Fatalf("Did not find any")
	}

	// test not found
	if x.ContainsAny([]string{"mmmm", "nnnnn"}) {
		t.Fatal("Found unexpected values")
	}

	y := NewSet[int]()
	v2 := []int{1, 2312, 123121, 999, -99}
	y.AddRange(v2)
	if !y.ContainsAny(v2) {
		t.Fatalf("Did not find any")
	}

	// test not found
	if y.ContainsAny([]int{9235, 2398532}) {
		t.Fatal("Found unexpected values")
	}
}

func TestSetStringify(t *testing.T) {
	t.Parallel()
	x := NewSet[string]()
	v := []string{"string1", "string2", "1234", "5678"}
	x.AddRange(v)
	z := x.Stringify()
	// order is random
	for _, i := range v {
		if !strings.Contains(z, i) {
			t.Fatalf("Did not find value %q in %q", i, z)
		}
	}

	y := NewSet[int]()
	v2 := []int{1, 2312, 123121, 999, -99}
	y.AddRange(v2)
	z = y.Stringify()
	// order is random
	for _, i := range v2 {
		if !strings.Contains(z, fmt.Sprint(i)) {
			t.Fatalf("Did not find value %q in %q", i, z)
		}
	}
}

func TestLineCounter(t *testing.T) {
	t.Parallel()
	var tt = []struct {
		testName string
		s        string
		expected int
	}{
		{"One Line", "test", 1},
		{"3 Lines", "TestString\nTest\n1234", 3},
		{"Trailing newline", "TestString\nTest\n1234\n", 3},
		{"3 Lines cr lf", "TestString\r\nTest\r\n1234", 3},
		{"Empty", "", 1},       // these are wrong, but I've found no good way to handle those
		{"Empty 2", "\n", 1},   // these are wrong, but I've found no good way to handle those
		{"Empty 3", "\r\n", 1}, // these are wrong, but I've found no good way to handle those
	}
	for _, x := range tt {
		x := x // NOTE: https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		t.Run(x.testName, func(t *testing.T) {
			t.Parallel()
			r := strings.NewReader(x.s)
			l, err := lineCounter(r)
			if err != nil {
				t.Fatalf("Got error: %v", err)
			}
			if l != x.expected {
				t.Fatalf("wrong line count! Got %d expected %d", l, x.expected)
			}
		})
	}
}

func TestLineCounterSlow(t *testing.T) {
	t.Parallel()
	var tt = []struct {
		testName string
		s        string
		expected int
	}{
		{"One Line", "test", 1},
		{"3 Lines", "TestString\nTest\n1234", 3},
		{"Trailing newline", "TestString\nTest\n1234\n", 3},
		{"3 Lines cr lf", "TestString\r\nTest\r\n1234", 3},
		{"Empty", "", 0},
		{"Empty 2", "\n", 0},
		{"Empty 3", "\r\n", 0},
	}
	for _, x := range tt {
		x := x // NOTE: https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		t.Run(x.testName, func(t *testing.T) {
			t.Parallel()
			r := strings.NewReader(x.s)
			l, err := lineCounterSlow(r)
			if err != nil {
				t.Fatalf("Got error: %v", err)
			}
			if l != x.expected {
				t.Fatalf("wrong line count! Got %d expected %d", l, x.expected)
			}
		})
	}
}

func BenchmarkLineCounter(b *testing.B) {
	r, err := os.Open("../rockyou.txt")
	if err != nil {
		b.Fatalf("Got error: %v", err)
	}
	defer r.Close()
	for i := 0; i < b.N; i++ {
		_, err := r.Seek(0, io.SeekStart)
		if err != nil {
			b.Fatalf("Got error: %v", err)
		}
		c, err := lineCounter(r)
		if err != nil {
			b.Fatalf("Got error: %v", err)
		}
		if c != 14344391 {
			b.Errorf("invalid count. Expected 14344391, got %d", c)
		}
	}
}

func BenchmarkLineCounterSlow(b *testing.B) {
	r, err := os.Open("../rockyou.txt")
	if err != nil {
		b.Fatalf("Got error: %v", err)
	}
	defer r.Close()
	for i := 0; i < b.N; i++ {
		_, err := r.Seek(0, io.SeekStart)
		if err != nil {
			b.Fatalf("Got error: %v", err)
		}
		c, err := lineCounterSlow(r)
		if err != nil {
			b.Fatalf("Got error: %v", err)
		}
		if c != 14336792 {
			b.Errorf("invalid count. Expected 14336792, got %d", c)
		}
	}
}

func TestLineCounterError(t *testing.T) {
	t.Parallel()
	r := iotest.TimeoutReader(strings.NewReader("test"))
	_, err := lineCounter(r)
	if !errors.Is(err, iotest.ErrTimeout) {
		t.Fatalf("Got wrong error! %v", err)
	}
}

func TestParseExtensions(t *testing.T) {
	t.Parallel()
	var tt = []struct {
		testName           string
		extensions         string
		expectedExtensions Set[string]
		expectedError      string
	}{
		{"Valid extensions", "php,asp,txt", Set[string]{Set: map[string]bool{"php": true, "asp": true, "txt": true}}, ""},
		{"Spaces", "php, asp , txt", Set[string]{Set: map[string]bool{"php": true, "asp": true, "txt": true}}, ""},
		{"Double extensions", "php,asp,txt,php,asp,txt", Set[string]{Set: map[string]bool{"php": true, "asp": true, "txt": true}}, ""},
		{"Leading dot", ".php,asp,.txt", Set[string]{Set: map[string]bool{"php": true, "asp": true, "txt": true}}, ""},
		{"Empty string", "", NewSet[string](), "invalid extension string provided"},
	}

	for _, x := range tt {
		x := x // NOTE: https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		t.Run(x.testName, func(t *testing.T) {
			t.Parallel()
			ret, err := ParseExtensions(x.extensions)
			if x.expectedError != "" {
				if err != nil && err.Error() != x.expectedError {
					t.Fatalf("Expected error %q but got %q", x.expectedError, err.Error())
				}
			} else if !reflect.DeepEqual(x.expectedExtensions, ret) {
				t.Fatalf("Expected %v but got %v", x.expectedExtensions, ret)
			}
		})
	}
}

func TestParseCommaSeparatedInt(t *testing.T) {
	t.Parallel()
	var tt = []struct {
		stringCodes   string
		expectedCodes []int
		expectedError string
	}{
		{"200,100,202", []int{200, 100, 202}, ""},
		{"200, 100 , 202", []int{200, 100, 202}, ""},
		{"200, 100, 202, 100", []int{200, 100, 202}, ""},
		{"200,AAA", []int{}, "invalid string given: AAA"},
		{"2000000000000000000000000000000", []int{}, "invalid string given: 2000000000000000000000000000000"},
		{"", []int{}, "invalid string provided"},
		{"200-205", []int{200, 201, 202, 203, 204, 205}, ""},
		{"200-202,203-205", []int{200, 201, 202, 203, 204, 205}, ""},
		{"200-202,204-205", []int{200, 201, 202, 204, 205}, ""},
		{"200-202,205", []int{200, 201, 202, 205}, ""},
		{"205,200,100-101,103-105", []int{100, 101, 103, 104, 105, 200, 205}, ""},
		{"200-200", []int{200}, ""},
		{"200 - 202", []int{200, 201, 202}, ""},
		{"200 -202", []int{200, 201, 202}, ""},
		{"200- 202", []int{200, 201, 202}, ""},
		{"200              -                202", []int{200, 201, 202}, ""},
		{"230-200", []int{}, "invalid range given: 230-200"},
		{"A-200", []int{}, "invalid range given: A-200"},
		{"230-A", []int{}, "invalid range given: 230-A"},
		{"200,202-205,A,206-210", []int{}, "invalid string given: A"},
		{"200,202-205,A-1,206-210", []int{}, "invalid range given: A-1"},
		{"200,202-205,1-A,206-210", []int{}, "invalid range given: 1-A"},
	}

	for _, x := range tt {
		x := x // NOTE: https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		t.Run(x.stringCodes, func(t *testing.T) {
			t.Parallel()
			want := NewSet[int]()
			want.AddRange(x.expectedCodes)
			ret, err := ParseCommaSeparatedInt(x.stringCodes)
			if x.expectedError != "" {
				if err != nil && err.Error() != x.expectedError {
					t.Fatalf("Expected error %q but got %q", x.expectedError, err.Error())
				}
			} else if !reflect.DeepEqual(want, ret) {
				t.Fatalf("Expected %v but got %v", want, ret)
			}
		})
	}
}

func BenchmarkParseExtensions(b *testing.B) {
	var tt = []struct {
		testName           string
		extensions         string
		expectedExtensions Set[string]
		expectedError      string
	}{
		{"Valid extensions", "php,asp,txt", Set[string]{Set: map[string]bool{"php": true, "asp": true, "txt": true}}, ""},
		{"Spaces", "php, asp , txt", Set[string]{Set: map[string]bool{"php": true, "asp": true, "txt": true}}, ""},
		{"Double extensions", "php,asp,txt,php,asp,txt", Set[string]{Set: map[string]bool{"php": true, "asp": true, "txt": true}}, ""},
		{"Leading dot", ".php,asp,.txt", Set[string]{Set: map[string]bool{"php": true, "asp": true, "txt": true}}, ""},
		{"Empty string", "", NewSet[string](), "invalid extension string provided"},
	}

	for _, x := range tt {
		x := x // NOTE: https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		b.Run(x.testName, func(b2 *testing.B) {
			for y := 0; y < b2.N; y++ {
				_, _ = ParseExtensions(x.extensions)
			}
		})
	}
}

func BenchmarkParseCommaSeparatedInt(b *testing.B) {
	var tt = []struct {
		testName      string
		stringCodes   string
		expectedCodes Set[int]
		expectedError string
	}{
		{"Valid codes", "200,100,202", Set[int]{Set: map[int]bool{100: true, 200: true, 202: true}}, ""},
		{"Spaces", "200, 100 , 202", Set[int]{Set: map[int]bool{100: true, 200: true, 202: true}}, ""},
		{"Double codes", "200, 100, 202, 100", Set[int]{Set: map[int]bool{100: true, 200: true, 202: true}}, ""},
		{"Invalid code", "200,AAA", NewSet[int](), "invalid string given: AAA"},
		{"Invalid integer", "2000000000000000000000000000000", NewSet[int](), "invalid string given: 2000000000000000000000000000000"},
		{"Empty string", "", NewSet[int](), "invalid string string provided"},
	}

	for _, x := range tt {
		x := x // NOTE: https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		b.Run(x.testName, func(b2 *testing.B) {
			for y := 0; y < b2.N; y++ {
				_, _ = ParseCommaSeparatedInt(x.stringCodes)
			}
		})
	}
}
