package cmd

import (
	"fmt"
	"math"
	"regexp"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	log "github.com/sirupsen/logrus"

	"github.com/hobbymarks/fdn/db"
	"github.com/hobbymarks/fdn/utils"
)

func ReplaceWords(inputName string) string {
	outName := inputName
	mask := func(s string) ([]string, []bool) {
		regescape := func(s string) string {
			s = strings.ReplaceAll(s, "+", "\\+")
			s = strings.ReplaceAll(s, "?", "\\?")
			s = strings.ReplaceAll(s, "*", "\\*")

			return s
		}

		words := []string{}
		wdmsk := []bool{}
		var termWords []db.TermWord
		_db := db.ConnectCFGDB()
		defer utils.DBClose(_db)
		rlt := _db.Find(&termWords)
		if rlt.Error != nil {
			log.Fatalf("retrive TermWord error %s", rlt.Error)
		}

		pts := []string{}
		for _, twd := range termWords {
			pts = append(pts, regescape(twd.OriginalLower))
		}
		rp := regexp.MustCompile(strings.Join(pts, "|"))
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

	words, wordMasks := mask(inputName)
	if len(words) != len(wordMasks) {
		log.Fatal("words not equal wordMasks")
	}

	newWords := []string{}

	var sep db.Separator
	var termWords []db.TermWord
	var toSepWords []db.ToSepWord
	_db := db.ConnectCFGDB()
	defer utils.DBClose(_db)
	rlt := _db.First(&sep)
	if rlt.Error != nil {
		log.Fatalf("retrieve Separator error %s", rlt.Error)
	}
	_sep := sep.Value
	rlt = _db.Find(&termWords)
	if rlt.Error != nil {
		log.Fatalf("retrieve TermWord error %s", rlt.Error)
	}
	rlt = _db.Find(&toSepWords)
	if rlt.Error != nil {
		log.Fatalf("retrieve ToSepWord error %s", rlt.Error)
	}
	slices.SortFunc(toSepWords, func(a, b db.ToSepWord) int {
		ra := utf8.RuneCountInString(a.Value)
		rb := utf8.RuneCountInString(b.Value)
		if ra != rb {
			return rb - ra
		}
		return strings.Compare(a.Value, b.Value)
	})

	rpCNS := regexp.MustCompile("[" + _sep + "]+")
	termWordMap := make(map[string]string)
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
	outName = strings.Join(newWords, "")
	outName = rpCNS.ReplaceAllString(outName, _sep)
	newWords = []string{}

	for _, wd := range strings.Split(outName, _sep) {
		if v, exist := termWordMap[wd]; exist {
			wd = v
		}
		newWords = append(newWords, wd)
	}
	outName = strings.Join(newWords, _sep)

	return outName
}

func ProcessHeadTail(inputName string) string {
	outName := inputName

	var sep db.Separator
	_db := db.ConnectCFGDB()
	defer utils.DBClose(_db)
	rlt := _db.First(&sep)
	if rlt.Error != nil {
		log.Fatalf("retrieve Separator error %s", rlt.Error)
	}
	_sep := sep.Value

	rpHTSeps := regexp.MustCompile("^" + _sep + "+" + "|" + _sep + "+" + "$")
	outName = rpHTSeps.ReplaceAllString(outName, "")

	return outName
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

func ArrayContainsElemenet[T comparable](s []T, e T) bool {
	return slices.Contains(s, e)
}

// FDNedFrom returns the FDN-normalized form of input by applying, in order, ReplaceWords
// (configured term and separator rules from the config DB), ProcessHeadTail (strip leading
// and trailing separator runs), and ASCHead (synthetic ASCII prefix when the first rune is
// not a Latin letter or digit).
func FDNedFrom(input string) string {
	// TODO(hm): Optimize name
	output := input
	output = ReplaceWords(input)
	output = ProcessHeadTail(output)
	// output = ASCHead(output)
	return output
}
