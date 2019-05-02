package haiku

import (
	"log"

	"github.com/antipasta/wildhaiku/config"
	"github.com/antipasta/wildhaiku/syllable"
	"github.com/antipasta/wildhaiku/twitter"
	"github.com/pkg/errors"
)

type Output struct {
	Haikus []syllable.Haiku
	Tweet  *twitter.Tweet
}

type Processor struct {
	inputChannel  <-chan *twitter.Tweet
	outputChannel chan<- *Output
	Config        *config.Streamer
	corpus        *syllable.CMUCorpus
}

func NewProcessor(cfg *config.Streamer, tweetIn <-chan *twitter.Tweet, processedOut chan<- *Output) (*Processor, error) {
	cmu, err := syllable.LoadCMUCorpus(cfg.CorpusPath)
	if err != nil {
		return nil, errors.Wrapf(err, "Error loading CMU corpus from %v", cfg.CorpusPath)
	}
	return &Processor{
		corpus:        cmu,
		inputChannel:  tweetIn,
		outputChannel: processedOut,
	}, nil
}

func (p *Processor) ProcessLoop() error {
	for tweet := range p.inputChannel {
		output := p.process(tweet)
		if len(p.inputChannel) > 0 {
			log.Printf("Channel size is %+v", len(p.inputChannel))
		}
		if len(output.Haikus) > 0 {
			p.outputChannel <- output
		}

	}
	return nil
}

func (p *Processor) process(t *twitter.Tweet) *Output {
	paragraph := p.corpus.ToSyllableParagraph(t.FullText())
	foundHaikus := paragraph.Subdivide(5, 7, 5)
	haikuStrings := [][]string{}
	for _, haiku := range foundHaikus {
		haikuStrings = append(haikuStrings, haiku.ToStringSlice())
	}
	return &Output{Tweet: t, Haikus: foundHaikus}
}
