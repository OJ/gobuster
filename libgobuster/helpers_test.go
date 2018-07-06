package libgobuster

import (
	"strings"
	"testing"
)

func TestNewStringSet(t *testing.T) {
	if newStringSet().Set == nil {
		t.Fatal("newStringSet returned nil Set")
	}
}

func TestNewIntSet(t *testing.T) {
	if newIntSet().Set == nil {
		t.Fatal("newIntSet returned nil Set")
	}
}

func TestStringSetAdd(t *testing.T) {
	x := newStringSet()
	x.Add("test")
	if len(x.Set) != 1 {
		t.Fatalf("Unexptected size. Should have 1 Got %v", len(x.Set))
	}
}

func TestStringSetAddDouble(t *testing.T) {
	x := newStringSet()
	x.Add("test")
	x.Add("test")
	if len(x.Set) != 1 {
		t.Fatalf("Unexptected size. Should have 1 Got %d", len(x.Set))
	}
}

func TestStringSetAddRange(t *testing.T) {
	x := newStringSet()
	x.AddRange([]string{"asdf", "ghjk"})
	if len(x.Set) != 2 {
		t.Fatalf("Unexptected size. Should have 2 Got %d", len(x.Set))
	}
}

func TestStringSetAddRangeDouble(t *testing.T) {
	x := newStringSet()
	x.AddRange([]string{"asdf", "ghjk", "asdf", "ghjk"})
	if len(x.Set) != 2 {
		t.Fatalf("Unexptected size. Should have 2 Got %d", len(x.Set))
	}
}

func TestStringSetContains(t *testing.T) {
	x := newStringSet()
	v := []string{"asdf", "ghjk", "1234", "5678"}
	x.AddRange(v)
	for _, y := range v {
		if !x.Contains(y) {
			t.Fatalf("Did not find value %s in array. %v", y, x.Set)
		}
	}
}

func TestStringSetContainsAny(t *testing.T) {
	x := newStringSet()
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
	x := newStringSet()
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
	x := newIntSet()
	x.Add(1)
	if len(x.Set) != 1 {
		t.Fatalf("Unexptected size. Should have 1 Got %d", len(x.Set))
	}
}

func TestIntSetAddDouble(t *testing.T) {
	x := newIntSet()
	x.Add(1)
	x.Add(1)
	if len(x.Set) != 1 {
		t.Fatalf("Unexptected size. Should have 1 Got %d", len(x.Set))
	}
}

func TestIntSetContains(t *testing.T) {
	x := newIntSet()
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
	x := newIntSet()
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
