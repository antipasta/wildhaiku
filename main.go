package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/ChimeraCoder/anaconda"
	"github.com/antipasta/tweetstream/haikudetector"
	"github.com/gomodule/oauth1/oauth"
)

var flagConfigPath string

type TweetStreamer struct {
}

type StreamerConfig struct {
	ConsumerKey    string
	ConsumerSecret string
	AccessToken    string
	AccessSecret   string
}

func LoadConfig(path string) (*StreamerConfig, error) {
	cfg := StreamerConfig{}
	cfgBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(cfgBytes, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, err
}

func init() {
	flag.StringVar(&flagConfigPath, "config", "config.json", "Path to config file")
}
func main() {
	/*
		flag.Parse()
		cmu, err := haikudetector.LoadCMUCorpus()
		if err != nil {
			panic(err)
		}
		syl := cmu.SentenceSyllables("I am a haiku, or at least I think I am maybe I am wrong I can potentially keep going too")
		test := syl.Subdivide(5, 7, 5)
		log.Printf("hey check it %+v", test)
		if err != nil {
			panic(err)
		}
	*/
	TwitterLoop()
	//log.Printf("hey %v", syl)
}

func TwitterLoop() {
	flag.Parse()
	cfg, err := LoadConfig(flagConfigPath)
	if err != nil {
		panic(err)
	}
	consumerKeys := oauth.Credentials{
		Token:  cfg.ConsumerKey,
		Secret: cfg.ConsumerSecret,
	}
	token := oauth.Credentials{
		Token:  cfg.AccessToken,
		Secret: cfg.AccessSecret,
	}
	client := oauth.Client{
		TemporaryCredentialRequestURI: "https://api.twitter.com/oauth/request_token",
		ResourceOwnerAuthorizationURI: "https://api.twitter.com/oauth/authorize",
		TokenRequestURI:               "https://api.twitter.com/oauth/access_token",
		Credentials:                   consumerKeys,
	}
	cmu, err := haikudetector.LoadCMUCorpus()
	if err != nil {
		panic(err)
	}
	syl := cmu.ParagraphSyllables("A summer river being crossed how pleasing with sandals in my hands!")
	if err != nil {
		panic(err)
	}
	log.Printf("hey %v", syl)

	httpClient := http.Client{}
	resp, err := client.Post(&httpClient, &token, "https://stream.twitter.com/1.1/statuses/filter.json", url.Values{"lang": []string{"en"}, "track": []string{"the", "be", "to", "of", "and", "in", "that", "have", "I", "it", "for", "not"}, "tweet_mode": []string{"extended"}})
	//resp, err := client.Post(&httpClient, &token, "https://stream.twitter.com/1.1/statuses/filter.json", url.Values{"lang": []string{"en"}, "track": []string{"haiku"}, "tweet_mode": []string{"extended"}})
	if err != nil {
		panic(err)
	}
	if resp.StatusCode != 200 {
		panic(resp.Status)
	}
	defer resp.Body.Close()
	buf := bufio.NewReader(resp.Body)
	for {
		line, err := buf.ReadBytes('\n')
		if err == io.EOF || len(line) == 0 {
			continue
		}
		if err != nil {
			panic(err)
		}
		t := anaconda.Tweet{}
		//log.Printf("Got %+v", string(line))
		err = json.Unmarshal(line, &t)
		if err != nil {
			continue
		}
		if t.FullText == "" || t.Lang != "en" {
			continue
		}
		if t.RetweetedStatus != nil {
			t = *t.RetweetedStatus
		}
		syl := cmu.SentenceSyllables(t.FullText)
		foundHaiku := syl.Subdivide(5, 7, 5)
		if len(foundHaiku) > 0 {
			log.Printf("[%s] https://twitter.com/%v/status/%v %v [%+v]", foundHaiku, t.User.ScreenName, t.IdStr, t.FullText, syl.Nouns())
		}
	}

}
