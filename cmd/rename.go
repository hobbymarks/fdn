package cmd

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"slices"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/hobbymarks/fdn/db"
	"github.com/hobbymarks/fdn/utils"
)

var termWordRegexCache struct {
	sync.RWMutex
	pat string
	re  *regexp.Regexp
}

func invalidateTermWordRegexCache() {
	termWordRegexCache.Lock()
	defer termWordRegexCache.Unlock()
	termWordRegexCache.pat = ""
	termWordRegexCache.re = nil
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

func ReplaceWords(inputName string) (string, error) {
	_db := db.ConnectCFGDB()
	defer utils.DBClose(_db)

	var termWords []db.TermWord
	if rlt := _db.Find(&termWords); rlt.Error != nil {
		return "", fmt.Errorf("retrieve TermWord: %w", rlt.Error)
	}

	var sep db.Separator
	if rlt := _db.First(&sep); rlt.Error != nil {
		return "", fmt.Errorf("retrieve Separator: %w", rlt.Error)
	}
	_sep := sep.Value

	var toSepWords []db.ToSepWord
	if rlt := _db.Find(&toSepWords); rlt.Error != nil {
		return "", fmt.Errorf("retrieve ToSepWord: %w", rlt.Error)
	}
	slices.SortFunc(toSepWords, func(a, b db.ToSepWord) int {
		ra := utf8.RuneCountInString(a.Value)
		rb := utf8.RuneCountInString(b.Value)
		if ra != rb {
			return rb - ra
		}
		return strings.Compare(a.Value, b.Value)
	})

	words, wordMasks := maskSegments(inputName, termWords)
	if len(words) != len(wordMasks) {
		return "", errors.New("words not equal wordMasks")
	}

	newWords := []string{}
	rpCNS := regexp.MustCompile("[" + _sep + "]+")
	termWordMap := make(map[string]string, len(termWords))
	for _, twd := range termWords {
		termWordMap[twd.OriginalLower] = twd.TargetWord
	}
	for idx, wd := range words {
		if !wordMasks[idx] {
			for _, sw := range toSepWords {
				wd = strings.ReplaceAll(wd, sw.Value, _sep)
			}
		}
		newWords = append(newWords, wd)
	}
	outName := strings.Join(newWords, "")
	outName = rpCNS.ReplaceAllString(outName, _sep)
	newWords = newWords[:0]

	for _, wd := range strings.Split(outName, _sep) {
		if v, exist := termWordMap[wd]; exist {
			wd = v
		}
		newWords = append(newWords, wd)
	}
	outName = strings.Join(newWords, _sep)

	return outName, nil
}

func ProcessHeadTail(inputName string) (string, error) {
	_db := db.ConnectCFGDB()
	defer utils.DBClose(_db)

	var sep db.Separator
	if rlt := _db.First(&sep); rlt.Error != nil {
		return "", fmt.Errorf("retrieve Separator: %w", rlt.Error)
	}
	_sep := sep.Value

	rpHTSeps := regexp.MustCompile("^" + _sep + "+" + "|" + _sep + "+" + "$")
	return rpHTSeps.ReplaceAllString(inputName, ""), nil
}

func ASCHead(inputName string) string {
	outName := inputName
	sa := []rune(outName)
	var ascH strings.Builder
	proxCS := func(c rune) string {
		if c > 'Z' {
			return fmt.Sprintf(
				"%c",
				int(c)-int(math.Ceil(float64(c-'Z')/26)*26),
			)
		} else if c < 'A' {
			return fmt.Sprintf("%c", int(c)+int(math.Ceil(float64('A'-c)/26)*26))
		} else {
			return fmt.Sprintf("%c", int(c))
		}
	}
	if len(sa) >= 1 {
		if !(unicode.IsDigit(sa[0])) && !(unicode.IsLower(sa[0])) && !(unicode.IsUpper(sa[0])) {
			uL := min(len(sa), 3)
			for i := range uL {
				ascH.WriteString(proxCS(sa[i]))
			}
		}
	}
	return ascH.String() + outName
}

func ArrayContainsElement[T comparable](s []T, e T) bool {
	return slices.Contains(s, e)
}

// FDNedFrom applies ReplaceWords then ProcessHeadTail using the config database.
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
