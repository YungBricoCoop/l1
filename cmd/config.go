// SPDX-FileCopyrightText: 2026 Elwan Mayencourt <mayencourt@elwan.ch>
// SPDX-License-Identifier: MIT

package cmd

import (
	"errors"
	"fmt"

	appconfig "github.com/YungBricoCoop/l1/internal/config"
	"github.com/YungBricoCoop/l1/internal/ui"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
)

const configSetArgCount = 2

func newConfigCmd(opts *rootOptions) *cobra.Command {
	var showPath bool
	var show bool

	configCmd := &cobra.Command{
		Use:   "config [key] [value]",
		Short: "Show or mutate l1 configuration",
		Long:  "Use 'l1 config --show' to display configuration or 'l1 config s3.url https://example' to set a value.",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := appconfig.ResolvePath(opts.configPath)
			if err != nil {
				return err
			}

			if showPath {
				fmt.Fprintln(cmd.OutOrStdout(), path)
				return nil
			}

			if show || len(args) == 0 {
				if len(args) > 0 {
					return errors.New("--show cannot be combined with key/value arguments")
				}
				return runConfigShow(cmd, path)
			}

			if len(args) != configSetArgCount {
				return errors.New("expected either --show/--show-path or <key> <value>")
			}

			return runConfigSet(cmd, path, args[0], args[1])
		},
	}

	configCmd.Flags().BoolVar(&showPath, "show-path", false, "print resolved config file path")
	configCmd.Flags().BoolVar(&show, "show", false, "print current config as TOML")
	configCmd.AddCommand(newConfigInitCmd(opts))
	configCmd.AddCommand(newConfigBackupCmd(opts))
	configCmd.AddCommand(newConfigBackupRestoreCmd(opts))

	return configCmd
}

func newConfigInitCmd(opts *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Create a default config.toml at the resolved config path",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, err := appconfig.ResolvePath(opts.configPath)
			if err != nil {
				return err
			}

			created, err := appconfig.InitFile(path)
			if err != nil {
				return err
			}

			logger := configCommandLogger(cmd, path)

			if created {
				fmt.Fprintln(cmd.OutOrStdout(), logger.Success(fmt.Sprintf("created config at %s", path)))
				return nil
			}

			fmt.Fprintln(cmd.OutOrStdout(), logger.Info(fmt.Sprintf("config already exists at %s", path)))
			return nil
		},
	}
}

func newConfigBackupCmd(opts *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "bkp",
		Short: "Backup config.toml to config.toml.bkp (overwrite if it exists)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, err := appconfig.ResolvePath(opts.configPath)
			if err != nil {
				return err
			}

			backupPath, err := appconfig.BackupFile(path)
			if err != nil {
				return err
			}

			logger := configCommandLogger(cmd, path)
			fmt.Fprintln(cmd.OutOrStdout(), logger.Success(fmt.Sprintf("backup written to %s", backupPath)))
			return nil
		},
	}
}

func newConfigBackupRestoreCmd(opts *rootOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "bkp-restore",
		Short: "Restore config.toml from config.toml.bkp",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, err := appconfig.ResolvePath(opts.configPath)
			if err != nil {
				return err
			}

			backupPath, err := appconfig.RestoreBackupFile(path)
			if err != nil {
				return err
			}

			logger := configCommandLogger(cmd, path)
			fmt.Fprintln(cmd.OutOrStdout(), logger.Success(fmt.Sprintf("config restored from %s", backupPath)))
			return nil
		},
	}
}

func runConfigShow(cmd *cobra.Command, path string) error {
	cfg, err := appconfig.LoadForEdit(path)
	if err != nil {
		return err
	}

	encoded, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), string(encoded))
	return nil
}

func runConfigSet(cmd *cobra.Command, path, key, rawValue string) error {
	cfg, err := appconfig.LoadForEdit(path)
	if err != nil {
		return err
	}

	parsedValue, err := appconfig.ParseValueForKey(key, rawValue)
	if err != nil {
		return err
	}

	setErr := appconfig.SetConfigValue(&cfg, key, parsedValue)
	if setErr != nil {
		return setErr
	}

	logger := newConfigLogger(ui.ShouldUseColor(cfg.UI.Color, cmd.OutOrStdout()), cfg.UI.Theme)

	validateErr := cfg.Validate()
	if validateErr != nil {
		return validateErr
	}

	saveErr := appconfig.Save(path, cfg)
	if saveErr != nil {
		return saveErr
	}

	fmt.Fprintln(cmd.OutOrStdout(), logger.Success(fmt.Sprintf("updated %s in %s", key, path)))
	return nil
}

func configCommandLogger(cmd *cobra.Command, path string) ui.Logger {
	cfg, err := appconfig.LoadForEdit(path)
	if err != nil {
		defaultCfg := appconfig.DefaultConfig()
		return newConfigLogger(ui.ShouldUseColor(defaultCfg.UI.Color, cmd.OutOrStdout()), defaultCfg.UI.Theme)
	}

	return newConfigLogger(ui.ShouldUseColor(cfg.UI.Color, cmd.OutOrStdout()), cfg.UI.Theme)
}
