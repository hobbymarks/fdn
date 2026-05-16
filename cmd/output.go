package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"log/slog"

	"github.com/fatih/color"
	"github.com/mattn/go-runewidth"
	"golang.org/x/term"

	"github.com/hobbymarks/go-difflib/difflib"
)

type UserInput string

const (
	All  UserInput = "all"
	A    UserInput = "a"
	Yes  UserInput = "yes"
	Y    UserInput = "y"
	No   UserInput = "no"
	N    UserInput = "n"
	Quit UserInput = "quit"
	Q    UserInput = "q"
)

func GetConfirm() UserInput {
	var cmsg string

	fmt.Print("Please confirm (all,yes,no,quit):")
	fmt.Scan(&cmsg)

	return UserInput(strings.ToLower(cmsg))
}

func noEffectTip() {
	var tipsDivider string

	if term.IsTerminal(0) {
		tw, _, err := term.GetSize(0)
		if err != nil {
			slog.Error(err.Error())
			tipsDivider = strings.Repeat("*", 80)
		} else {
			tipsDivider = strings.Repeat("*", tw)
		}
		fmt.Println(tipsDivider)
		fmt.Println(
			"--> 'will to' ==> 'to',add flag '-i' or '-c' to take effect",
		)
	}
}

func OutputResult(
	origin string,
	processed string,
	inplace bool,
	fullpath bool,
) {
	if !fullpath {
		origin = filepath.Base(origin)
		processed = filepath.Base(processed)
	}
	if plainStyle {
		fmt.Println("   ", origin)
		if inplace {
			fmt.Println("==>", processed)
		} else {
			fmt.Println("-->", processed)
		}
	} else {
		origin = strings.ReplaceAll(origin, " ", "▯")
		processed = strings.ReplaceAll(processed, " ", "▯")

		a := strings.Split(origin, "")
		b := strings.Split(processed, "")

		seqm := difflib.NewMatcher(a, b)
		red := color.New(color.FgRed).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()
		richOrigin := ""
		richProcessed := ""
		sw := runewidth.StringWidth
		for _, opc := range seqm.GetOpCodes() {
			switch opc.Tag {
			case 'r':
				as := strings.Join(a[opc.I1:opc.I2], "")
				bs := strings.Join(b[opc.J1:opc.J2], "")
				slog.Debug("R:" + as + bs)
				if pretty {
					if sw(as) > sw(bs) {
						richOrigin += red(as)
						richProcessed += green(bs) + strings.Repeat(" ", sw(as)-sw(bs))
					} else if sw(as) < sw(bs) {
						richOrigin += red(as) + strings.Repeat(" ", sw(bs)-sw(as))
						richProcessed += green(bs)
					} else {
						richOrigin += red(as)
						richProcessed += green(bs)
					}
				} else {
					richOrigin += red(as)
					richProcessed += green(bs)
				}
			case 'd':
				as := strings.Join(a[opc.I1:opc.I2], "")
				slog.Debug("D:" + as)
				if pretty {
					richOrigin += red(as)
					richProcessed += strings.Repeat(" ", sw(as))
				} else {
					richOrigin += red(as)
				}
			case 'i':
				as := strings.Join(a[opc.I1:opc.I2], "")
				bs := strings.Join(b[opc.J1:opc.J2], "")
				slog.Debug("I:" + as + bs)
				if pretty {
					richOrigin += as + strings.Repeat(" ", sw(bs))
					richProcessed += green(bs)
				} else {
					richOrigin += as
					richProcessed += green(bs)
				}
			case 'e':
				as := strings.Join(a[opc.I1:opc.I2], "")
				bs := strings.Join(b[opc.J1:opc.J2], "")
				slog.Debug("E:" + as + bs)
				richOrigin += as
				richProcessed += bs
			}
		}
		fmt.Println("   ", richOrigin)
		if inplace {
			fmt.Println("==>", richProcessed)
		} else {
			fmt.Println("-->", richProcessed)
		}
	}
}
