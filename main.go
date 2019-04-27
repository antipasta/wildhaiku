package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/antipasta/wildhaiku/syllable"
	"github.com/gomodule/oauth1/oauth"
	"github.com/gookit/color"
	"github.com/pkg/errors"
)

var flagConfigPath string

type HaikuOutput struct {
	Haikus []syllable.Haiku
	Tweet  *Tweet
}

type TweetStreamer struct {
	Config         *StreamerConfig
	ConsumerKeys   *oauth.Credentials
	Token          *oauth.Credentials
	Client         *oauth.Client
	httpClient     *http.Client
	ProcessChannel chan *Tweet
	corpus         *syllable.CMUCorpus
	OutputChannel  chan *HaikuOutput
	OutFile        *os.File
}

type StreamerConfig struct {
	ConsumerKey      string
	ConsumerSecret   string
	AccessToken      string
	AccessSecret     string
	TrackingKeywords []string
	CorpusPath       string
	OutputPath       string
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
	go ts.OutputLoop()
	err = ts.StreamLoop()
	if err != nil {
		panic(err)
	}

}

func NewTweetStreamer(cfg *StreamerConfig) *TweetStreamer {
	processChannel := make(chan *Tweet, 10000)
	outChannel := make(chan *HaikuOutput, 10000)
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
	cmu, err := syllable.LoadCMUCorpus(cfg.CorpusPath)
	if err != nil {
		panic(err)
	}
	ts := TweetStreamer{Config: cfg, ConsumerKeys: &consumerKeys, Token: &token, Client: &client, httpClient: &http.Client{}, ProcessChannel: processChannel, OutputChannel: outChannel, corpus: cmu}
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
		t := Tweet{}
		err = json.Unmarshal(line, &t)
		if err != nil {
			log.Printf("Error parsing %+v", line)
			continue
		}
		if t.RetweetedStatus != nil {
			t = *t.RetweetedStatus
		}
		if t.FullText() == "" || t.Lang != "en" {
			continue
		}
		ts.ProcessChannel <- &t
	}
}

type Tweet struct {
	IDStr string `json:"id_str"`
	Lang  string `json:"lang"`
	User  struct {
		ScreenName string `json:"screen_name"`
	} `json:"user"`
	Text            string         `json:"text,omitempty"`
	ExtendedTweet   *ExtendedTweet `json:"extended_tweet,omitempty"`
	RetweetedStatus *Tweet         `json:"retweeted_status,omitempty"`
}

type ExtendedTweet struct {
	FullText string `json:"full_text,omitempty"`
}

func (t *Tweet) FullText() string {
	if t.ExtendedTweet != nil && len(t.ExtendedTweet.FullText) > 0 {
		return t.ExtendedTweet.FullText
	}
	return t.Text
}

func (ts *TweetStreamer) Process(t *Tweet) *HaikuOutput {
	paragraph := ts.corpus.ToSyllableParagraph(t.FullText())
	foundHaikus := paragraph.Subdivide(5, 7, 5)
	haikuStrings := [][]string{}
	for _, haiku := range foundHaikus {
		haikuStrings = append(haikuStrings, haiku.ToStringArray())
	}
	return &HaikuOutput{Tweet: t, Haikus: foundHaikus}
}

func (ts *TweetStreamer) ProcessLoop() error {
	for tweet := range ts.ProcessChannel {
		output := ts.Process(tweet)
		if len(output.Haikus) > 0 {
			ts.OutputChannel <- output
		}

	}
	return nil
}

func (ts *TweetStreamer) Output(out *HaikuOutput) error {
	if len(out.Haikus) == 0 {
		return nil
	}
	t := out.Tweet
	log.Printf("https://twitter.com/%v/status/%v %v", t.User.ScreenName, t.IDStr, t.FullText())
	for i, foundHaiku := range out.Haikus {
		color.Cyan.Printf("%d. %s\n", i+1, foundHaiku.String())

	}
	bytes, err := json.Marshal(out)
	if err != nil {
		panic(err)
	}
	//log.Printf(string(bytes))
	_, err = ts.OutFile.WriteString(fmt.Sprintf("%s\n", string(bytes)))
	if err != nil {
		return errors.Wrapf(err, "Error writing to file %s", ts.OutFile.Name())
	}
	return nil
}

func (ts *TweetStreamer) OutputLoop() error {
	now := time.Now().UTC()
	fileName := fmt.Sprintf("haiku_%s.json", now.Format(time.RFC3339))
	filePath := fmt.Sprintf("%s/%s", ts.Config.OutputPath, fileName)
	var err error
	ts.OutFile, err = os.Create(filePath)
	if err != nil {
		return errors.Wrapf(err, "Error creating file %s", filePath)
	}
	for tweet := range ts.OutputChannel {
		ts.Output(tweet)
	}
	return nil
}
