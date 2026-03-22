// SPDX-FileCopyrightText: 2026 Elwan Mayencourt <mayencourt@elwan.ch>
// SPDX-License-Identifier: MIT

package cmd

import (
	"path/filepath"
	"strings"
)

func resolveTransferBucket(defaultBucket, override string) string {
	bucket := strings.TrimSpace(override)
	if bucket != "" {
		return bucket
	}

	return strings.TrimSpace(defaultBucket)
}

func resolveTransferKey(filePath, override string) string {
	key := strings.TrimSpace(override)
	if key != "" {
		return key
	}

	return filepath.Base(filePath)
}
