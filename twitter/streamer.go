/*Package twitter contains utilities for connecting and parsing Tweets from a Twitter JSON data stream
 */
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

// Streamer is responsible for connecting to and reading from a Twitter public API stream
type Streamer struct {
	Config         *config.WildHaiku
	ConsumerKeys   *oauth.Credentials
	Token          *oauth.Credentials
	Client         *oauth.Client
	httpClient     *http.Client
	ProcessChannel chan *Tweet
}

// NewStreamer returns a twitter.Streamer object
func NewStreamer(cfg *config.WildHaiku) *Streamer {
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
	ts := Streamer{
		Config:         cfg,
		ConsumerKeys:   &consumerKeys,
		Token:          &token,
		Client:         &client,
		httpClient:     &http.Client{},
		ProcessChannel: processChannel,
	}
	return &ts
}

// Connect connects to a Twitter public API stream and returns the response for reading
func (ts *Streamer) Connect() (*http.Response, error) {
	resp, err := ts.Client.Post(
		ts.httpClient,
		ts.Token,
		"https://stream.twitter.com/1.1/statuses/filter.json",
		url.Values{"lang": []string{"en"},
			"track":      ts.Config.TrackingKeywords,
			"tweet_mode": []string{"extended"}})
	if err != nil {
		return nil, errors.Wrapf(err, "Caught error when connecting to twitter stream")
	}

	if resp.StatusCode != http.StatusOK {
		all, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.Errorf("Received non-OK error code [%v] [%v] when connecting to twitter stream: %v", resp.StatusCode, resp.Status, string(all))

	}
	return resp, nil
}

// StreamLoop reads off of a JSON stream of public tweets and sends json decoded Tweets to the ProcessChannel
func (ts *Streamer) StreamLoop(stream io.Reader) error {
	buf := bufio.NewReader(stream)
	for {
		t, err := ts.tweetFromInput(buf)
		if err != nil {
			return err
		}
		if t == nil {
			continue
		}
		ts.ProcessChannel <- t
	}
}

func (ts *Streamer) tweetFromInput(reader *bufio.Reader) (*Tweet, error) {
	inBytes, err := reader.ReadBytes('\n')
	if err == io.EOF {
		return nil, err
	}
	if len(inBytes) == 0 {
		return nil, errors.Errorf("No bytes received")
	}
	if string(inBytes) == "\r\n" {
		// return nil and keep going
		log.Printf("Got keepalive ping")
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "Error when trying to read from tweet stream")
	}

	return ts.parseTweet(inBytes)
}

func (ts *Streamer) parseTweet(inBytes []byte) (*Tweet, error) {
	t := Tweet{}
	err := json.Unmarshal(inBytes, &t)
	if err != nil {
		log.Printf("Error json decoding line [%v]: %v", string(inBytes), err)
		return nil, errors.Errorf("Error json decoding line [%v]: %v", string(inBytes), err)
	}
	if t.RetweetedStatus != nil {
		// A standard retweet does not have any additional text, may as well work off original for proper attribution
		t = *t.RetweetedStatus
	}
	if t.FullText() == "" {
		return nil, nil
	}
	if t.Lang != "en" {
		return nil, nil
	}
	return &t, nil
}
