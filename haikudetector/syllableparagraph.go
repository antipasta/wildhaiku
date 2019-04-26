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
	tokenFilters := []TokenFilterFunc{}
	sentences := sentenceDoc.Sentences()
	for i, sentence := range sentences {
		if i == 0 {
			tokenFilters = append(tokenFilters, c.TrimStartingUnknowns)
		}
		if i == len(sentences)-1 {
			tokenFilters = append(tokenFilters, c.TrimTrailingUnknowns)
		}
		sentenceObj, err := c.ToSyllableSentence(sentence.Text, tokenFilters...)
		if err != nil {
			// Got an error mid sentence after filtering, bail
			//log.Printf("Got error when parsing sentence syllables %v", err)
			if paragraph.ToCombinedSentence().TotalSyllables() >= 17 {
				// Without this sentence we have more than enough to attempt to find a haiku, return what we found so far
				return paragraph
			}
			// we didnt get enough for a potential hiaku, bail
			return SyllableParagraph{}
		}
		paragraph = append(paragraph, sentenceObj)
	}
	return paragraph

}
