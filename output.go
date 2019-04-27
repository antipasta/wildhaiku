package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/antipasta/wildhaiku/syllable"
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
but
to
is
in?
he?
she?
as
our
*/

var suffixBlacklist = map[string]bool{
	"or":  true,
	"a":   true,
	"and": true,
	"are": true,
	"his": true,
	"but": true,
	"to":  true,
	"is":  true,
	"in":  true,
	"he":  true,
	"she": true,
	"as":  true,
	"our": true,
	"the": true,
	"of":  true,
	"if":  true,
	"an":  true,
	//"my":  true,
}

func (ts *TweetStreamer) Output(out *HaikuOutput) error {
	if len(out.Haikus) == 0 {
		return nil
	}
	filteredHaikus := []syllable.Haiku{}
	for i, foundHaiku := range out.Haikus {
		if !suffixBlacklist[strings.ToLower(foundHaiku.FinalWord())] {
			filteredHaikus = append(filteredHaikus, foundHaiku)
			color.Cyan.Printf("%d. %s\n\n", i+1, foundHaiku.String())
		}

	}
	if len(filteredHaikus) == 0 {
		// filtered out some junk
		return nil
	}
	out.Haikus = filteredHaikus
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
