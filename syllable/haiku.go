package syllable

import (
	"bytes"
	"encoding/json"
	"strings"
)

type Haiku Paragraph

func (h Haiku) MarshalJSON() ([]byte, error) {
	haikuStrArr := h.ToStringArray()
	return json.Marshal(haikuStrArr)
}

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
func (h Haiku) ToStringSlice() []string {
	haikuArr := h.ToStringArray()
	return haikuArr[:]
}
func (h Haiku) String() string {
	return strings.Join(h.ToStringSlice(), "\n")
}

func (h Haiku) FinalWord() string {
	finalLine := h[len(h)-1]
	return finalLine[len(finalLine)-1].Word.Text
}
