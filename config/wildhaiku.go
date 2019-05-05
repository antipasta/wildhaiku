/*Package config contains the configuration necessary to run the wildhaiku application
 */
package config

import (
	"encoding/json"
	"io/ioutil"
)

// WildHaiku holds all configuration needed to run the WildHaiku daemon
type WildHaiku struct {
	ConsumerKey        string
	ConsumerSecret     string
	AccessToken        string
	AccessSecret       string
	TrackingKeywords   []string
	CorpusPath         string
	OutputPath         string
	ProcessWorkerCount int
}

// Load takes the path to the wildhaiku config, and returns an instance of a *WildHaiku config. Errors if file cannot be read or does not parse correctly
func Load(path string) (*WildHaiku, error) {
	cfg := WildHaiku{}
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
