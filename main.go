/*Package wildhaiku implements wildhaiku, a utility for:

 * Reading tweets off of a Public API stream

 * Parsing out Haikus(one or more sentences where, when combined, are split on lines of 5 syllables, 7 syllables, 5 syllables).

 * Outputting found haikus and their corresponding Tweets to disk
 */
package main

import (
	"flag"
	"log"
	"time"

	"github.com/antipasta/wildhaiku/archive"
	"github.com/antipasta/wildhaiku/config"
	"github.com/antipasta/wildhaiku/haiku"
	"github.com/antipasta/wildhaiku/twitter"
)

var flagConfigPath string

func init() {
	flag.StringVar(&flagConfigPath, "config", "config.json", "Path to config file")
}

func main() {
	flag.Parse()
	cfg, err := config.Load(flagConfigPath)
	if err != nil {
		log.Fatalf("Error loading config file[%v]: %v", flagConfigPath, err)
	}
	ts := twitter.NewStreamer(cfg)
	diskArchiver, err := archive.NewDiskArchiver(cfg)
	if err != nil {
		log.Fatalf("Error initializing disk archiver: %v", err)
	}
	haikuProcessor, err := haiku.NewProcessor(cfg, ts.ProcessChannel, diskArchiver.ArchiveChannel)
	if err != nil {
		log.Fatalf("Error initializing haiku processor: %v", err)
	}

	go diskArchiver.OutputLoop()

	for i := 0; i < cfg.ProcessWorkerCount; i++ {
		go haikuProcessor.ProcessLoop()
	}

	for {
		resp, err := ts.Connect()
		if err != nil {
			resp.Body.Close()
			log.Printf("Got error when connecting to twitter stream, sleeping and reconnecting: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		err = ts.StreamLoop(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Printf("Got stream error %+v. Reconnecting", err)
		}
	}
}
