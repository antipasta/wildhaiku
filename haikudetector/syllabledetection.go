package haikudetector

import "strings"

//JUNK I FOUND
func isVowel(chr_ rune) bool {
	chr := string(chr_)
	if strings.EqualFold(chr, "a") || strings.EqualFold(chr, "e") ||
		strings.EqualFold(chr, "i") ||
		strings.EqualFold(chr, "o") ||
		strings.EqualFold(chr, "u") ||
		strings.EqualFold(chr, "y") {
		return true
	}
	return false
}

func HasSuffix(s string, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

func syllableCount(word string) uint {
	var count uint
	count = 0
	lastWasVowel := false
	word = strings.ToLower(word)
	word = strings.Replace(word, "ome", "um", -1)
	word = strings.Replace(word, "ime", "im", -1)
	word = strings.Replace(word, "imn", "imen", -1)
	word = strings.Replace(word, "ine", "in", -1)
	word = strings.Replace(word, "ely", "ly", -1)
	word = strings.Replace(word, "ure", "ur", -1)
	word = strings.Replace(word, "ery", "ry", -1)
	for _, chr := range word {
		if isVowel(chr) {
			if lastWasVowel == false {
				count += 1
			}
			lastWasVowel = true
		} else {
			lastWasVowel = false
		}
	}
	if HasSuffix(word, "ing") || HasSuffix(word, "ings") {
		if len(word) > 4 && isVowel(rune(word[len(word)-4])) {
			count += 1
		}
	}
	if HasSuffix(word, "e") && HasSuffix(word, "le") == false {
		count -= 1
	}
	if HasSuffix(word, "e's") {
		if len(word) > 5 && isVowel(rune(word[len(word)-5])) {
			count -= 1
		}
	}
	if HasSuffix(word, "ed") && HasSuffix(word, "ted") == false &&
		HasSuffix(word, "ded") == false {
		count -= 1
	}
	if count < 1 {
		count = 1
	}
	return count
}
