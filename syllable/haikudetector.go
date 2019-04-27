package syllable

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
	Dict       map[string]int
}
type Word struct {
	Word      prose.Token
	Syllables int
}

func LoadCMUCorpus(path string) (*CMUCorpus, error) {
	c := CMUCorpus{Dict: map[string]int{},
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
		c.Dict[words[0]] = c.countFromPhenomes(words[1:])
	}

	return &c, nil
}

func (c *CMUCorpus) countFromPhenomes(phenomes []string) int {
	syllables := 0

	for _, phenome := range phenomes {
		for _, char := range phenome {
			if unicode.IsDigit(char) {
				syllables++
				break
			}
		}
	}

	return syllables

}
func (c *CMUCorpus) SyllableCount(word string) (int, error) {
	lowerWord := strings.ToLower(word)
	phenomes, exists := c.Dict[lowerWord]
	if !exists {
		return 0, errors.Errorf("Word not found %v", lowerWord)
	}
	return phenomes, nil
}

func (c *CMUCorpus) HasSyllableCount(word string) bool {
	lowerWord := strings.ToLower(word)
	phenomes, exists := c.Dict[lowerWord]
	return exists && phenomes > 0
}

func IsSymbolOrPunct(token *prose.Token) bool {
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
		if c.HasSyllableCount(tokens[0].Text) || IsSymbolOrPunct(&tokens[0]) {
			return tokens
		}
		tokens = tokens[1:]
	}
	return tokens
}

func (c *CMUCorpus) TrimTrailingUnknowns(tokens []prose.Token) []prose.Token {
	for len(tokens) > 0 {
		lastIndex := len(tokens) - 1
		if c.HasSyllableCount(tokens[lastIndex].Text) || IsSymbolOrPunct(&tokens[lastIndex]) {
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

func (c *CMUCorpus) ToSyllableSentence(sentence string, filters ...TokenFilterFunc) (Sentence, error) {
	syllableSentence := Sentence{}
	//sentenceDoc, err := prose.NewDocument(sentence)
	sentenceDoc, err := prose.NewDocument(sentence, prose.WithExtraction(false), prose.WithTagging(false), prose.WithTokenization(true))
	if err != nil {
		return nil, errors.Wrapf(err, "Error parsing new document %+v", sentence)
	}
	for _, v := range TokenizeFunc(sentenceDoc.Tokens).Filter(filters...) {
		count, err := c.SyllableCount(v.Text)
		if err != nil {
			if IsSymbolOrPunct(&v) {
				syllableSentence = append(syllableSentence, Word{Word: v, Syllables: 0})
				continue
			}
			return Sentence{}, errors.Errorf("Could not find count for [%+v]", v)
		}
		syllableSentence = append(syllableSentence, Word{Word: v, Syllables: count})
	}
	return syllableSentence, nil
}
