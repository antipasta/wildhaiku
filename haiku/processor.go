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
	Config        *config.WildHaiku
	corpus        *syllable.CMUCorpus
}

func NewProcessor(cfg *config.WildHaiku, tweetIn <-chan *twitter.Tweet, processedOut chan<- *Output) (*Processor, error) {
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
	paragraph := p.corpus.NewParagraph(t.FullText())
	foundHaikus := paragraph.Subdivide(5, 7, 5)
	haikuStrings := [][3]string{}
	for _, haiku := range foundHaikus {
		haikuStrings = append(haikuStrings, haiku.ToStringArray())
	}
	return &Output{Tweet: t, Haikus: foundHaikus}
}
