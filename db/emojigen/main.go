// Emojigen rebuilds ../default_emoji_sepwords.txt from db/emoji-data.txt,
// db/emoji-sequences.txt, and db/emoji-zwj-sequences.txt. It is not linked into the fdn
// binary and is not invoked by the CLI at runtime.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

func main() {
	dir := flag.String("dir", ".", "directory containing emoji-data.txt, emoji-sequences.txt, emoji-zwj-sequences.txt")
	out := flag.String("o", "default_emoji_sepwords.txt", "output path: each line is literal<TAB>U+HEX U+HEX ...")
	flag.Parse()

	dataPath := filepath.Join(*dir, "emoji-data.txt")
	seqPath := filepath.Join(*dir, "emoji-sequences.txt")
	zwjPath := filepath.Join(*dir, "emoji-zwj-sequences.txt")

	seen := make(map[string]struct{})

	b, err := os.ReadFile(dataPath)
	if err != nil {
		exitErr(err)
	}
	for line := range strings.SplitSeq(string(b), "\n") {
		for _, s := range parseEmojiDataLine(line) {
			addSepWord(seen, s)
		}
	}

	b, err = os.ReadFile(seqPath)
	if err != nil {
		exitErr(err)
	}
	for line := range strings.SplitSeq(string(b), "\n") {
		for _, s := range parseEmojiSequenceLine(line) {
			addSepWord(seen, s)
		}
	}

	b, err = os.ReadFile(zwjPath)
	if err != nil {
		exitErr(err)
	}
	for line := range strings.SplitSeq(string(b), "\n") {
		for _, s := range parseEmojiSequenceLine(line) {
			addSepWord(seen, s)
		}
	}

	var all []string
	for s := range seen {
		if s == "" {
			continue
		}
		all = append(all, s)
	}
	sort.Slice(all, func(i, j int) bool {
		ri := utf8.RuneCountInString(all[i])
		rj := utf8.RuneCountInString(all[j])
		if ri != rj {
			return ri > rj
		}
		return all[i] < all[j]
	})

	var sb strings.Builder
	for _, s := range all {
		if strings.ContainsAny(s, "\n\r\t") {
			continue
		}
		sb.WriteString(s)
		sb.WriteByte('\t')
		sb.WriteString(unicodeCodepointAnnotation(s))
		sb.WriteByte('\n')
	}
	if err := os.WriteFile(*out, []byte(sb.String()), 0o644); err != nil {
		exitErr(err)
	}
	fmt.Fprintf(os.Stderr, "wrote %d strings to %s\n", len(all), *out)
}

func exitErr(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func unicodeCodepointAnnotation(s string) string {
	var b strings.Builder
	for i, r := range s {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString("U+")
		b.WriteString(strings.ToUpper(strconv.FormatInt(int64(r), 16)))
	}
	return b.String()
}

func addSepWord(seen map[string]struct{}, s string) {
	if s == "" || omitEmojiSepWord(s) {
		return
	}
	seen[s] = struct{}{}
}

func omitEmojiSepWord(s string) bool {
	if utf8.RuneCountInString(s) != 1 {
		return false
	}
	r, _ := utf8.DecodeRuneInString(s)
	return r >= '0' && r <= '9'
}

var emojiDataProps = map[string]struct{}{
	"Emoji":                 {},
	"Extended_Pictographic": {},
	"Emoji_Presentation":    {},
	"Emoji_Modifier_Base":   {},
	"Emoji_Modifier":        {},
	"Emoji_Component":       {},
}

func parseEmojiDataLine(line string) []string {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return nil
	}
	if !strings.Contains(line, ";") {
		return nil
	}
	parts := strings.SplitN(line, ";", 3)
	if len(parts) < 2 {
		return nil
	}
	prop := strings.TrimSpace(parts[1])
	if i := strings.IndexByte(prop, '#'); i >= 0 {
		prop = strings.TrimSpace(prop[:i])
	}
	if _, ok := emojiDataProps[prop]; !ok {
		return nil
	}
	return expandCPField(strings.TrimSpace(parts[0]))
}

func parseEmojiSequenceLine(line string) []string {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return nil
	}
	if !strings.Contains(line, ";") {
		return nil
	}
	parts := strings.SplitN(line, ";", 3)
	if len(parts) < 2 {
		return nil
	}
	return expandCPField(strings.TrimSpace(parts[0]))
}

func expandCPField(field string) []string {
	field = strings.TrimSpace(field)
	if field == "" {
		return nil
	}
	toks := strings.Fields(field)
	if len(toks) == 1 && strings.Contains(toks[0], "..") {
		return expandRangeToken(toks[0])
	}
	var rs []rune
	for _, t := range toks {
		if strings.Contains(t, "..") {
			for _, s := range expandRangeToken(t) {
				rs = append(rs, []rune(s)...)
			}
			continue
		}
		v, err := parseHexRune(t)
		if err != nil {
			return nil
		}
		rs = append(rs, v)
	}
	if len(rs) == 0 {
		return nil
	}
	return []string{string(rs)}
}

func expandRangeToken(tok string) []string {
	tok = strings.TrimSpace(tok)
	a, b, ok := strings.Cut(tok, "..")
	if !ok {
		v, err := parseHexRune(tok)
		if err != nil {
			return nil
		}
		return []string{string(rune(v))}
	}
	lo, err1 := parseHexUint32(strings.TrimSpace(a))
	hi, err2 := parseHexUint32(strings.TrimSpace(b))
	if err1 != nil || err2 != nil || lo > hi {
		return nil
	}
	var out []string
	for cp := lo; cp <= hi; cp++ {
		if cp > 0x10FFFF || (cp >= 0xD800 && cp <= 0xDFFF) {
			continue
		}
		out = append(out, string(rune(cp)))
	}
	return out
}

func parseHexRune(s string) (rune, error) {
	u, err := parseHexUint32(s)
	if err != nil {
		return 0, err
	}
	if u > 0x10FFFF || (u >= 0xD800 && u <= 0xDFFF) {
		return 0, strconv.ErrRange
	}
	return rune(u), nil
}

func parseHexUint32(s string) (uint32, error) {
	s = strings.TrimSpace(s)
	if len(s) > 8 {
		return 0, errors.New("hex field too long")
	}
	v, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return 0, err
	}
	return uint32(v), nil
}
