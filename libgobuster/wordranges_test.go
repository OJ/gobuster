package libgobuster

import "testing"

// helper function
func stringInList(s string, list []string) bool {
	for _, item := range list {
		if s == item {
			return true
		}
	}

	return false
}

func TestSingleNumericRange(t *testing.T) {
	url := "/foo/bar/[1-5]/trailing"

	expandedWords := ExpandWords(ParseTokens(url))
	expectedNumWords := 5
	numWords := len(expandedWords)
	if numWords != expectedNumWords {
		t.Errorf("Expected %d tokens, got %d", expectedNumWords, numWords)
	}
}

func TestNoRange(t *testing.T) {
	url := "/foo/bar/baz"

	expandedWords := ExpandWords(ParseTokens(url))
	if !stringInList(url, expandedWords) {
		t.Errorf("Could not find literal %s in expanded Words", url)
	}
}

func TestSingleWordWithRange(t *testing.T) {
	url := "foo-[1-5]-[12-13]"

	expandedWords := ExpandWords(ParseTokens(url))
	expectedWords := 10
	numWords := len(expandedWords)
	if numWords != expectedWords {
		t.Errorf("Expected %d tokens, got %d", expectedWords, numWords)
	}
}

func TestDottedRange(t *testing.T) {
	url := "foo-[1-5].[1-2].[12-13]"

	expandedWords := ExpandWords(ParseTokens(url))
	expectedWords := 20
	numWords := len(expandedWords)
	if numWords != expectedWords {
		t.Errorf("Expected %d tokens, got %d", expectedWords, numWords)
	}
	expectedWord1 := "foo-3.2.12"
	expectedWord2 := "foo-1.1.12"
	expectedWord3 := "foo-5.2.13"
	if !stringInList(expectedWord1, expandedWords) {
		t.Errorf("Expected Word not found: %d", expectedWord1)
	}
	if !stringInList(expectedWord2, expandedWords) {
		t.Errorf("Expected Word not found: %d", expectedWord2)
	}
	if !stringInList(expectedWord3, expandedWords) {
		t.Errorf("Expected Word not found: %d", expectedWord3)
	}
}

func TestMultipleNumericRange(t *testing.T) {
	url := "/foo/bar/[1-5]/baz/[12-13]"

	expandedWords := ExpandWords(ParseTokens(url))

	if len(expandedWords) != 10 {
		t.Errorf("Expected %d tokens, got %d", 10, len(expandedWords))
	}

	expectedWord1 := "/foo/bar/1/baz/12"
	expectedWord2 := "/foo/bar/5/baz/13"
	if !stringInList(expectedWord1, expandedWords) {
		t.Errorf("Expected Word not found: %d", expectedWord1)
	}
	if !stringInList(expectedWord2, expandedWords) {
		t.Errorf("Expected Word not found: %d", expectedWord2)
	}
}
