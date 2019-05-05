package syllable

import (
	"bytes"
	"encoding/json"
	"strings"
)

// Haiku is a Paragraph(array of sentences) with some additional functionality for formatting Haikus for output
type Haiku Paragraph

// MarshalJSON satisfies the Marshaler interface, to JSONify a Haiku using ToStringArray()
func (h Haiku) MarshalJSON() ([]byte, error) {
	haikuStrArr := h.ToStringArray()
	return json.Marshal(haikuStrArr)
}

// ToStringArray returns a string output of the haiku, where each item in the array[3] is a line of the haiku
func (h Haiku) ToStringArray() [3]string {
	haikuLines := [3]string{}
	for lineIndex, line := range h {
		haikuLine := bytes.Buffer{}
		for wordIndex := range line {
			if line[wordIndex].Syllables == 0 && IsSymbolOrPunct(&line[wordIndex].Word) {
				haikuLine.WriteString(line[wordIndex].Word.Text)
				continue
			}
			if wordIndex > 0 && wordIndex < len(line) {
				// Works in most cases, will need refactor for proper spacing for quotes
				haikuLine.WriteString(" ")
			}
			haikuLine.WriteString(line[wordIndex].Word.Text)

		}
		haikuLines[lineIndex] = haikuLine.String()
	}
	return haikuLines
}
func (h Haiku) toStringSlice() []string {
	haikuArr := h.ToStringArray()
	return haikuArr[:]
}

// String stringifies the haiku for output
func (h Haiku) String() string {
	return strings.Join(h.toStringSlice(), "\n")
}

// FinalWord returns the final word of the haiku. Used for filtering haikus that end in a 'dangling' word
func (h Haiku) FinalWord() string {
	finalLine := h[len(h)-1]
	return finalLine[len(finalLine)-1].Word.Text
}
