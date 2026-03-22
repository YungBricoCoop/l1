// SPDX-FileCopyrightText: 2026 Elwan Mayencourt <mayencourt@elwan.ch>
// SPDX-License-Identifier: MIT

package cmd

import (
	"charm.land/lipgloss/v2"
	"github.com/YungBricoCoop/l1/internal/ui"
)

func newS3Logger(color bool) ui.Logger {
	tagStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFF5F5")).
		Background(lipgloss.Color("#832be7")).
		Padding(0, 1)

	return ui.NewLogger(ui.NewStyles(color), ui.NewTag("s3", tagStyle))
}

func newConfigLogger(color bool) ui.Logger {
	tagStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#111111")).
		Padding(0, 1)

	return ui.NewLogger(ui.NewStyles(color), ui.NewTag("config", tagStyle))
}
