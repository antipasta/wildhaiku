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
	InputChannel  <-chan *twitter.Tweet
	OutputChannel chan<- *Output
	Config        *config.Streamer
	corpus        *syllable.CMUCorpus
}

func NewProcessor(cfg *config.Streamer) (*Processor, error) {
	cmu, err := syllable.LoadCMUCorpus(cfg.CorpusPath)
	if err != nil {
		return nil, errors.Wrapf(err, "Error loading CMU corpus from %v", cfg.CorpusPath)
	}
	return &Processor{
		corpus: cmu,
	}, nil
}

func (p *Processor) ProcessLoop() error {
	for tweet := range p.InputChannel {
		output := p.Process(tweet)
		if len(p.InputChannel) > 0 {
			log.Printf("Channel size is %+v", len(p.InputChannel))
		}
		if len(output.Haikus) > 0 {
			p.OutputChannel <- output
		}

	}
	return nil
}

func (p *Processor) Process(t *twitter.Tweet) *Output {
	paragraph := p.corpus.ToSyllableParagraph(t.FullText())
	foundHaikus := paragraph.Subdivide(5, 7, 5)
	haikuStrings := [][]string{}
	for _, haiku := range foundHaikus {
		haikuStrings = append(haikuStrings, haiku.ToStringSlice())
	}
	return &Output{Tweet: t, Haikus: foundHaikus}
}
