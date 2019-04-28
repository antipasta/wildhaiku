package twitter

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

func (t *Tweet) FullText() string {
	if t.ExtendedTweet != nil && len(t.ExtendedTweet.FullText) > 0 {
		return t.ExtendedTweet.FullText
	}
	return t.Text
}
