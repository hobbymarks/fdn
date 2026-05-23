package cmd

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/hobbymarks/fdn/db"
)

var termWordRegexCache struct {
	sync.RWMutex
	pat string
	re  *regexp.Regexp
}

var configDataCache struct {
	sync.Mutex
	termWords  []db.TermWord
	separator  db.Separator
	toSepWords []db.ToSepWord
	loaded     bool
}

var tempReplacements map[string]string

func SetTempReplacements(m map[string]string) {
	tempReplacements = m
}

func InvalidateCaches() {
	termWordRegexCache.Lock()
	termWordRegexCache.pat = ""
	termWordRegexCache.re = nil
	termWordRegexCache.Unlock()

	configDataCache.Lock()
	configDataCache.loaded = false
	configDataCache.Unlock()
}

func termWordAlternationRE(pattern string) *regexp.Regexp {
	if pattern == "" {
		return nil
	}
	termWordRegexCache.RLock()
	if termWordRegexCache.pat == pattern && termWordRegexCache.re != nil {
		re := termWordRegexCache.re
		termWordRegexCache.RUnlock()
		return re
	}
	termWordRegexCache.RUnlock()

	termWordRegexCache.Lock()
	defer termWordRegexCache.Unlock()
	if termWordRegexCache.pat == pattern && termWordRegexCache.re != nil {
		return termWordRegexCache.re
	}
	termWordRegexCache.re = regexp.MustCompile(pattern)
	termWordRegexCache.pat = pattern
	return termWordRegexCache.re
}

func escapeTermForRegex(s string) string {
	s = strings.ReplaceAll(s, "+", "\\+")
	s = strings.ReplaceAll(s, "?", "\\?")
	s = strings.ReplaceAll(s, "*", "\\*")
	return s
}

func termAlternationPattern(termWords []db.TermWord) string {
	if len(termWords) == 0 {
		return ""
	}
	pts := make([]string, 0, len(termWords))
	for _, twd := range termWords {
		pts = append(pts, escapeTermForRegex(twd.OriginalLower))
	}
	return strings.Join(pts, "|")
}

func maskSegments(s string, termWords []db.TermWord) ([]string, []bool) {
	pat := termAlternationPattern(termWords)
	rp := termWordAlternationRE(pat)
	if rp == nil {
		return []string{s}, []bool{false}
	}

	words := []string{}
	wdmsk := []bool{}
	allSliceIndex := rp.FindAllStringIndex(s, -1)
	cur := 0
	for _, slice := range allSliceIndex {
		if slice[0] > cur {
			words = append(words, s[cur:slice[0]])
			wdmsk = append(wdmsk, false)
		}
		words = append(words, s[slice[0]:slice[1]])
		wdmsk = append(wdmsk, true)
		cur = slice[1]
	}
	if cur < len(s) {
		words = append(words, s[cur:])
		wdmsk = append(wdmsk, false)
	}
	return words, wdmsk
}

var sepReplacerCache struct {
	sync.RWMutex
	sep string
	re  *regexp.Regexp
}

func getSepCollapseRe(sep string) *regexp.Regexp {
	sepReplacerCache.RLock()
	if sepReplacerCache.sep == sep && sepReplacerCache.re != nil {
		re := sepReplacerCache.re
		sepReplacerCache.RUnlock()
		return re
	}
	sepReplacerCache.RUnlock()

	sepReplacerCache.Lock()
	defer sepReplacerCache.Unlock()
	if sepReplacerCache.sep == sep && sepReplacerCache.re != nil {
		return sepReplacerCache.re
	}
	sepReplacerCache.re = regexp.MustCompile("[" + sep + "]+")
	sepReplacerCache.sep = sep
	return sepReplacerCache.re
}

func loadConfigCache() error {
	configDataCache.Lock()
	defer configDataCache.Unlock()
	if configDataCache.loaded {
		return nil
	}

	conn, err := db.ConnectCFGDB()
	if err != nil {
		return err
	}

	if result := conn.Find(&configDataCache.termWords); result.Error != nil {
		return fmt.Errorf("retrieve TermWord: %w", result.Error)
	}

	if result := conn.First(&configDataCache.separator); result.Error != nil {
		return fmt.Errorf("retrieve Separator: %w", result.Error)
	}

	if result := conn.Find(&configDataCache.toSepWords); result.Error != nil {
		return fmt.Errorf("retrieve ToSepWord: %w", result.Error)
	}
	slices.SortFunc(configDataCache.toSepWords, func(a, b db.ToSepWord) int {
		ra := utf8.RuneCountInString(a.Value)
		rb := utf8.RuneCountInString(b.Value)
		if ra != rb {
			return rb - ra
		}
		return strings.Compare(a.Value, b.Value)
	})

	configDataCache.loaded = true
	return nil
}

func ReplaceWords(inputName string) (string, error) {
	if err := loadConfigCache(); err != nil {
		return "", err
	}

	termWords := configDataCache.termWords
	sepStr := configDataCache.separator.Value
	toSepWords := configDataCache.toSepWords

	if len(tempReplacements) > 0 {
		for orig, target := range tempReplacements {
			termWords = append(termWords, db.TermWord{
				OriginalLower: strings.ToLower(orig),
				TargetWord:    target,
			})
		}
	}

	words, wordMasks := maskSegments(inputName, termWords)
	if len(words) != len(wordMasks) {
		return "", errors.New("words not equal wordMasks")
	}

	newWords := []string{}
	rpCNS := getSepCollapseRe(sepStr)
	termWordMap := make(map[string]string, len(termWords))
	for _, twd := range termWords {
		termWordMap[twd.OriginalLower] = twd.TargetWord
	}
	for idx, wd := range words {
		if !wordMasks[idx] {
			for _, sw := range toSepWords {
				wd = strings.ReplaceAll(wd, sw.Value, sepStr)
			}
		}
		newWords = append(newWords, wd)
	}
	outName := strings.Join(newWords, "")
	outName = rpCNS.ReplaceAllString(outName, sepStr)
	newWords = newWords[:0]

	for wd := range strings.SplitSeq(outName, sepStr) {
		if v, exist := termWordMap[wd]; exist {
			wd = v
		}
		newWords = append(newWords, wd)
	}
	outName = strings.Join(newWords, sepStr)

	return outName, nil
}

func ProcessHeadTail(inputName string) (string, error) {
	if err := loadConfigCache(); err != nil {
		return "", err
	}
	sepStr := configDataCache.separator.Value

	rpHTSeps := regexp.MustCompile("^" + sepStr + "+" + "|" + sepStr + "+" + "$")
	return rpHTSeps.ReplaceAllString(inputName, ""), nil
}

func FDNedFrom(input string) (string, error) {
	out, err := ReplaceWords(input)
	if err != nil {
		return "", err
	}
	out, err = ProcessHeadTail(out)
	if err != nil {
		return "", err
	}
	return out, nil
}
