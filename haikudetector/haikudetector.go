package haikudetector

import (
	"html"
	"io/ioutil"
	"strings"
	"unicode"

	"github.com/pkg/errors"
	prose "gopkg.in/antipasta/prose.v2"
)

type PreProcessFunc func(string) string

func TrimUnknowns(input string) string {
	return input
}

type CMUCorpus struct {
	PreProcess []PreProcessFunc
	Dict       map[string][]string
}
type SyllableWord struct {
	Word      prose.Token
	Syllables int
}

func LoadCMUCorpus(path string) (*CMUCorpus, error) {
	c := CMUCorpus{Dict: map[string][]string{},
		PreProcess: []PreProcessFunc{html.UnescapeString, TrimUnknowns},
	}
	cmuBytes, err := ioutil.ReadFile(path)
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
		return 0, errors.Errorf("Word not found %v", lowerWord)
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

func (c *CMUCorpus) IsSymbol(word string) bool {
	return len(word) == 1 && unicode.IsSymbol(rune(word[0]))
}

func (c *CMUCorpus) SentenceSyllables(sentence string) (SyllableSentence, error) {
	syllableSentence := SyllableSentence{}
	sentenceDoc, err := prose.NewDocument(sentence)
	if err != nil {
		return nil, errors.Wrapf(err, "Error parsing new document %+v", sentence)
	}
	var total int
	for _, v := range sentenceDoc.Tokens() {
		if len(v.Tag) == 1 || c.IsSymbol(v.Text) {
			// symbol
			syllableSentence = append(syllableSentence, SyllableWord{Word: v, Syllables: 0})
			continue
		}
		count, err := c.SyllableCount(v.Text)
		if err != nil {
			if total == 0 {
				// Skips unknown words at beginning of sentence. TODO should be beginning (and end) of paragraph
				continue
			}
			return SyllableSentence{}, errors.Errorf("Could not find count for [%+v]", v)
		}
		total += count
		syllableSentence = append(syllableSentence, SyllableWord{Word: v, Syllables: count})
	}
	return syllableSentence, nil
}
