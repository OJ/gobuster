package libgobuster

func WriteToFile(output string, s *State) {
	_, err := s.OutputFile.WriteString(output)
	if err != nil {
		panic("[!] Unable to write to file " + s.OutputFileName)
	}
}
