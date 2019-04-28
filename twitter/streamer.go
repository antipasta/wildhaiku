package twitter

import (
	"bufio"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/antipasta/wildhaiku/config"
	"github.com/gomodule/oauth1/oauth"
	"github.com/pkg/errors"
)

type Streamer struct {
	Config         *config.Streamer
	ConsumerKeys   *oauth.Credentials
	Token          *oauth.Credentials
	Client         *oauth.Client
	httpClient     *http.Client
	ProcessChannel chan *Tweet
	//corpus         *syllable.CMUCorpus
	//OutputChannel  chan *output.Haiku
	//OutFile *os.File
}

func NewStreamer(cfg *config.Streamer) *Streamer {
	processChannel := make(chan *Tweet, 10000)
	//outChannel := make(chan *output.Haiku, 10000)
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
	ts := Streamer{Config: cfg, ConsumerKeys: &consumerKeys, Token: &token, Client: &client, httpClient: &http.Client{}, ProcessChannel: processChannel}
	return &ts
}
func (ts *Streamer) Connect() (*http.Response, error) {
	resp, err := ts.Client.Post(ts.httpClient, ts.Token, "https://stream.twitter.com/1.1/statuses/filter.json", url.Values{"lang": []string{"en"}, "track": ts.Config.TrackingKeywords, "tweet_mode": []string{"extended"}})
	if err != nil {
		return nil, errors.Wrapf(err, "Caught error when connecting to twitter stream")
	}

	if resp.StatusCode != http.StatusOK {
		all, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.Errorf("Received non-OK error code [%v] [%v] when connecting to twitter stream: %v", resp.StatusCode, resp.Status, string(all))

	}
	return resp, nil
}

func (ts *Streamer) StreamLoop() error {
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
			log.Printf("Error parsing %+v", string(line))
			continue
		}
		if t.RetweetedStatus != nil {
			//t = *t.RetweetedStatus
		}
		if t.FullText() == "" || t.Lang != "en" {
			continue
		}
		ts.ProcessChannel <- &t
	}
}
