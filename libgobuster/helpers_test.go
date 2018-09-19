package libgobuster

import (
	"strings"
	"testing"
	"testing/iotest"
)

func TestNewStringSet(t *testing.T) {
	if NewStringSet().Set == nil {
		t.Fatal("newStringSet returned nil Set")
	}
}

func TestNewIntSet(t *testing.T) {
	if NewIntSet().Set == nil {
		t.Fatal("newIntSet returned nil Set")
	}
}

func TestStringSetAdd(t *testing.T) {
	x := NewStringSet()
	x.Add("test")
	if len(x.Set) != 1 {
		t.Fatalf("Unexptected size. Should have 1 Got %v", len(x.Set))
	}
}

func TestStringSetAddDouble(t *testing.T) {
	x := NewStringSet()
	x.Add("test")
	x.Add("test")
	if len(x.Set) != 1 {
		t.Fatalf("Unexptected size. Should have 1 Got %d", len(x.Set))
	}
}

func TestStringSetAddRange(t *testing.T) {
	x := NewStringSet()
	x.AddRange([]string{"asdf", "ghjk"})
	if len(x.Set) != 2 {
		t.Fatalf("Unexptected size. Should have 2 Got %d", len(x.Set))
	}
}

func TestStringSetAddRangeDouble(t *testing.T) {
	x := NewStringSet()
	x.AddRange([]string{"asdf", "ghjk", "asdf", "ghjk"})
	if len(x.Set) != 2 {
		t.Fatalf("Unexptected size. Should have 2 Got %d", len(x.Set))
	}
}

func TestStringSetContains(t *testing.T) {
	x := NewStringSet()
	v := []string{"asdf", "ghjk", "1234", "5678"}
	x.AddRange(v)
	for _, y := range v {
		if !x.Contains(y) {
			t.Fatalf("Did not find value %s in array. %v", y, x.Set)
		}
	}
}

func TestStringSetContainsAny(t *testing.T) {
	x := NewStringSet()
	v := []string{"asdf", "ghjk", "1234", "5678"}
	x.AddRange(v)
	if !x.ContainsAny(v) {
		t.Fatalf("Did not find any")
	}

	// test not found
	if x.ContainsAny([]string{"mmmm", "nnnnn"}) {
		t.Fatal("Found unexpected values")
	}
}

func TestStringSetStringify(t *testing.T) {
	x := NewStringSet()
	v := []string{"asdf", "ghjk", "1234", "5678"}
	x.AddRange(v)
	z := x.Stringify()
	// order is random
	for _, y := range v {
		if !strings.Contains(z, y) {
			t.Fatalf("Did not find value %q in %q", y, z)
		}
	}
}

func TestIntSetAdd(t *testing.T) {
	x := NewIntSet()
	x.Add(1)
	if len(x.Set) != 1 {
		t.Fatalf("Unexptected size. Should have 1 Got %d", len(x.Set))
	}
}

func TestIntSetAddDouble(t *testing.T) {
	x := NewIntSet()
	x.Add(1)
	x.Add(1)
	if len(x.Set) != 1 {
		t.Fatalf("Unexptected size. Should have 1 Got %d", len(x.Set))
	}
}

func TestIntSetContains(t *testing.T) {
	x := NewIntSet()
	v := []int{1, 2, 3, 4}
	for _, y := range v {
		x.Add(y)
	}
	for _, y := range v {
		if !x.Contains(y) {
			t.Fatalf("Did not find value %d in array. %v", y, x.Set)
		}
	}
}

func TestIntSetStringify(t *testing.T) {
	x := NewIntSet()
	v := []int{1, 3, 2, 4}
	expected := "1,2,3,4"
	for _, y := range v {
		x.Add(y)
	}
	z := x.Stringify()
	// should be sorted
	if expected != z {
		t.Fatalf("Expected %q got %q", expected, z)
	}
}

func TestLineCounter(t *testing.T) {
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
		t.Run(x.testName, func(t *testing.T) {
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
	r := iotest.TimeoutReader(strings.NewReader("test"))
	_, err := lineCounter(r)
	if err != iotest.ErrTimeout {
		t.Fatalf("Got wrong error! %v", err)
	}
}
