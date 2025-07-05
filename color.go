package main

import (
	"fmt"
	"os"
	"strings"
)

// ANSI color codes
const (
	Reset  = "\033[0m"
	Bold   = "\033[1m"
	Dim    = "\033[2m"
	
	// Regular colors
	Black   = "\033[30m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
	Gray    = "\033[90m"
	
	// Bright colors
	BrightRed     = "\033[91m"
	BrightGreen   = "\033[92m"
	BrightYellow  = "\033[93m"
	BrightBlue    = "\033[94m"
	BrightMagenta = "\033[95m"
	BrightCyan    = "\033[96m"
	BrightWhite   = "\033[97m"
)

var colorEnabled = true

func init() {
	// Disable colors if NO_COLOR is set or if not in a terminal
	if os.Getenv("NO_COLOR") != "" {
		colorEnabled = false
		return
	}
	
	// Check if stdout is a terminal
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) == 0 {
		colorEnabled = false
	}
}

// Color functions
func color(code, text string) string {
	if !colorEnabled {
		return text
	}
	return code + text + Reset
}

func bold(text string) string {
	return color(Bold, text)
}

func dim(text string) string {
	return color(Dim, text)
}

func red(text string) string {
	return color(Red, text)
}

func green(text string) string {
	return color(Green, text)
}

func yellow(text string) string {
	return color(Yellow, text)
}

func blue(text string) string {
	return color(Blue, text)
}

func magenta(text string) string {
	return color(Magenta, text)
}

func cyan(text string) string {
	return color(Cyan, text)
}

func gray(text string) string {
	return color(Gray, text)
}

func brightRed(text string) string {
	return color(BrightRed, text)
}

func brightGreen(text string) string {
	return color(BrightGreen, text)
}

func brightYellow(text string) string {
	return color(BrightYellow, text)
}

func brightCyan(text string) string {
	return color(BrightCyan, text)
}

// Semantic color functions
func success(text string) string {
	return brightGreen(text)
}

func errorMsg(text string) string {
	return brightRed(text)
}

func warning(text string) string {
	return brightYellow(text)
}

func info(text string) string {
	return brightCyan(text)
}

func priority(p string) string {
	switch p {
	case "p1":
		return brightRed(bold("[P1]"))
	case "p2":
		return yellow("[P2]")
	case "p3":
		return blue("[P3]")
	default:
		return gray("[" + strings.ToUpper(p) + "]")
	}
}

func status(s string) string {
	switch s {
	case "done":
		return green("✓")
	case "open":
		return cyan("○")
	case "paused":
		return yellow("⏸")
	case "delegated":
		return blue("→")
	case "dropped":
		return gray("✗")
	default:
		return gray("?")
	}
}

func tag(t string) string {
	return magenta("#" + t)
}

func project(p string) string {
	return blue("@" + p)
}

func area(a string) string {
	return cyan(a)
}

func due(text string, overdue bool) string {
	if overdue {
		return brightRed(text)
	}
	return yellow(text)
}

func estimate(e int) string {
	return gray(fmt.Sprintf("~%d", e))
}

func index(i int) string {
	return bold(fmt.Sprintf("%3d.", i))
}

func filename(f string) string {
	return dim(f)
}

func date(d string) string {
	return gray(d)
}

func count(n int, label string) string {
	return fmt.Sprintf("%s %s", bold(fmt.Sprintf("%d", n)), label)
}