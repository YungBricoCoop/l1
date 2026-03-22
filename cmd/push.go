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

func newPushCmd(opts *rootOptions) *cobra.Command {
	var bucket string
	var objectKey string

	pushCmd := &cobra.Command{
		Use:   "push <file>",
		Short: "Upload a file to S3",
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
			reporter := newS3ProgressReporter(cmd.OutOrStdout(), logger, "uploading", cfg.UI.Progress)

			fmt.Fprintln(
				cmd.OutOrStdout(),
				logger.Info(fmt.Sprintf("uploading %s -> s3://%s/%s", args[0], resolvedBucket, resolvedKey)),
			)

			etag, err := storage.PushFile(cmd.Context(), cfg, storage.PushInput{
				FilePath: args[0],
				Bucket:   resolvedBucket,
				Key:      resolvedKey,
				Progress: reporter.Callback,
			})
			if err != nil {
				return err
			}

			if etag == "" {
				fmt.Fprintln(
					cmd.OutOrStdout(),
					logger.Success(fmt.Sprintf("uploaded %s to s3://%s/%s", args[0], resolvedBucket, resolvedKey)),
				)
				return nil
			}

			fmt.Fprintln(
				cmd.OutOrStdout(),
				logger.Success(
					fmt.Sprintf("uploaded %s to s3://%s/%s (etag: %s)", args[0], resolvedBucket, resolvedKey, etag),
				),
			)
			return nil
		},
	}

	pushCmd.Flags().StringVar(&bucket, "bucket", "", "override bucket for this upload")
	pushCmd.Flags().StringVar(&objectKey, "key", "", "override object key (defaults to file basename)")

	return pushCmd
}
