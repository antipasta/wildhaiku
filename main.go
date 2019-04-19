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
	"github.com/antipasta/wildhaiku/haikudetector"
	"github.com/gomodule/oauth1/oauth"
	"github.com/gookit/color"
)

var flagConfigPath string

type TweetStreamer struct {
}

type StreamerConfig struct {
	ConsumerKey      string
	ConsumerSecret   string
	AccessToken      string
	AccessSecret     string
	TrackingKeywords []string
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
func multiTest() {
	cmu, err := haikudetector.LoadCMUCorpus()
	if err != nil {
		panic(err)
	}
	//text := "here is some bad text. it is not a haiku. haiku starting here, such a bold test for this app. would love if it worked"
	//text := "here's one haiku. it is an okay haiku. another is here, such a bold test for this app. would love if it worked"
	//deduped := "@PokerStars What does this mean - \"The Beatles\" prize?  Is this a birthday joke on me, as I did not receive the ticket in my account?  Thanks anyway, was still a good 64th Birthday! @PokerStars "
	//bustedencoding := "In a relationship communication is key. Getting mad at your girl bc she's expressing what upsets her is lame. Instead of flipping the situation &amp; blaming it on her, ask her where u went wrong &amp; how u can fix it. Even if it's not that deep to u, it could be a serious matter to her"
	bustedSkippedWords := "Please don't cut back ivy too hard in your garden! At this time of year, it's fantastic nesting habitat for birds. In autumn, its flowers will be a late-season food supply for pollinators like bees, butterflies, hoverflies. Then winter berries for birds 🐝"

	paragraph := cmu.ToSyllableParagraph(bustedSkippedWords)
	foundHaikus := paragraph.Subdivide(5, 7, 5)
	if len(foundHaikus) > 0 {
		for _, haiku := range foundHaikus {
			log.Printf("Found haiku %+v", haiku)
		}
	}

}
func main() {
	flag.Parse()
	//multiTest()
	TwitterLoop()
}

func TwitterLoop() {
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

	httpClient := http.Client{}
	resp, err := client.Post(&httpClient, &token, "https://stream.twitter.com/1.1/statuses/filter.json", url.Values{"lang": []string{"en"}, "track": cfg.TrackingKeywords, "tweet_mode": []string{"extended"}})
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
		paragraph := cmu.ToSyllableParagraph(t.FullText)
		foundHaikus := paragraph.Subdivide(5, 7, 5)
		if len(foundHaikus) > 0 {
			log.Printf("https://twitter.com/%v/status/%v %v", t.User.ScreenName, t.IdStr, t.FullText)
			for i, foundHaiku := range foundHaikus {
				color.Cyan.Printf("%d. %s\n", i+1, foundHaiku)

			}
		}
	}

}
