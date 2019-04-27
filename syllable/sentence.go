package syllable

import (
	"strings"
)

type Sentence []Word

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
	curSentence := Haiku{}
	wordIndex := 0
	if s.TotalSyllables() < 17 {
		return curSentence
	}
	for _, sylSize := range sylSizes {
		curLineSize := 0
		haikuLine := Sentence{}
		for ; wordIndex < len(s); wordIndex++ {
			curWord := s[wordIndex]
			curSize := curWord.Syllables
			if curWord.Syllables == 0 {
				haikuLine = append(haikuLine, s[wordIndex])
				continue
			}
			if curLineSize+curSize > sylSize {
				if curLineSize == sylSize {
					break
				}
				return Haiku{}
			}
			curLineSize += curSize
			haikuLine = append(haikuLine, s[wordIndex])
		}
		if curLineSize == sylSize {
			curSentence = append(curSentence, haikuLine)
		}

	}
	return curSentence
}
