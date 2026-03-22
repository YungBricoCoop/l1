// SPDX-FileCopyrightText: 2026 Elwan Mayencourt <mayencourt@elwan.ch>
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"
)

const Version = "0.1.0"

func main() {
	fmt.Fprintf(os.Stdout, "mycli version %s\n", Version)
}
