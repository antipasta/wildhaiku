package haikudetector

import (
	"bytes"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"strings"
	"unicode"

	prose "gopkg.in/jdkato/prose.v2"
)

type CMUCorpus struct {
	Dict map[string][]string
}

func LoadCMUCorpus() (*CMUCorpus, error) {
	c := CMUCorpus{Dict: map[string][]string{}}
	cmuBytes, err := ioutil.ReadFile("haikudetector/cmudict.dict")
	if err != nil {
		return nil, err
	}
	cmuStr := string(cmuBytes)
	cmuLines := strings.Split(cmuStr, "\n")
	for _, line := range cmuLines {
		words := strings.Split(line, " ")
		c.Dict[words[0]] = words[1:]
	}

	return &c, nil
}

func (c *CMUCorpus) SyllableCount(word string) (int, error) {
	syllables := 0
	lowerWord := strings.ToLower(word)
	phenomes, exists := c.Dict[lowerWord]
	if !exists || len(phenomes) == 0 {
		return 0, fmt.Errorf("Word not found %v", lowerWord)
	}

	for _, phenome := range phenomes {
		for _, char := range phenome {
			if unicode.IsDigit(char) {
				syllables++
				break
			}
		}
	}

	return syllables, nil

}

type SyllableWord struct {
	Word      prose.Token
	Syllables int
}

type SyllableParagraph []SyllableSentence
type SyllableSentence []SyllableWord

func (s SyllableSentence) TotalSyllables() int {
	total := 0
	for _, word := range s {
		total += word.Syllables
	}
	return total
}

func (s SyllableSentence) Nouns() []string {
	nouns := []string{}
	for _, word := range s {
		if strings.HasPrefix(word.Word.Tag, "N") {
			nouns = append(nouns, word.Word.Text)
		}
	}
	return nouns
}

type Haiku [][]prose.Token

func (h Haiku) String() string {
	haiku := bytes.Buffer{}
	for _, line := range h {
		for wordIndex := range line {
			if len(line[wordIndex].Tag) == 1 {
				continue
			}
			if haiku.Len() > 0 {
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

func (s SyllableSentence) Subdivide(sylSizes ...int) Haiku {
	curSentence := [][]prose.Token{}
	wordIndex := 0
	if s.TotalSyllables() < 17 {
		return curSentence
	}
	for _, sylSize := range sylSizes {
		haikuLine := []prose.Token{}
		for ; wordIndex <= len(s); wordIndex++ {
			sylSize -= s[wordIndex].Syllables
			//log.Printf("Word is %v, sylSize is %v, word syllables is %v", s[wordIndex].Word, sylSize, s[wordIndex].Syllables)
			if sylSize < 0 {
				// not haiku material
				return [][]prose.Token{}
			}
			haikuLine = append(haikuLine, s[wordIndex].Word)
			if sylSize == 0 {
				curSentence = append(curSentence, haikuLine)
				wordIndex++
				//log.Printf("curSentence is now %+v", curSentence)
				break
			}
		}

	}
	return curSentence
}

func (p SyllableParagraph) ToCombinedSentence() SyllableSentence {
	combinedSentence := SyllableSentence{}
	for _, sentence := range p {
		combinedSentence = append(combinedSentence, sentence...)
	}
	return combinedSentence

}

func (c *CMUCorpus) ToSyllableParagraph(sentence string) SyllableParagraph {
	paragraph := SyllableParagraph{}
	sentence = html.UnescapeString(sentence)
	sentenceDoc, err := prose.NewDocument(sentence)
	if err != nil {
		log.Fatal(err)
	}
	for _, sentence := range sentenceDoc.Sentences() {
		paragraph = append(paragraph, c.SentenceSyllables(sentence.Text))
	}
	return paragraph

}

func (c *CMUCorpus) SentenceSyllables(sentence string) SyllableSentence {
	syllableSentence := SyllableSentence{}
	sentenceDoc, err := prose.NewDocument(sentence)
	if err != nil {
		log.Fatal(err)
	}
	var total int
	for _, v := range sentenceDoc.Tokens() {
		if len(v.Tag) == 1 {
			// symbol
			syllableSentence = append(syllableSentence, SyllableWord{Word: v, Syllables: 0})
			continue
		}
		count, err := c.SyllableCount(v.Text)
		if err != nil {
			//TODO do we just give up here? for now continuing on but thats incorrect
			//log.Printf("Got error on syllable count for word %+v [%+v] %+v", v.Tag, v.Text, err)
			//return SyllableSentence{}
			break
		}
		total += count
		syllableSentence = append(syllableSentence, SyllableWord{Word: v, Syllables: count})
	}
	//log.Printf("Total for [%+v]: %v", sentence, total)
	return syllableSentence
}
