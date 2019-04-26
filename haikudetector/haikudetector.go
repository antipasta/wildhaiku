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
		PreProcess: []PreProcessFunc{html.UnescapeString},
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

func (c *CMUCorpus) HasSyllableCount(word string) bool {
	lowerWord := strings.ToLower(word)
	phenomes, exists := c.Dict[lowerWord]
	return exists && len(phenomes) > 0
}

func (c *CMUCorpus) IsSymbol(token *prose.Token) bool {
	if len(token.Text) != 1 {
		return false
	}
	char := rune(token.Text[0])
	return len(token.Tag) == 1 || unicode.IsSymbol(char) || unicode.IsPunct(char)
}

type TokenizeFunc func() []prose.Token
type TokenFilterFunc func([]prose.Token) []prose.Token

func (c *CMUCorpus) TrimStartingUnknowns(tokens []prose.Token) []prose.Token {
	for len(tokens) > 0 {
		if c.HasSyllableCount(tokens[0].Text) || c.IsSymbol(&tokens[0]) {
			return tokens
		}
		tokens = tokens[1:]
	}
	return tokens
}

func (c *CMUCorpus) TrimTrailingUnknowns(tokens []prose.Token) []prose.Token {
	for len(tokens) > 0 {
		lastIndex := len(tokens) - 1
		if c.HasSyllableCount(tokens[lastIndex].Text) || c.IsSymbol(&tokens[lastIndex]) {
			return tokens
		}
		tokens = tokens[0:lastIndex]
	}
	return tokens
}

func (tf TokenizeFunc) Filter(filterFuncs ...TokenFilterFunc) []prose.Token {
	tokens := tf()
	for _, filterFunc := range filterFuncs {
		tokens = filterFunc(tokens)
	}
	return tokens
}

func (c *CMUCorpus) ToSyllableSentence(sentence string) (SyllableSentence, error) {
	syllableSentence := SyllableSentence{}
	sentenceDoc, err := prose.NewDocument(sentence)
	if err != nil {
		return nil, errors.Wrapf(err, "Error parsing new document %+v", sentence)
	}
	for _, v := range TokenizeFunc(sentenceDoc.Tokens).Filter(c.TrimStartingUnknowns, c.TrimTrailingUnknowns) {
		if c.IsSymbol(&v) {
			syllableSentence = append(syllableSentence, SyllableWord{Word: v, Syllables: 0})
			continue
		}
		count, err := c.SyllableCount(v.Text)
		if err != nil {
			return SyllableSentence{}, errors.Errorf("Could not find count for [%+v]", v)
		}
		syllableSentence = append(syllableSentence, SyllableWord{Word: v, Syllables: count})
	}
	return syllableSentence, nil
}
