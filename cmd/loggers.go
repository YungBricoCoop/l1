// SPDX-FileCopyrightText: 2026 Elwan Mayencourt <mayencourt@elwan.ch>
// SPDX-License-Identifier: MIT

package cmd

import (
	"github.com/YungBricoCoop/l1/internal/ui"
)

func newS3Logger(color bool, themeName string) ui.Logger {
	styles := ui.NewStyles(color, themeName)
	return ui.NewLogger(styles, ui.NewTag("s3", styles.S3TagStyle()))
}

func newConfigLogger(color bool, themeName string) ui.Logger {
	styles := ui.NewStyles(color, themeName)
	return ui.NewLogger(styles, ui.NewTag("config", styles.ConfigTagStyle()))
}

func newGitignoreLogger(color bool, themeName string) ui.Logger {
	styles := ui.NewStyles(color, themeName)
	return ui.NewLogger(styles, ui.NewTag("gi", styles.ConfigTagStyle()))
}
