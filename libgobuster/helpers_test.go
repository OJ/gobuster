package libgobuster

import (
	"errors"
	"fmt"
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
		t.Fatalf("Unexpected string size. Should have 1 (dedup) Got %v", len(x.Set))
	}

	y := NewSet[int]()
	y.Add(1)
	y.Add(1)
	if len(y.Set) != 1 {
		t.Fatalf("Unexpected int size. Should have 1 (depdup) Got %v", len(y.Set))
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
		t.Fatalf("Unexpected string size. Should have 2 (dedup) Got %v", len(x.Set))
	}

	y := NewSet[int]()
	y.AddRange([]int{1, 2, 1, 2})
	if len(y.Set) != 2 {
		t.Fatalf("Unexpected int size. Should have 2 (dedup) Got %v", len(y.Set))
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
		{"Trailing newline", "TestString\nTest\n1234\n", 4},
		{"3 Lines cr lf", "TestString\r\nTest\r\n1234", 3},
		{"Empty", "", 1},
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

func TestLineCounterError(t *testing.T) {
	t.Parallel()
	r := iotest.TimeoutReader(strings.NewReader("test"))
	_, err := lineCounter(r)
	if !errors.Is(err, iotest.ErrTimeout) {
		t.Fatalf("Got wrong error! %v", err)
	}
}
