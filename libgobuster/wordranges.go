package libgobuster

import (
	"fmt"
	"regexp"
	"strconv"
)

type RangeStringToken struct {
	Text       string
	IsRange    bool
	RangeStart int
	RangeEnd   int
}

func ParseTokens(word string) []RangeStringToken {
	var tokens []RangeStringToken

	pattern, _ := regexp.Compile(`\[(\d+)\-(\d+)\]|\.|([\w\-_\/]+)`)
	matches := pattern.FindAllStringSubmatch(word, -1)

	for _, match := range matches {
		if match[1] == "" {
			text := string(match[0])
			tokens = append(tokens, RangeStringToken{Text: text, IsRange: false})
		} else {
			start, _ := strconv.Atoi(string(match[1]))
			end, _ := strconv.Atoi(string(match[2]))

			tokens = append(
				tokens, RangeStringToken{
					IsRange:    true,
					RangeStart: start,
					RangeEnd:   end})
		}
	}

	return tokens
}

func ExpandWords(tokens []RangeStringToken) []string {
	var expandedWords []string

	for i, _ := range tokens {
		var newWords []string
		if tokens[i].IsRange {
			if len(expandedWords) < 1 {
				for j := tokens[i].RangeStart; j <= tokens[i].RangeEnd; j++ {
					url := fmt.Sprintf("%s%d", "/", j)
					newWords = append(newWords, url)
				}
			} else {
				for _, expandedWord := range expandedWords {

					for j := tokens[i].RangeStart; j <= tokens[i].RangeEnd; j++ {
						url := fmt.Sprintf("%s%d", expandedWord, j)
						newWords = append(newWords, url)
					}
				}

			}
		} else {
			if len(expandedWords) < 1 {
				newWords = append(newWords, tokens[i].Text)
			} else {
				for _, expandedWord := range expandedWords {
					newWords = append(newWords, expandedWord+tokens[i].Text)
				}
			}

		}
		expandedWords = newWords
	}
	return expandedWords
}
