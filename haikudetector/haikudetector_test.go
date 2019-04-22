package haikudetector

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
			Input:          "here is some bad text. it is not a haiku. haiku starting here, such a bold test for this app. would love if it worked",
			ExpectedOutput: "haiku starting here,\nsuch a bold test for this app.\nwould love if it worked\n",
			Corpus:         cmu,
		},
		{
			Input:          "@PokerStars What does this mean - \"The Beatles\" prize?  Is this a birthday joke on me, as I did not receive the ticket in my account?  Thanks anyway, was still a good 64th Birthday! @PokerStars ",
			ExpectedOutput: "Is this a birthday\njoke on me, as I did not\nreceive the ticket\n",
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
		t.Errorf("Found no haikus")
	}
	if foundHaikus[0].String() != eh.ExpectedOutput {
		t.Errorf("Output [%v] does not match expected [%+v]", foundHaikus[0].String(), eh.ExpectedOutput)
	}
	return foundHaikus
}
