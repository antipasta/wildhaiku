package haikudetector

import (
	"log"

	"gopkg.in/jdkato/prose.v2"
)

func ParagraphSyllables(paragraph string) []uint {
	var syllables []uint
	/*
		filename := "hyph-en-us.pat.txt"
		r, err := os.Open(filename)
		if err != nil {
			log.Fatal(err)
		}
		l, err := hyphenation.New(r)
		if err != nil {
			log.Fatal(err)
		}
	*/

	doc, err := prose.NewDocument(paragraph)
	if err != nil {
		log.Fatal(err)
	}
	for _, sentence := range doc.Sentences() {
		sentenceDoc, err := prose.NewDocument(sentence.Text)
		if err != nil {
			log.Fatal(err)
		}
		var total uint
		for _, v := range sentenceDoc.Tokens() {
			if v.Tag == "." {
				continue
			}
			//h := l.Hyphenate(v)
			count := syllableCount(v.Text)
			//log.Printf("word %+v: %+v, meta is %+v", v.Text, count, v.Tag)
			total += count
			/*
				if len(h) == 0 {
					total++
					continue
				}
				total += len(h)
				if firstBoundary := h[0]; firstBoundary <= 1 {
					log.Printf("first skipped for %+v", v)
					continue
				}
				if lastBoundary := h[len(h)-1]; lastBoundary >= len(v)-1 {
					log.Printf("last skipped for %+v", v)
					continue
				}
				total++
			*/
		}
		log.Printf("Total for [%+v]: %v", sentence, total)
		syllables = append(syllables, total)
	}
	return syllables
}
