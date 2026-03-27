// SPDX-FileCopyrightText: 2026 Elwan Mayencourt <mayencourt@elwan.ch>
// SPDX-License-Identifier: MIT

package cmd

import "github.com/spf13/cobra"

type rootOptions struct {
	configPath string
}

func Execute(version string) error {
	return newRootCmd(version).Execute()
}

func newRootCmd(version string) *cobra.Command {
	opts := &rootOptions{}

	rootCmd := &cobra.Command{
		Use:           "l1",
		Short:         "l1 is a personal CLI for common workflows",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version,
	}

	rootCmd.PersistentFlags().StringVar(
		&opts.configPath,
		"config",
		"",
		"override config file path",
	)

	rootCmd.AddCommand(newPushCmd(opts))
	rootCmd.AddCommand(newPullCmd(opts))
	rootCmd.AddCommand(newConfigCmd(opts))
	rootCmd.AddCommand(newGitignoreCmd(opts))

	return rootCmd
}
