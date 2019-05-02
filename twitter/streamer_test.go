package twitter

import (
	"io"
	"os"
	"testing"

	"github.com/antipasta/wildhaiku/config"
)

func TestStreamer(t *testing.T) {
	testCfg := config.Streamer{}
	s := NewStreamer(&testCfg)
	tweetFile, err := os.Open("sampletweets.json")
	if err != nil {
		t.Errorf("Error opening sampletweets.json %v", err)
	}
	err = s.StreamLoop(tweetFile)
	if err != nil && err != io.EOF {
		t.Errorf("Error when streaming %v", err)
	}
	if len(s.ProcessChannel) != 591 {
		// 108 tweets  in file are non english or have no text in tweet body
		t.Errorf("expected 591 tweets in channel, got %v", len(s.ProcessChannel))
	}
	for len(s.ProcessChannel) > 0 {
		tweet := <-s.ProcessChannel
		if tweet == nil {
			t.Errorf("Got nil tweet on channel")
		}
		if tweet.Lang != "en" {
			t.Errorf("Got non english language tweet %s", tweet.Lang)
		}
		if tweet.FullText() == "" {
			t.Errorf("Got tweet with empty tweet body")
		}
	}

}
