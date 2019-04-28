package config

import (
	"encoding/json"
	"io/ioutil"
)

type Streamer struct {
	ConsumerKey        string
	ConsumerSecret     string
	AccessToken        string
	AccessSecret       string
	TrackingKeywords   []string
	CorpusPath         string
	OutputPath         string
	ProcessWorkerCount int
}

func Load(path string) (*Streamer, error) {
	cfg := Streamer{}
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
