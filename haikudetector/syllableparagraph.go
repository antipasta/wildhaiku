package haikudetector

import (
	"fmt"
	"log"

	prose "gopkg.in/antipasta/prose.v2"
)

type SyllableParagraph []SyllableSentence

func (p SyllableParagraph) Subdivide(sylSizes ...int) []Haiku {

	haikuMap := map[string]Haiku{}
	haikus := []Haiku{}
	for i := range p {
		haiku := p[i:].ToCombinedSentence().Subdivide(sylSizes...)
		haikuStr := fmt.Sprintf("%s", haiku)
		if len(haiku) > 0 {
			if _, exists := haikuMap[haikuStr]; !exists {
				haikuMap[haikuStr] = haiku
				haikus = append(haikus, haiku)
			}
		}
	}
	return haikus
}
func (p SyllableParagraph) ToCombinedSentence() SyllableSentence {
	combinedSentence := SyllableSentence{}
	for _, sentence := range p {
		combinedSentence = append(combinedSentence, sentence...)
	}
	return combinedSentence

}

func (c *CMUCorpus) ToSyllableParagraph(sentence string) SyllableParagraph {
	for _, pFunc := range c.PreProcess {
		sentence = pFunc(sentence)
	}
	paragraph := SyllableParagraph{}
	sentenceDoc, err := prose.NewDocument(sentence)
	if err != nil {
		log.Fatal(err)
	}
	for _, sentence := range sentenceDoc.Sentences() {
		sentenceObj, err := c.ToSyllableSentence(sentence.Text)
		if err != nil {
			//log.Printf("Got error when parsing sentence syllables %v", err)
			//return SyllableParagraph{}
			continue
		}
		paragraph = append(paragraph, sentenceObj)
	}
	return paragraph

}
