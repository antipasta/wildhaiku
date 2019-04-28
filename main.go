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

func main() {
	flag.Parse()
	cfg, err := config.Load(flagConfigPath)
	if err != nil {
		panic(err)
	}
	ts := twitter.NewStreamer(cfg)
	diskArchiver, err := archive.NewDiskArchiver(cfg)
	if err != nil {
		panic(err)
	}
	haikuProcessor, err := haiku.NewProcessor(cfg)
	haikuProcessor.InputChannel = ts.ProcessChannel
	haikuProcessor.OutputChannel = diskArchiver.ArchiveChannel
	if err != nil {
		panic(err)
	}

	go diskArchiver.OutputLoop()

	for i := 0; i < cfg.ProcessWorkerCount; i++ {
		go haikuProcessor.ProcessLoop()
	}

	for {
		err = ts.StreamLoop()
		if err != nil {
			log.Printf("Got stream error %+v. Reconnecting", err)
		}
	}
}
