package syllable

type Sentence []Word

// TotalSyllables counts syllables of all Words in the sentence
func (s Sentence) TotalSyllables() int {
	total := 0
	for _, word := range s {
		total += word.Syllables
	}
	return total
}

// Subdivide splits a Sentence into a slice of Sentences, splitting on syllable boundaries that are specified in sylSizes, and returning a Haiku
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
				// append any symbols or punctuation
				haikuLine = append(haikuLine, s[wordIndex])
				continue
			}
			if curLineSize == sylSize {
				// next line of haiku
				break
			}
			if curLineSize+curSize > sylSize {
				// we're over, give up
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
