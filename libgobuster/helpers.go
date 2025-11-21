package libgobuster

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// Set is a set of Ts
type Set[T comparable] struct {
	Set map[T]bool
}

// NewSet creates a new initialized Set
func NewSet[T comparable]() Set[T] {
	return Set[T]{Set: map[T]bool{}}
}

// Add an element to a set
func (set *Set[T]) Add(s T) bool {
	_, found := set.Set[s]
	set.Set[s] = true
	return !found
}

// AddRange adds a list of elements to a set
func (set *Set[T]) AddRange(ss []T) {
	for _, s := range ss {
		set.Set[s] = true
	}
}

// Contains tests if an element is in a set
func (set *Set[T]) Contains(s T) bool {
	_, found := set.Set[s]
	return found
}

// ContainsAny checks if any of the elements exist
func (set *Set[T]) ContainsAny(ss []T) bool {
	for _, s := range ss {
		if set.Set[s] {
			return true
		}
	}
	return false
}

// Length returns the length of the Set
func (set *Set[T]) Length() int {
	return len(set.Set)
}

// Stringify the set
func (set *Set[T]) Stringify() string {
	values := make([]string, len(set.Set))
	i := 0
	for s := range set.Set {
		values[i] = fmt.Sprint(s)
		i++
	}
	return strings.Join(values, ",")
}

// this method is much faster than lineCounter_slow but has the following errors:
// - empty files are reported as 1 line
// - files only containing a newline are reported as 1 line
// - also counts lines with comments
func lineCounter(r io.Reader) (int, error) {
	buf := make([]byte, 32*1024)
	count := 1
	lineSep := []byte{'\n'}
	var lastChar byte

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		// store last character received if we got any bytes
		if c > 0 {
			lastChar = buf[c-1]
		}

		switch {
		case errors.Is(err, io.EOF):
			// account for trailing new line
			if lastChar == '\n' {
				count--
			}
			return count, nil

		case err != nil:
			return -1, err
		}
	}
}

func lineCounterSlow(r io.Reader) (int, error) {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)
	var count int
	for scanner.Scan() {
		w := scanner.Text()
		if w == "" {
			continue
		}

		count++
	}
	if err := scanner.Err(); err != nil {
		return -1, err
	}
	return count, nil
}

// DefaultUserAgent returns the default user agent to use in HTTP requests
func DefaultUserAgent() string {
	return fmt.Sprintf("gobuster/%s", VERSION)
}

// ParseExtensions parses the extensions provided as a comma separated list
func ParseExtensions(extensions string) (Set[string], error) {
	ret := NewSet[string]()

	if extensions == "" {
		return ret, nil
	}

	for _, e := range strings.Split(extensions, ",") {
		e = strings.TrimSpace(e)
		// remove leading . from extensions
		ret.Add(strings.TrimPrefix(e, "."))
	}
	return ret, nil
}

func ParseExtensionsFile(file string) ([]string, error) {
	var ret []string

	stream, err := os.Open(file)
	if err != nil {
		return ret, err
	}
	defer stream.Close()

	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		e := scanner.Text()
		e = strings.TrimSpace(e)
		// remove leading . from extensions
		ret = append(ret, strings.TrimPrefix(e, "."))
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ret, nil
}

// ParseCommaSeparatedInt parses the status codes provided as a comma separated list
func ParseCommaSeparatedInt(inputString string) (Set[int], error) {
	ret := NewSet[int]()

	if inputString == "" {
		return ret, nil
	}

	for _, part := range strings.Split(inputString, ",") {
		part = strings.TrimSpace(part)
		// check for range
		if strings.Contains(part, "-") {
			re := regexp.MustCompile(`^\s*(\d+)\s*-\s*(\d+)\s*$`)
			match := re.FindStringSubmatch(part)
			if match == nil || len(match) != 3 {
				return NewSet[int](), fmt.Errorf("invalid range given: %s", part)
			}
			from := strings.TrimSpace(match[1])
			to := strings.TrimSpace(match[2])
			fromI, err := strconv.Atoi(from)
			if err != nil {
				return NewSet[int](), fmt.Errorf("invalid string in range %s: %s", part, from)
			}
			toI, err := strconv.Atoi(to)
			if err != nil {
				return NewSet[int](), fmt.Errorf("invalid string in range %s: %s", part, to)
			}
			if toI < fromI {
				return NewSet[int](), fmt.Errorf("invalid range given: %s", part)
			}
			for i := fromI; i <= toI; i++ {
				ret.Add(i)
			}
		} else {
			i, err := strconv.Atoi(part)
			if err != nil {
				return NewSet[int](), fmt.Errorf("invalid string given: %s", part)
			}
			ret.Add(i)
		}
	}
	return ret, nil
}

// Windows reserved characters: < > : " | ? * and control characters (0-31)
var filenameInvalidChars = regexp.MustCompile(`[<>:"|?*\x00-\x1f]`)

// sanitizeFilename removes or replaces invalid characters from a filename
// to make it safe for use on Windows, macOS, and Linux filesystems
func SanitizeFilename(filename string) string {
	if filename == "" {
		return "unnamed"
	}

	// Remove leading/trailing whitespace
	filename = strings.TrimSpace(filename)

	// Replace path separators and other problematic characters
	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")

	filename = filenameInvalidChars.ReplaceAllString(filename, "_")

	// Remove non-printable Unicode characters
	filename = strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) {
			return r
		}
		return '_'
	}, filename)

	// Windows reserved names (case-insensitive)
	reservedNames := []string{
		"CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
	}

	// Check if filename (without extension) is a reserved name
	nameOnly := strings.TrimSuffix(filename, filepath.Ext(filename))
	for _, reserved := range reservedNames {
		if strings.EqualFold(nameOnly, reserved) {
			filename = "_" + filename
			break
		}
	}

	// Remove trailing dots and spaces (Windows requirement)
	filename = strings.TrimRight(filename, ". ")

	// Ensure filename isn't empty after sanitization
	if filename == "" {
		filename = "unnamed"
	}

	filename = filepath.Base(filename)

	// Limit length to 255 characters (common filesystem limit)
	if len(filename) > 255 {
		ext := filepath.Ext(filename)
		base := strings.TrimSuffix(filename, ext)
		maxBase := 255 - len(ext)
		if maxBase > 0 {
			filename = base[:maxBase] + ext
		} else {
			filename = filename[:255]
		}
	}

	return filename
}
