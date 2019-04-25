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
	"github.com/pkg/errors"
)

var flagConfigPath string

type TweetStreamer struct {
	Config         *StreamerConfig
	ConsumerKeys   *oauth.Credentials
	Token          *oauth.Credentials
	Client         *oauth.Client
	httpClient     *http.Client
	ProcessChannel chan *anaconda.Tweet
	corpus         *haikudetector.CMUCorpus
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

func main() {
	flag.Parse()
	cfg, err := LoadConfig(flagConfigPath)
	if err != nil {
		panic(err)
	}
	ts := NewTweetStreamer(cfg)
	go ts.ProcessLoop()
	err = ts.StreamLoop()
	if err != nil {
		panic(err)
	}

}

func NewTweetStreamer(cfg *StreamerConfig) *TweetStreamer {
	channel := make(chan *anaconda.Tweet, 10000)
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
	cmu, err := haikudetector.LoadCMUCorpus("haikudetector/cmudict.dict")
	if err != nil {
		panic(err)
	}
	ts := TweetStreamer{Config: cfg, ConsumerKeys: &consumerKeys, Token: &token, Client: &client, httpClient: &http.Client{}, ProcessChannel: channel, corpus: cmu}
	return &ts
}

func (ts *TweetStreamer) Connect() (*http.Response, error) {
	resp, err := ts.Client.Post(ts.httpClient, ts.Token, "https://stream.twitter.com/1.1/statuses/filter.json", url.Values{"lang": []string{"en"}, "track": ts.Config.TrackingKeywords, "tweet_mode": []string{"extended"}})
	if err != nil {
		return nil, errors.Wrapf(err, "Caught error when connecting to twitter stream")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("Received non-OK error code [%v] [%v] when connecting to twitter stream", resp.StatusCode, resp.Status)

	}
	return resp, nil
}

func (ts *TweetStreamer) StreamLoop() error {
	resp, err := ts.Connect()
	if err != nil {
		return err
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
		ts.ProcessChannel <- &t
	}
}

func (ts *TweetStreamer) Process(t *anaconda.Tweet) []haikudetector.Haiku {
	paragraph := ts.corpus.ToSyllableParagraph(t.FullText)
	foundHaikus := paragraph.Subdivide(5, 7, 5)
	if len(foundHaikus) > 0 {
		log.Printf("https://twitter.com/%v/status/%v %v", t.User.ScreenName, t.IdStr, t.FullText)
		for i, foundHaiku := range foundHaikus {
			color.Cyan.Printf("%d. %s\n", i+1, foundHaiku)

		}
	}
	return foundHaikus
}
func (ts *TweetStreamer) ProcessLoop() error {
	for tweet := range ts.ProcessChannel {
		ts.Process(tweet)
	}
	return nil
}
