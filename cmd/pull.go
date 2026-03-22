// SPDX-FileCopyrightText: 2026 Elwan Mayencourt <mayencourt@elwan.ch>
// SPDX-License-Identifier: MIT

package cmd

import (
	"fmt"

	appconfig "github.com/YungBricoCoop/l1/internal/config"
	"github.com/YungBricoCoop/l1/internal/storage"
	"github.com/YungBricoCoop/l1/internal/ui"
	"github.com/spf13/cobra"
)

func newPullCmd(opts *rootOptions) *cobra.Command {
	var bucket string
	var objectKey string

	pullCmd := &cobra.Command{
		Use:   "pull <file>",
		Short: "Download a file from S3",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := appconfig.ResolvePath(opts.configPath)
			if err != nil {
				return err
			}

			cfg, err := appconfig.Load(path)
			if err != nil {
				return err
			}

			validateErr := cfg.ValidateForPush(bucket)
			if validateErr != nil {
				return validateErr
			}

			logger := newS3Logger(ui.ShouldUseColor(cfg.UI.Color, cmd.OutOrStdout()))
			resolvedBucket := resolveTransferBucket(cfg.S3.DefaultBucket, bucket)
			resolvedKey := resolveTransferKey(args[0], objectKey)
			reporter := newS3ProgressReporter(cmd.OutOrStdout(), logger, "downloading", cfg.UI.Progress)

			fmt.Fprintln(
				cmd.OutOrStdout(),
				logger.Info(fmt.Sprintf("downloading s3://%s/%s -> %s", resolvedBucket, resolvedKey, args[0])),
			)

			pullErr := storage.PullFile(cmd.Context(), cfg, storage.PullInput{
				FilePath: args[0],
				Bucket:   resolvedBucket,
				Key:      resolvedKey,
				Progress: reporter.Callback,
			})
			if pullErr != nil {
				return pullErr
			}

			fmt.Fprintln(
				cmd.OutOrStdout(),
				logger.Success(fmt.Sprintf("downloaded s3://%s/%s to %s", resolvedBucket, resolvedKey, args[0])),
			)
			return nil
		},
	}

	pullCmd.Flags().StringVar(&bucket, "bucket", "", "override bucket for this download")
	pullCmd.Flags().StringVar(&objectKey, "key", "", "override object key (defaults to destination basename)")

	return pullCmd
}
