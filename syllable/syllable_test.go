package syllable

import (
	"testing"
)

type ExpectedHaiku struct {
	Input          string
	ExpectedOutput [3]string
	Corpus         *CMUCorpus
}

func TestParagraph(t *testing.T) {
	cmu, err := NewCMUCorpus("cmudict.dict")
	if err != nil {
		t.Errorf("Error loading cmu dictionary %+v", err)
	}
	p, err := cmu.NewParagraph("This is a haiku. Another sentence, cool test. Good time had by all")
	if err != nil {
		t.Errorf("Error creating paragraph %s", err)
	}
	if len(p) != 3 {
		t.Errorf("Paragraph should be of length 3")
	}
	if p.TotalSyllables() != 17 {
		t.Errorf("Paragraph syllables should be 17")
	}
	errorParagraph, err := cmu.NewParagraph("An unknown word, argblarg. Another sentence, cool test. Good time had by all")
	if err == nil {
		t.Errorf("Should have gotten an error when creating a paragraph with an unknown word in the middle")
	}
	if len(errorParagraph) > 0 {
		t.Errorf("Should have gotten a paragraph with length 0 since an unknown word is in the middle of the text")
	}
}

func TestSentence(t *testing.T) {
	cmu, err := NewCMUCorpus("cmudict.dict")
	if err != nil {
		t.Errorf("Error loading cmu dictionary %+v", err)
	}
	unknownToken, err := cmu.NewSentence("What a cool sentence. #wow")
	if err == nil {
		t.Errorf("Should get an error due to unknown trailing token")
	}
	if len(unknownToken) != 0 {
		t.Errorf("Should get an empty sentence due to unknown trailing token")
	}
	trimmed, err := cmu.NewSentence("@someone What a cool sentence. #wow", cmu.TrimStartingUnknowns, cmu.TrimTrailingUnknowns)
	if err != nil {
		t.Errorf("Should get no error since unknown tokens were trimmed, got %s", err)
	}
	if trimmed.TotalSyllables() != 5 {
		t.Errorf("Should get 5 syllables for trimmed sentence")
	}
}

func TestProblematicHaikus(t *testing.T) {
	cmu, err := NewCMUCorpus("cmudict.dict")
	if err != nil {
		t.Errorf("Error loading cmu dictionary %+v", err)
	}
	cases := []ExpectedHaiku{
		{
			Input:          "@startingjunk #hi testing a trim of both starting and trailing junk in the same sentence #devlyfe",
			ExpectedOutput: [3]string{"testing a trim of", "both starting and trailing junk", "in the same sentence"},
			Corpus:         cmu,
		},
		{
			Input:          "here is some bad text. it is not a haiku. haiku starting here, such a bold test for this app. would love if it worked",
			ExpectedOutput: [3]string{"haiku starting here,", "such a bold test for this app.", "would love if it worked"},
			Corpus:         cmu,
		},
		{
			Input:          "here is some bad text. it is not a haiku. haiku starting here, such a bold test for this app. would love if it worked #trailingjunk",
			ExpectedOutput: [3]string{"haiku starting here,", "such a bold test for this app.", "would love if it worked"},
			Corpus:         cmu,
		},
		{
			Input:          "@PokerStars What does this mean - \"The Beatles\" prize?  Is this a birthday joke on me, as I did not receive the ticket in my account?  Thanks anyway, was still a good 64th Birthday! @PokerStars ",
			ExpectedOutput: [3]string{"Is this a birthday", "joke on me, as I did not", "receive the ticket"},
			Corpus:         cmu,
		},
		{
			Input:          "Bill Barr is the Honey Badger. Honey Badger ain't scared of nothing. Broad shoulders, loose skin. Chuck Schumer? Honey Badger don't care. Gerry Nadler? Honey Badger don't care. Nancy Pelosi? Honey Badger don't care.",
			ExpectedOutput: [3]string{"Honey Badger ain't", "scared of nothing. Broad shoulders,", "loose skin. Chuck Schumer?"},
			Corpus:         cmu,
		},
		{
			Input:          "@LindaBr67589020 @AdamParkhomenko @marcorubio He said that he had no knowledge &amp; yet he now claims that they have known about it since May of last year!  I'm not a mathematician but doesn't November follow May?  He made Nelson look crazy to help Rick Scott win.  He helped Putin &amp; hurt our national security for a party win!  ",
			ExpectedOutput: [3]string{"He said that he had", "no knowledge & yet he now", "claims that they have known"},
			Corpus:         cmu,
		},
		{
			Input:          "How many cans of tuna are ok to eat at once? They’re so small...",
			ExpectedOutput: [3]string{"How many cans of", "tuna are ok to eat", "at once? They're so small..."},
			Corpus:         cmu,
		},
		{
			Input:          "Punctuation test; Hope & test that it works right??? Only time, will tell!!!",
			ExpectedOutput: [3]string{"Punctuation test;", "Hope & test that it works right???", "Only time, will tell!!!"},
			Corpus:         cmu,
		},
	}
	for _, h := range cases {
		//log.Printf("Testing %+v", h)
		h.HaikuTest(t)
	}

}

func (eh *ExpectedHaiku) HaikuTest(t *testing.T) []Haiku {
	paragraph, err := eh.Corpus.NewParagraph(eh.Input)
	if err != nil {
		t.Errorf("Got error %v on input %v", err, eh.Input)
	}
	foundHaikus := paragraph.Subdivide(5, 7, 5)
	if len(foundHaikus) == 0 {
		t.Errorf("Found no haikus for text %v", eh.Input)
	} else if foundHaikus[0].ToStringArray() != eh.ExpectedOutput {
		t.Errorf("Output [%v] does not match expected [%+v]", foundHaikus[0].ToStringArray(), eh.ExpectedOutput)
	}
	return foundHaikus
}
