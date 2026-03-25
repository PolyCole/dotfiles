package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

const (
	colorCyan  = "\033[36m"
	colorReset = "\033[0m"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: startup-message <messages-file>")
		os.Exit(1)
	}

	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "error reading file:", err)
		os.Exit(1)
	}

	messages := parseMessages(string(data))
	if len(messages) == 0 {
		return
	}

	width := termWidth()
	msg := messages[rand.Intn(len(messages))]
	lines := wordWrap(msg, width/2)
	printCentered(lines, width)
}

// parseMessages splits on blank lines. If no blank lines exist, each line is its own message.
func parseMessages(data string) []string {
	if strings.Contains(strings.TrimSpace(data), "\n\n") {
		return parseBlocks(data)
	}
	var messages []string
	for _, line := range strings.Split(data, "\n") {
		if t := strings.TrimSpace(line); t != "" {
			messages = append(messages, t)
		}
	}
	return messages
}

func parseBlocks(data string) []string {
	var messages []string
	var current strings.Builder
	for _, line := range strings.Split(data, "\n") {
		if strings.TrimSpace(line) == "" {
			if current.Len() > 0 {
				messages = append(messages, current.String())
				current.Reset()
			}
		} else {
			if current.Len() > 0 {
				current.WriteString(" ")
			}
			current.WriteString(strings.TrimSpace(line))
		}
	}
	if current.Len() > 0 {
		messages = append(messages, current.String())
	}
	return messages
}

func termWidth() int {
	if cols := os.Getenv("COLUMNS"); cols != "" {
		if w, err := strconv.Atoi(cols); err == nil && w > 0 {
			return w
		}
	}
	return 80
}

func wordWrap(text string, width int) []string {
	var lines []string
	var current strings.Builder

	for _, word := range strings.Fields(text) {
		if current.Len() == 0 {
			current.WriteString(word)
		} else if current.Len()+1+len(word) <= width {
			current.WriteString(" ")
			current.WriteString(word)
		} else {
			lines = append(lines, current.String())
			current.Reset()
			current.WriteString(word)
		}
	}
	if current.Len() > 0 {
		lines = append(lines, current.String())
	}
	return lines
}

func printCentered(lines []string, width int) {
	for _, line := range lines {
		padding := (width - len(line)) / 2
		if padding < 0 {
			padding = 0
		}
		fmt.Printf("%s%s%s%s\n", strings.Repeat(" ", padding), colorCyan, line, colorReset)
	}
}
