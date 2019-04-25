package haikudetector

import (
	"bytes"

	prose "gopkg.in/antipasta/prose.v2"
)

type Haiku [][]prose.Token

func (h Haiku) String() string {
	haiku := bytes.Buffer{}
	for _, line := range h {
		for wordIndex := range line {
			if len(line[wordIndex].Tag) == 1 {
				continue
			}
			if wordIndex > 0 {
				haiku.WriteString(" ")
			}
			haiku.WriteString(line[wordIndex].Text)
			if wordIndex+1 < len(line) && len(line[wordIndex+1].Tag) == 1 {
				haiku.WriteString(line[wordIndex+1].Text)
			}
		}
		haiku.WriteString("\n")
	}
	return haiku.String()
}
