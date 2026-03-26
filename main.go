// SPDX-FileCopyrightText: 2026 Elwan Mayencourt <mayencourt@elwan.ch>
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"

	"github.com/YungBricoCoop/l1/cmd"
	"github.com/YungBricoCoop/l1/internal/theme"
	"github.com/YungBricoCoop/l1/internal/ui"
)

const Version = "0.1.0"

func main() {
	if err := cmd.Execute(Version); err != nil {
		styles := ui.NewStyles(ui.ShouldUseColor(true, os.Stderr), theme.DefaultName())
		fmt.Fprintln(os.Stderr, styles.Error(fmt.Sprintf("error: %v", err)))
		os.Exit(1)
	}
}
