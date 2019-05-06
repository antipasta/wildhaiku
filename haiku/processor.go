/*Package haiku is used for processing Tweets and parsing out found Haikus
 */
package haiku

import (
	"log"

	"github.com/antipasta/wildhaiku/config"
	"github.com/antipasta/wildhaiku/syllable"
	"github.com/antipasta/wildhaiku/twitter"
	"github.com/pkg/errors"
)

// Output is a Tweet bundled with all found haikus from that tweet, used for outputting to file
type Output struct {
	Haikus []syllable.Haiku
	Tweet  *twitter.Tweet
}

// Processor reads in tweets on a channel, and outputs them to an output channel
type Processor struct {
	inputChannel  <-chan *twitter.Tweet
	outputChannel chan<- *Output
	Config        *config.WildHaiku
	corpus        *syllable.CMUCorpus
}

// NewProcessor creates a new instance of the processor class, using specified input and output channels
func NewProcessor(cfg *config.WildHaiku, tweetIn <-chan *twitter.Tweet, processedOut chan<- *Output) (*Processor, error) {
	cmu, err := syllable.NewCMUCorpus(cfg.CorpusPath)
	if err != nil {
		return nil, errors.Wrapf(err, "Error loading CMU corpus from %v", cfg.CorpusPath)
	}
	return &Processor{
		corpus:        cmu,
		inputChannel:  tweetIn,
		outputChannel: processedOut,
	}, nil
}

// ProcessLoop reads in Tweets on input channel, and if any haikus are found,outputs an Output object  on the output channel
func (p *Processor) ProcessLoop() error {
	for tweet := range p.inputChannel {
		output := p.process(tweet)
		if output == nil {
			// Could not find haiku
			continue
		}
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
	paragraph, err := p.corpus.NewParagraph(t.FullText())
	if err != nil {
		return nil
	}
	foundHaikus := paragraph.Subdivide(5, 7, 5)
	haikuStrings := [][3]string{}
	for _, haiku := range foundHaikus {
		haikuStrings = append(haikuStrings, haiku.ToStringArray())
	}
	return &Output{Tweet: t, Haikus: foundHaikus}
}
