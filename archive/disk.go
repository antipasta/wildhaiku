/*Package archive is used to save Haikus found within Tweets to disk
 */
package archive

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/antipasta/wildhaiku/config"
	"github.com/antipasta/wildhaiku/haiku"
	"github.com/antipasta/wildhaiku/syllable"
	"github.com/gookit/color"
	"github.com/pkg/errors"
)

var suffixBlacklist = map[string]bool{
	"or":   true,
	"a":    true,
	"and":  true,
	"are":  true,
	"his":  true,
	"but":  true,
	"to":   true,
	"is":   true,
	"in":   true,
	"he":   true,
	"she":  true,
	"as":   true,
	"our":  true,
	"the":  true,
	"of":   true,
	"if":   true,
	"an":   true,
	"my":   true,
	"your": true,
}

// DiskArchiver receives *haiku.Output over a channel and writes to disk
type DiskArchiver struct {
	ArchiveChannel chan *haiku.Output
	outFile        *os.File
	outFilePath    string
	symlinkPath    string
	Config         *config.WildHaiku
}

// NewDiskArchiver creates an instance of DiskArchiver, errors if it cannot access path specified in config.OutputPath
func NewDiskArchiver(cfg *config.WildHaiku) (*DiskArchiver, error) {
	archiveChan := make(chan *haiku.Output, 10000)
	now := time.Now().UTC()
	fileName := fmt.Sprintf("haiku_%s.json", now.Format(time.RFC3339))
	absOutPath, err := filepath.Abs(cfg.OutputPath)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not determine absolute path for %s", absOutPath)
	}
	if _, err := os.Stat(absOutPath); err != nil {
		return nil, errors.Wrapf(err, "Error accessing output file path %s", absOutPath)
	}

	filePath := filepath.Join(absOutPath, fileName)
	symLink := filepath.Join(absOutPath, "current.json")
	return &DiskArchiver{Config: cfg, ArchiveChannel: archiveChan, outFilePath: filePath, symlinkPath: symLink}, nil
}

func (da *DiskArchiver) output(out *haiku.Output) error {
	if len(out.Haikus) == 0 {
		return nil
	}
	filteredHaikus := []syllable.Haiku{}
	for _, foundHaiku := range out.Haikus {
		if !suffixBlacklist[strings.ToLower(foundHaiku.FinalWord())] {
			filteredHaikus = append(filteredHaikus, foundHaiku)
			log.Printf("https://twitter.com/%s/status/%s", out.Tweet.User.ScreenName, out.Tweet.IDStr)
			color.Cyan.Printf("%s\n\n", foundHaiku.String())
		}

	}
	if len(filteredHaikus) == 0 {
		// filtered out some junk
		return nil
	}
	out.Haikus = filteredHaikus
	bytes, err := json.Marshal(out)
	if err != nil {
		return errors.Wrapf(err, "Error marshalling json for %+v", out)
	}
	_, err = da.outFile.WriteString(fmt.Sprintf("%s\n", string(bytes)))
	if err != nil {
		return errors.Wrapf(err, "Error writing to file %s", da.outFile.Name())
	}
	return nil
}

//OutputLoop writes haiku.Output to disk, in a timestampped file based on when the function is first entered. Also creates a symlink current.json pointing at the file being written to
func (da *DiskArchiver) OutputLoop() error {
	var err error
	da.outFile, err = os.Create(da.outFilePath)
	if err != nil {
		return errors.Wrapf(err, "Error creating file %s", da.outFilePath)
	}
	defer da.outFile.Close()
	if _, err := os.Lstat(da.symlinkPath); err == nil {
		err = os.Remove(da.symlinkPath)
		if err != nil {
			return errors.Wrapf(err, "Error removing symlink %v", da.symlinkPath)
		}
	}
	err = os.Symlink(da.outFilePath, da.symlinkPath)
	if err != nil {
		return err
	}
	log.Printf("Writing to file %s (and symlink %s)", da.outFilePath, da.symlinkPath)
	for tweet := range da.ArchiveChannel {
		err = da.output(tweet)
		if err != nil {
			log.Printf("Got error %v when saving tweet %+v to disk, skipping", err, tweet)
		}
	}
	return nil
}
