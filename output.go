package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gookit/color"
	"github.com/pkg/errors"
)

/*
TODO filter out anything ending in
or
a
and
are
his
*/

func (ts *TweetStreamer) Output(out *HaikuOutput) error {
	if len(out.Haikus) == 0 {
		return nil
	}
	t := out.Tweet
	log.Printf("https://twitter.com/%v/status/%v %v", t.User.ScreenName, t.IDStr, t.FullText())
	for i, foundHaiku := range out.Haikus {
		color.Cyan.Printf("%d. %s\n\n", i+1, foundHaiku.String())

	}
	bytes, err := json.Marshal(out)
	if err != nil {
		panic(err)
	}
	//log.Printf(string(bytes))
	_, err = ts.OutFile.WriteString(fmt.Sprintf("%s\n", string(bytes)))
	if err != nil {
		return errors.Wrapf(err, "Error writing to file %s", ts.OutFile.Name())
	}
	return nil
}

func (ts *TweetStreamer) OutputLoop() error {
	now := time.Now().UTC()
	fileName := fmt.Sprintf("haiku_%s.json", now.Format(time.RFC3339))
	filePath := fmt.Sprintf("%s/%s", ts.Config.OutputPath, fileName)
	var err error
	ts.OutFile, err = os.Create(filePath)
	if err != nil {
		return errors.Wrapf(err, "Error creating file %s", filePath)
	}
	for tweet := range ts.OutputChannel {
		ts.Output(tweet)
	}
	return nil
}
