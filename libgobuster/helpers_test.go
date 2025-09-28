package libgobuster

import (
	"errors"
	"io"
	"os"
	"reflect"
	"strconv"
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
		if !strings.Contains(z, strconv.Itoa(i)) {
			t.Fatalf("Did not find value %q in %q", i, z)
		}
	}
}

func TestLineCounter(t *testing.T) {
	t.Parallel()
	tt := []struct {
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
	tt := []struct {
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
	for b.Loop() {
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
	for b.Loop() {
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
	tt := []struct {
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
	tt := []struct {
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
	tt := []struct {
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
		b.Run(x.testName, func(b2 *testing.B) {
			for b2.Loop() {
				_, _ = ParseExtensions(x.extensions)
			}
		})
	}
}

func BenchmarkParseCommaSeparatedInt(b *testing.B) {
	tt := []struct {
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
		b.Run(x.testName, func(b2 *testing.B) {
			for b2.Loop() {
				_, _ = ParseCommaSeparatedInt(x.stringCodes)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "unnamed",
		},
		{
			name:     "normal filename",
			input:    "test.txt",
			expected: "test.txt",
		},
		{
			name:     "filename with spaces",
			input:    "  test file.txt  ",
			expected: "test file.txt",
		},
		{
			name:     "filename with path separators",
			input:    "folder/test\\file.txt",
			expected: "folder_test_file.txt",
		},
		{
			name:     "filename with Windows invalid characters",
			input:    "test<file>name:with|invalid?chars*.txt",
			expected: "test_file_name_with_invalid_chars_.txt",
		},
		{
			name:     "filename with control characters",
			input:    "test\x00file\x1fname.txt",
			expected: "test_file_name.txt",
		},
		{
			name:     "filename with non-printable Unicode",
			input:    "test\u200bfile\u2028name.txt",
			expected: "test_file_name.txt",
		},
		{
			name:     "Windows reserved name - CON",
			input:    "CON.txt",
			expected: "_CON.txt",
		},
		{
			name:     "Windows reserved name - PRN (lowercase)",
			input:    "prn.log",
			expected: "_prn.log",
		},
		{
			name:     "Windows reserved name - COM1",
			input:    "COM1",
			expected: "_COM1",
		},
		{
			name:     "Windows reserved name - LPT9",
			input:    "lpt9.dat",
			expected: "_lpt9.dat",
		},
		{
			name:     "filename with reserved name as part of longer name",
			input:    "CONfig.txt",
			expected: "CONfig.txt",
		},
		{
			name:     "filename with trailing dots and spaces",
			input:    "test.txt..   ",
			expected: "test.txt",
		},
		{
			name:     "filename that becomes empty after sanitization",
			input:    "...   ",
			expected: "unnamed",
		},
		{
			name:     "very long filename",
			input:    strings.Repeat("a", 300) + ".txt",
			expected: strings.Repeat("a", 251) + ".txt", // 255 total
		},
		{
			name:     "long filename with long extension",
			input:    strings.Repeat("a", 250) + "." + strings.Repeat("b", 10),
			expected: strings.Repeat("a", 244) + "." + strings.Repeat("b", 10), // 255 total
		},
		{
			name:     "long filename where extension is too long",
			input:    "test." + strings.Repeat("b", 260),
			expected: ("test." + strings.Repeat("b", 260))[:255],
		},
		{
			name:     "whitespace only",
			input:    "   \t\n  ",
			expected: "unnamed",
		},
		{
			name:     "mixed invalid characters and reserved name",
			input:    "aux|with<invalid>chars.log",
			expected: "aux_with_invalid_chars.log",
		},
		{
			name:     "reserved name",
			input:    "AUX.log",
			expected: "_AUX.log",
		},
		{
			name:     "Unicode filename",
			input:    "тест файл.txt",
			expected: "тест файл.txt",
		},
		{
			name:     "filename with quotes",
			input:    `"quoted filename".txt`,
			expected: "_quoted filename_.txt",
		},
		{
			name:     "filename with pipe character",
			input:    "file|with|pipes.txt",
			expected: "file_with_pipes.txt",
		},
		{
			name:     "filename with question marks",
			input:    "what?is?this?.txt",
			expected: "what_is_this_.txt",
		},
		{
			name:     "filename with asterisks",
			input:    "wild*card*name*.txt",
			expected: "wild_card_name_.txt",
		},
		{
			name:     "only extension",
			input:    ".hidden",
			expected: ".hidden",
		},
		{
			name:     "reserved name without extension",
			input:    "NUL",
			expected: "_NUL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, result, tt.expected)
			}

			// Additional validation: ensure result is safe
			if len(result) > 255 {
				t.Errorf("sanitizeFilename(%q) returned filename too long: %d characters", tt.input, len(result))
			}

			if strings.ContainsAny(result, `<>:"|?*`) {
				t.Errorf("sanitizeFilename(%q) still contains invalid characters: %q", tt.input, result)
			}

			if strings.Contains(result, "/") || strings.Contains(result, "\\") {
				t.Errorf("sanitizeFilename(%q) still contains path separators: %q", tt.input, result)
			}

			if result == "" {
				t.Errorf("sanitizeFilename(%q) returned empty string", tt.input)
			}
		})
	}
}
