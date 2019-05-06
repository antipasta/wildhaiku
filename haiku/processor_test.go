package haiku

import (
	"testing"

	"github.com/antipasta/wildhaiku/syllable"
	"github.com/antipasta/wildhaiku/twitter"
)

func TestProcess(t *testing.T) {
	cmu, err := syllable.NewCMUCorpus("../syllable/cmudict.dict")
	if err != nil {
		t.Errorf("Error loading cmu dictionary %+v", err)
	}
	p := &Processor{corpus: cmu}
	tweet := twitter.Tweet{Text: "no haikus here"}
	output := p.process(&tweet)
	if output == nil {
		t.Errorf("Expected to get a value back")
	}
	if len(output.Haikus) > 0 {
		t.Errorf("Expected to find no haikus for tweet %+v, got %+v", tweet, output.Haikus)
	}
	haikuTweet := twitter.Tweet{Text: "this is a haiku. hope the test finds it alright, i think that it should."}
	output = p.process(&haikuTweet)
	if output == nil {
		t.Errorf("Expected to get a value back")
	}
	if len(output.Haikus) != 1 {
		t.Errorf("Expected to find 1 haiku for tweet %+v, got %+v", tweet, output.Haikus)
	}
	expected := [3]string{"this is a haiku.", "hope the test finds it alright,", "i think that it should."}
	if output.Haikus[0].ToStringArray() != expected {
		t.Errorf("Haikus %+v did not match expected %+v", output.Haikus[0], expected)
	}
}
