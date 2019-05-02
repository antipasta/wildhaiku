package main

import (
	"flag"
	"log"

	"github.com/antipasta/wildhaiku/archive"
	"github.com/antipasta/wildhaiku/config"
	"github.com/antipasta/wildhaiku/haiku"
	"github.com/antipasta/wildhaiku/twitter"
)

var flagConfigPath string

func init() {
	flag.StringVar(&flagConfigPath, "config", "config.json", "Path to config file")
}

func NewStreamChannel() chan *twitter.Tweet {
	return make(chan *twitter.Tweet, 10000)
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

	for resp, err := ts.Connect(); ; {
		defer resp.Body.Close()
		if err != nil {
			log.Fatalf("Got error when connecting to twitter stream: %v", err)
		}
		err = ts.StreamLoop(resp.Body)
		if err != nil {
			log.Printf("Got stream error %+v. Reconnecting", err)
		}
	}
}
