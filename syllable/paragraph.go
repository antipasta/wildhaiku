package syllable

import (
	"fmt"
	"log"

	prose "gopkg.in/antipasta/prose.v2"
)

type Paragraph []Sentence

func (p Paragraph) Subdivide(sylSizes ...int) []Haiku {

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
func (p Paragraph) ToCombinedSentence() Sentence {
	combinedSentence := Sentence{}
	for i := range p {
		combinedSentence = append(combinedSentence, p[i]...)
	}
	return combinedSentence

}

func (p Paragraph) TotalSyllables() int {
	total := 0
	for _, sentence := range p {
		total += sentence.TotalSyllables()
	}
	return total
}

func (c *CMUCorpus) NewParagraph(sentence string) Paragraph {
	for _, pFunc := range c.PreProcess {
		sentence = pFunc(sentence)
	}
	paragraph := Paragraph{}
	sentenceDoc, err := prose.NewDocument(sentence,
		prose.WithExtraction(false),
		prose.WithTagging(false),
		prose.WithTokenization(false),
		prose.UsingModel(nil))
	if err != nil {
		log.Fatal(err)
	}
	sentences := sentenceDoc.Sentences()
	for i, sentence := range sentences {
		tokenFilters := []TokenFilterFunc{}
		if i == 0 {
			tokenFilters = append(tokenFilters, c.TrimStartingUnknowns)
		}
		if i == len(sentences)-1 {
			tokenFilters = append(tokenFilters, c.TrimTrailingUnknowns)
		}
		sentenceObj, err := c.NewSentence(sentence.Text, tokenFilters...)
		if err != nil {
			// Got an error mid sentence after filtering, bail
			//log.Printf("Got error when parsing sentence syllables %v", err)
			if paragraph.TotalSyllables() >= 17 {
				// Without this sentence we have more than enough to attempt to find a haiku, return what we found so far
				return paragraph
			}
			// we didnt get enough for a potential hiaku, bail
			return Paragraph{}
		}
		paragraph = append(paragraph, sentenceObj)
	}
	return paragraph

}
