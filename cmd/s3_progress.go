// SPDX-FileCopyrightText: 2026 Elwan Mayencourt <mayencourt@elwan.ch>
// SPDX-License-Identifier: MIT

package cmd

import (
	"fmt"
	"io"

	"github.com/YungBricoCoop/l1/internal/storage"
	"github.com/YungBricoCoop/l1/internal/ui"
)

const (
	unknownProgressStepBytes int64 = 2 * 1024 * 1024
	progressPercentScale     int64 = 100
	progressStepPercent      int64 = 10
)

type s3ProgressReporter struct {
	out              io.Writer
	logger           ui.Logger
	action           string
	enabled          bool
	lastStep         int64
	lastBytes        int64
	lastPrintedBytes int64
	lastPrintedTotal int64
	hasPrinted       bool
	finished         bool
}

func newS3ProgressReporter(out io.Writer, logger ui.Logger, action string, enabled bool) *s3ProgressReporter {
	return &s3ProgressReporter{
		out:      out,
		logger:   logger,
		action:   action,
		enabled:  enabled,
		lastStep: -1,
	}
}

func (r *s3ProgressReporter) Callback(update storage.ProgressUpdate) {
	if !r.enabled || r.finished {
		return
	}

	if update.Done {
		r.finished = true
		r.print(update.BytesTransferred, update.TotalBytes)
		return
	}

	if update.TotalBytes > 0 {
		percent := (update.BytesTransferred * progressPercentScale) / update.TotalBytes
		step := percent / progressStepPercent
		if step <= r.lastStep {
			return
		}
		r.lastStep = step
		r.print(update.BytesTransferred, update.TotalBytes)
		return
	}

	if update.BytesTransferred-r.lastBytes < unknownProgressStepBytes {
		return
	}

	r.lastBytes = update.BytesTransferred
	r.print(update.BytesTransferred, update.TotalBytes)
}

func (r *s3ProgressReporter) print(bytes, total int64) {
	if r.hasPrinted && r.lastPrintedBytes == bytes && r.lastPrintedTotal == total {
		return
	}

	r.hasPrinted = true
	r.lastPrintedBytes = bytes
	r.lastPrintedTotal = total
	fmt.Fprintln(r.out, r.logger.Progress(r.action, bytes, total))
}
