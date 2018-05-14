package libgobuster

import (
	"fmt"
)

func WriteToFile(output string, s *State) {
	_, err := s.OutputFile.WriteString(output)
	if err != nil {
		panic(fmt.Sprintf("[!] Unable to write to file %s", s.OutputFileName))
	}
}
