package syllable

import (
	"testing"
)

type ExpectedHaiku struct {
	Input          string
	ExpectedOutput string
	Corpus         *CMUCorpus
}

func TestProblematicHaikus(t *testing.T) {
	cmu, err := LoadCMUCorpus("cmudict.dict")
	if err != nil {
		t.Errorf("Error loading cmu dictionary %+v", err)
	}
	cases := []ExpectedHaiku{
		{
			Input:          "@startingjunk #hi testing a trim of both starting and trailing junk in the same sentence #devlyfe",
			ExpectedOutput: "testing a trim of\nboth starting and trailing junk\nin the same sentence\n",
			Corpus:         cmu,
		},
		{
			Input:          "here is some bad text. it is not a haiku. haiku starting here, such a bold test for this app. would love if it worked",
			ExpectedOutput: "haiku starting here,\nsuch a bold test for this app.\nwould love if it worked\n",
			Corpus:         cmu,
		},
		{
			Input:          "here is some bad text. it is not a haiku. haiku starting here, such a bold test for this app. would love if it worked #trailingjunk",
			ExpectedOutput: "haiku starting here,\nsuch a bold test for this app.\nwould love if it worked\n",
			Corpus:         cmu,
		},
		{
			Input:          "@PokerStars What does this mean - \"The Beatles\" prize?  Is this a birthday joke on me, as I did not receive the ticket in my account?  Thanks anyway, was still a good 64th Birthday! @PokerStars ",
			ExpectedOutput: "Is this a birthday\njoke on me, as I did not\nreceive the ticket\n",
			Corpus:         cmu,
		},
		{
			Input:          "Bill Barr is the Honey Badger. Honey Badger ain't scared of nothing. Broad shoulders, loose skin. Chuck Schumer? Honey Badger don't care. Gerry Nadler? Honey Badger don't care. Nancy Pelosi? Honey Badger don't care.",
			ExpectedOutput: "Honey Badger ain't\nscared of nothing. Broad shoulders,\nloose skin. Chuck Schumer?\n",
			Corpus:         cmu,
		},
	}
	for _, h := range cases {
		h.HaikuTest(t)
	}

}

func (eh *ExpectedHaiku) HaikuTest(t *testing.T) []Haiku {
	paragraph := eh.Corpus.ToSyllableParagraph(eh.Input)
	foundHaikus := paragraph.Subdivide(5, 7, 5)
	if len(foundHaikus) == 0 {
		t.Errorf("Found no haikus for text %v", eh.Input)
	} else if foundHaikus[0].String() != eh.ExpectedOutput {
		t.Errorf("Output [%v] does not match expected [%+v]", foundHaikus[0].String(), eh.ExpectedOutput)
	}
	return foundHaikus
}
