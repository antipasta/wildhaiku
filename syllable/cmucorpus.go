package syllable

import (
	"html"
	"io/ioutil"
	"strings"
	"unicode"

	"github.com/pkg/errors"
	prose "gopkg.in/antipasta/prose.v2"
)

// PreProcessFunc is used to transform strings before process
type PreProcessFunc func(string) string
type tokenizeFunc func() []prose.Token

// TokenFilterFunc is any function that takes a slice of Tokens and returns a slice of Tokens. Used for filtering out cruft, examples include TrimStartingUnknowns and TrimTrailingUnknowns
type TokenFilterFunc func([]prose.Token) []prose.Token

// CMUCorpus is used for looking up syllable counts after some preprocessing of string
type CMUCorpus struct {
	PreProcess []PreProcessFunc
	Dict       map[string]int
}

// Word contains a token and its corresponding syllable count
type Word struct {
	Word      prose.Token
	Syllables int
}

func (tf tokenizeFunc) filter(filterFuncs ...TokenFilterFunc) []prose.Token {
	tokens := tf()
	for _, filterFunc := range filterFuncs {
		tokens = filterFunc(tokens)
	}
	return tokens
}

// NewCMUCorpus Reads cmu corpus file off disk and converts it to a mapping of word to syllablecount, returning *CMUCorpus
func NewCMUCorpus(path string) (*CMUCorpus, error) {
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

// SyllableCount Returns syllable count of word, errors if word not found
func (c *CMUCorpus) SyllableCount(word string) (int, error) {
	lowerWord := strings.ToLower(word)
	phenomes, exists := c.Dict[lowerWord]
	if !exists {
		return 0, errors.Errorf("Word not found %v", lowerWord)
	}
	return phenomes, nil
}

// HasSyllableCount Checks if a word is in the cpu corpus and has a syllable count
func (c *CMUCorpus) HasSyllableCount(word string) bool {
	lowerWord := strings.ToLower(word)
	phenomes, exists := c.Dict[lowerWord]
	return exists && phenomes > 0
}

// IsSymbolOrPunct Checks if token is a non word or digit character
func IsSymbolOrPunct(token *prose.Token) bool {
	if len(token.Text) != 1 {
		return false
	}
	char := rune(token.Text[0])
	return len(token.Tag) == 1 || unicode.IsSymbol(char) || unicode.IsPunct(char)
}

// TrimStartingUnknowns Trims tokens that are not in cmu corpus from start of token slice, returning trimmed slice
func (c *CMUCorpus) TrimStartingUnknowns(tokens []prose.Token) []prose.Token {
	for len(tokens) > 0 {
		if c.HasSyllableCount(tokens[0].Text) || IsSymbolOrPunct(&tokens[0]) {
			return tokens
		}
		tokens = tokens[1:]
	}
	return tokens
}

// TrimTrailingUnknowns Trims tokens that are not in cmu corpus from end of token slice, returning trimmed slice
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

// NewSentence tokenizes a string, potentially performs filtering, looks up syllable counts, and then returns a Sentence, which is an array of []Words
func (c *CMUCorpus) NewSentence(sentence string, filters ...TokenFilterFunc) (Sentence, error) {
	syllableSentence := Sentence{}

	sentenceDoc, err := prose.NewDocument(sentence,
		prose.WithTokenization(true),
		prose.WithExtraction(false),
		prose.WithTagging(false),
		prose.UsingModel(nil))
	if err != nil {
		return nil, errors.Wrapf(err, "Error parsing new document %+v", sentence)
	}
	for _, v := range tokenizeFunc(sentenceDoc.Tokens).filter(filters...) {
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
