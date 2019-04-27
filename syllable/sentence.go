package syllable

import (
	"strings"

	prose "gopkg.in/antipasta/prose.v2"
)

type Sentence []SyllableWord

func (s Sentence) TotalSyllables() int {
	total := 0
	for _, word := range s {
		total += word.Syllables
	}
	return total
}

func (s Sentence) Nouns() []string {
	nouns := []string{}
	for _, word := range s {
		if strings.HasPrefix(word.Word.Tag, "N") {
			nouns = append(nouns, word.Word.Text)
		}
	}
	return nouns
}
func (s Sentence) Subdivide(sylSizes ...int) Haiku {
	curSentence := [][]prose.Token{}
	wordIndex := 0
	if s.TotalSyllables() < 17 {
		return curSentence
	}
	for _, sylSize := range sylSizes {
		curLineSize := 0
		haikuLine := []prose.Token{}
		for ; wordIndex < len(s); wordIndex++ {
			curWord := s[wordIndex]
			curSize := curWord.Syllables
			if curWord.Syllables == 0 {
				haikuLine = append(haikuLine, s[wordIndex].Word)
				continue
			}
			if curLineSize+curSize > sylSize {
				if curLineSize == sylSize {
					break
				}
				return [][]prose.Token{}
			}
			curLineSize += curSize
			haikuLine = append(haikuLine, s[wordIndex].Word)
		}
		if curLineSize == sylSize {
			curSentence = append(curSentence, haikuLine)
		}

	}
	return curSentence
}
