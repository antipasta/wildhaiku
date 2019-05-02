package twitter

import (
	"bufio"
	"encoding/json"
	"io"
	"io/ioutil"
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
}

func NewStreamer(cfg *config.Streamer) *Streamer {
	processChannel := make(chan *Tweet, 10000)
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
func (ts *Streamer) connect() (*http.Response, error) {
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
	resp, err := ts.connect()
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	buf := bufio.NewReader(resp.Body)
	for {
		t, err := ts.TweetFromInput(buf)
		if err != nil {
			return err
		}
		if t == nil {
			continue
		}
		ts.ProcessChannel <- t
	}
}

func (ts *Streamer) TweetFromInput(reader *bufio.Reader) (*Tweet, error) {
	inBytes, err := reader.ReadBytes('\n')
	if err == io.EOF {
		return nil, nil
	}
	if len(inBytes) == 0 {
		// return nil and keep going
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "Error when trying to read from tweet stream")
	}
	return ts.ParseTweet(inBytes)
}

func (ts *Streamer) ParseTweet(inBytes []byte) (*Tweet, error) {
	t := Tweet{}
	err := json.Unmarshal(inBytes, &t)
	if err != nil {
		return nil, errors.Errorf("Error json decoding line [%v]: %v", string(inBytes), err)
	}
	if t.FullText() == "" {
		return nil, nil
	}
	if t.Lang != "en" {
		return nil, nil
	}
	return &t, nil
}
