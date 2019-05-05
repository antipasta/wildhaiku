/*Package twitter is used for interacting with tweet streams from Twitter's Public API
 */
package twitter

// Tweet is a representation of a subset of fields of a Tweet from Twitter's API
type Tweet struct {
	IDStr string `json:"id_str"`
	Lang  string `json:"lang"`
	User  struct {
		ScreenName string `json:"screen_name"`
	} `json:"user"`
	Text          string `json:"text,omitempty"`
	ExtendedTweet *struct {
		FullText string `json:"full_text,omitempty"`
	} `json:"extended_tweet,omitempty"`
	RetweetedStatus *Tweet `json:"retweeted_status,omitempty"`
}

// FullText returns the full text of the tweet. t.ExtendedTweet.FullText if it exists, else t.Text
func (t *Tweet) FullText() string {
	if t.ExtendedTweet != nil && len(t.ExtendedTweet.FullText) > 0 {
		return t.ExtendedTweet.FullText
	}
	return t.Text
}
