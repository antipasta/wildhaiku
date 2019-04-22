package haikudetector

import (
	"log"
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
	h := ExpectedHaiku{
		Input:          "here is some bad text. it is not a haiku. haiku starting here, such a bold test for this app. would love if it worked",
		ExpectedOutput: "haiku starting here,\nsuch a bold test for this app.\nwould love if it worked\n",
		Corpus:         cmu,
	}
	found := h.HaikuTest()
	if len(found) == 0 {
		t.Errorf("Found no haikus")
	}
	if found[0].String() != h.ExpectedOutput {
		t.Errorf("Output [%v] does not match expected [%+v]", found[0].String(), h.ExpectedOutput)
	}

}

func (eh *ExpectedHaiku) HaikuTest() []Haiku {
	paragraph := eh.Corpus.ToSyllableParagraph(eh.Input)
	foundHaikus := paragraph.Subdivide(5, 7, 5)
	if len(foundHaikus) > 0 {
		for _, haiku := range foundHaikus {
			log.Printf("Found haiku %+v", haiku)
		}
	}
	return foundHaikus
}
