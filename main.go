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
	flag.Parse()
	cfg, err := LoadConfig(flagConfigPath)
	if err != nil {
		panic(err)
	}
	consumerKeys := oauth.Credentials{
		Token:  cfg.ConsumerKey,
		Secret: cfg.ConsumerSecret,
	}
	log.Printf("consumer is %+v", consumerKeys)
	client := oauth.Client{
		TemporaryCredentialRequestURI: "https://api.twitter.com/oauth/request_token",
		ResourceOwnerAuthorizationURI: "https://api.twitter.com/oauth/authorize",
		TokenRequestURI:               "https://api.twitter.com/oauth/access_token",
		Credentials: oauth.Credentials{
			Token:  cfg.AccessToken,
			Secret: cfg.AccessSecret,
		}}

	httpClient := http.Client{}
	resp, err := client.Get(&httpClient, &consumerKeys, "https://stream.twitter.com/1.1/statuses/sample.json", url.Values{})
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
		log.Printf("Got %+v", string(line))
		err = json.Unmarshal(line, &t)
		if err != nil {
			continue
		}
		syllables := haikudetector.ParagraphSyllables(t.FullText)
		log.Printf("%+v %v: %v", syllables, t.User.ScreenName, t.FullText)
	}

}
