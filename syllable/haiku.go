package syllable

import (
	"bytes"
	"encoding/json"
	"strings"
)

type Haiku Paragraph

func (h Haiku) MarshalJSON() ([]byte, error) {
	haikuStrArr := h.ToStringSlice()
	return json.Marshal(haikuStrArr)
}

func (h Haiku) ToStringArray() [3]string {
	haikuLines := [3]string{}
	for lineIndex, line := range h {
		haikuLine := bytes.Buffer{}
		for wordIndex := range line {
			if line[wordIndex].Syllables == 0 && IsSymbolOrPunct(&line[wordIndex].Word) {
				continue
			}
			if wordIndex > 0 && haikuLine.Len() > 0 {
				haikuLine.WriteString(" ")
			}
			haikuLine.WriteString(line[wordIndex].Word.Text)

			if wordIndex+1 < len(line) && line[wordIndex+1].Syllables == 0 && IsSymbolOrPunct(&line[wordIndex+1].Word) {
				haikuLine.WriteString(line[wordIndex+1].Word.Text)
			}
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
