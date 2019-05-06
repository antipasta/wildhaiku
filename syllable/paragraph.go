/*Package syllable contains classes and functions for parsing text, getting counts of syllables, and returning text subdivided by specific syllable counts
 */
package syllable

import (
	"fmt"

	"github.com/pkg/errors"
	prose "gopkg.in/antipasta/prose.v2"
)

// Paragraph is a slice of Sentences, used to process an entire tweet looking for Haikus
type Paragraph []Sentence

// Subdivide attempts to look for haikus, starting from each Sentence in the paragraph slice, and proceeding til end of all further sentences in the slice.
func (p Paragraph) Subdivide(sylSizes ...int) []Haiku {

	haikuMap := map[string]Haiku{}
	haikus := []Haiku{}
	for i := range p {
		haiku := p[i:].toCombinedSentence().Subdivide(sylSizes...)
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
func (p Paragraph) toCombinedSentence() Sentence {
	combinedSentence := Sentence{}
	for i := range p {
		combinedSentence = append(combinedSentence, p[i]...)
	}
	return combinedSentence

}

// TotalSyllables returns count of total syllables in Paragraph
func (p Paragraph) TotalSyllables() int {
	total := 0
	for _, sentence := range p {
		total += sentence.TotalSyllables()
	}
	return total
}

// NewParagraph takes a string as input, runs PreProcess functions on it, and then converts it to a Paragraph(slice of Sentences)
func (c *CMUCorpus) NewParagraph(sentence string) (Paragraph, error) {
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
		return paragraph, err
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
				return paragraph, nil
			}
			// we didnt get enough for a potential hiaku, bail
			return Paragraph{}, errors.Wrapf(err, "Could not form haiku from given input")
		}
		paragraph = append(paragraph, sentenceObj)
	}
	return paragraph, nil

}
