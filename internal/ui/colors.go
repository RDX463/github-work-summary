package ui

import (
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

const (
	ansiReset   = "\x1b[0m"
	ansiBold    = "\x1b[1m"
	ansiDim     = "\x1b[2m"
	ansiRed     = "\x1b[31m"
	ansiGreen   = "\x1b[32m"
	ansiYellow  = "\x1b[33m"
	ansiBlue    = "\x1b[34m"
	ansiMagenta = "\x1b[35m"
	ansiCyan    = "\x1b[36m"
	ansiGray    = "\x1b[90m"
)

// ColorEnabled reports whether ANSI colors should be rendered for writer w.
//
// Rules:
// - GWS_COLOR=always enables colors.
// - GWS_COLOR=never disables colors.
// - NO_COLOR disables colors.
// - Otherwise colors are enabled only for terminal outputs.
func ColorEnabled(w io.Writer) bool {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("GWS_COLOR")))
	switch mode {
	case "always", "on", "1", "true":
		return true
	case "never", "off", "0", "false":
		return false
	}
	if strings.TrimSpace(os.Getenv("NO_COLOR")) != "" {
		return false
	}

	type fdWriter interface {
		Fd() uintptr
	}
	file, ok := w.(fdWriter)
	if !ok {
		return false
	}
	return term.IsTerminal(int(file.Fd()))
}

func style(w io.Writer, code, text string) string {
	if text == "" || !ColorEnabled(w) {
		return text
	}
	return code + text + ansiReset
}

func Bold(w io.Writer, text string) string    { return style(w, ansiBold, text) }
func Dim(w io.Writer, text string) string     { return style(w, ansiDim, text) }
func Red(w io.Writer, text string) string     { return style(w, ansiRed, text) }
func Green(w io.Writer, text string) string   { return style(w, ansiGreen, text) }
func Yellow(w io.Writer, text string) string  { return style(w, ansiYellow, text) }
func Blue(w io.Writer, text string) string    { return style(w, ansiBlue, text) }
func Magenta(w io.Writer, text string) string { return style(w, ansiMagenta, text) }
func Cyan(w io.Writer, text string) string    { return style(w, ansiCyan, text) }
func Gray(w io.Writer, text string) string    { return style(w, ansiGray, text) }
