package syllable

import (
	"bytes"
	"encoding/json"
	"strings"

	prose "gopkg.in/antipasta/prose.v2"
)

type Haiku [][]prose.Token

func (h Haiku) MarshalJSON() ([]byte, error) {
	haikuStrArr := h.ToStringSlice()
	return json.Marshal(haikuStrArr)
}

func (h Haiku) ToStringArray() [3]string {
	haikuLines := [3]string{}
	for lineIndex, line := range h {
		haikuLine := bytes.Buffer{}
		for wordIndex := range line {
			if IsSymbolOrPunct(&line[wordIndex]) {
				continue
			}
			if wordIndex > 0 {
				haikuLine.WriteString(" ")
			}
			haikuLine.WriteString(line[wordIndex].Text)
			if wordIndex+1 < len(line) && IsSymbolOrPunct(&line[wordIndex+1]) {
				haikuLine.WriteString(line[wordIndex+1].Text)
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
