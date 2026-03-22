// SPDX-FileCopyrightText: 2026 Elwan Mayencourt <mayencourt@elwan.ch>
// SPDX-License-Identifier: MIT

package ui

import (
	"fmt"
	"io"
	"os"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/term"
)

// Styles defines global message styling shared across all commands.
type Styles struct {
	color bool

	infoStyle     lipgloss.Style
	successStyle  lipgloss.Style
	errorStyle    lipgloss.Style
	progressStyle lipgloss.Style
}

// Tag defines a command-specific label style, such as [s3] or [config].
type Tag struct {
	Name  string
	Style lipgloss.Style
}

// Logger formats tagged command output using global styles plus a command tag.
type Logger struct {
	styles Styles
	tag    Tag
}

const (
	uiPercentScale int64 = 100
	uiKiBBase      int64 = 1024
)

func NewStyles(color bool) Styles {
	return Styles{
		color:         color,
		infoStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("#D0D7DE")),
		successStyle:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#3FB950")),
		errorStyle:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F85149")),
		progressStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#58A6FF")),
	}
}

func NewTag(name string, style lipgloss.Style) Tag {
	return Tag{Name: name, Style: style}
}

func NewLogger(styles Styles, tag Tag) Logger {
	return Logger{styles: styles, tag: tag}
}

func ShouldUseColor(configEnabled bool, out io.Writer) bool {
	if !configEnabled {
		return false
	}

	file, ok := out.(*os.File)
	if !ok {
		return false
	}

	return term.IsTerminal(file.Fd())
}

func (s Styles) Info(message string) string {
	return s.renderMessage(s.infoStyle, message)
}

func (s Styles) Success(message string) string {
	return s.renderMessage(s.successStyle, message)
}

func (s Styles) Error(message string) string {
	return s.renderMessage(s.errorStyle, message)
}

func (s Styles) Progress(action string, currentBytes, totalBytes int64) string {
	if totalBytes > 0 {
		percent := (currentBytes * uiPercentScale) / totalBytes
		percent = min(percent, uiPercentScale)

		message := fmt.Sprintf("%s %3d%% (%s/%s)", action, percent, humanBytes(currentBytes), humanBytes(totalBytes))
		return s.renderMessage(s.progressStyle, message)
	}

	message := fmt.Sprintf("%s %s", action, humanBytes(currentBytes))
	return s.renderMessage(s.progressStyle, message)
}

func (l Logger) Info(message string) string {
	return l.withTag(l.styles.Info(message))
}

func (l Logger) Success(message string) string {
	return l.withTag(l.styles.Success(message))
}

func (l Logger) Error(message string) string {
	return l.withTag(l.styles.Error(message))
}

func (l Logger) Progress(action string, currentBytes, totalBytes int64) string {
	return l.withTag(l.styles.Progress(action, currentBytes, totalBytes))
}

func (l Logger) withTag(message string) string {
	tag := l.renderTag()
	if tag == "" {
		return message
	}

	return fmt.Sprintf("%s %s", tag, message)
}

func (l Logger) renderTag() string {
	if l.tag.Name == "" {
		return ""
	}

	label := fmt.Sprintf("[%s]", l.tag.Name)
	if !l.styles.color {
		return label
	}

	return l.tag.Style.Render(label)
}

func (s Styles) renderMessage(style lipgloss.Style, value string) string {
	if !s.color {
		return value
	}

	return style.Render(value)
}

func humanBytes(bytes int64) string {
	if bytes < uiKiBBase {
		return fmt.Sprintf("%dB", bytes)
	}

	units := []string{"KB", "MB", "GB", "TB"}
	value := float64(bytes)
	unitIndex := -1
	for value >= float64(uiKiBBase) && unitIndex+1 < len(units) {
		value /= float64(uiKiBBase)
		unitIndex++
	}

	if unitIndex < 0 {
		return fmt.Sprintf("%dB", bytes)
	}

	return fmt.Sprintf("%.1f%s", value, units[unitIndex])
}
