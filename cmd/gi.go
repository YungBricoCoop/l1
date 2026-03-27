// SPDX-FileCopyrightText: 2026 Elwan Mayencourt <mayencourt@elwan.ch>
// SPDX-License-Identifier: MIT

package cmd

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	appconfig "github.com/YungBricoCoop/l1/internal/config"
	"github.com/YungBricoCoop/l1/internal/ui"
	"github.com/spf13/cobra"
)

const (
	gitignoreAPIBaseURL  = "https://www.toptal.com/developers/gitignore/api"
	gitignoreListArg     = "list"
	gitignoreHTTPTimeout = 30 * time.Second
)

func newGitignoreCmd(opts *rootOptions) *cobra.Command {
	var list bool
	var outputPath string

	giCmd := &cobra.Command{
		Use:     "gi [template...]",
		Aliases: []string{"gitignore"},
		Short:   "Generate .gitignore templates from gitignore.io",
		Long: "Fetch templates from gitignore.io. Example: 'l1 gi p m v' expands to " +
			"'python,macos,visualstudiocode'.",
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGitignoreGenerate(cmd, opts, args, list, outputPath)
		},
	}

	giCmd.Flags().BoolVar(&list, "list", false, "print available gitignore templates")
	giCmd.Flags().StringVarP(&outputPath, "output", "o", ".gitignore", "output file path for generated gitignore")

	return giCmd
}

func runGitignoreGenerate(
	cmd *cobra.Command,
	opts *rootOptions,
	args []string,
	list bool,
	outputPath string,
) error {
	listMode, err := shouldRunGitignoreList(list, args)
	if err != nil {
		return err
	}
	if listMode {
		return runGitignoreList(cmd)
	}

	templateArgs, err := resolveGitignoreTemplateArgs(args, opts)
	if err != nil {
		return err
	}

	templates := normalizeGitignoreTemplates(templateArgs)
	if len(templates) == 0 {
		return errors.New("no valid templates provided")
	}

	content, err := fetchGitignoreURL(cmd, buildGitignoreTemplateURL(templates))
	if err != nil {
		return err
	}

	writeErr := writeGitignoreFile(outputPath, content)
	if writeErr != nil {
		return writeErr
	}

	logger := gitignoreCommandLogger(cmd, opts)
	fmt.Fprintln(cmd.OutOrStdout(), logger.Success(fmt.Sprintf("wrote %s", outputPath)))
	return nil
}

func shouldRunGitignoreList(list bool, args []string) (bool, error) {
	if list {
		if len(args) > 0 {
			return false, errors.New("--list cannot be combined with template arguments")
		}
		return true, nil
	}

	if len(args) == 1 && strings.EqualFold(args[0], gitignoreListArg) {
		return true, nil
	}

	return false, nil
}

func resolveGitignoreTemplateArgs(args []string, opts *rootOptions) ([]string, error) {
	if len(args) > 0 {
		return args, nil
	}

	defaultTemplates, err := gitignoreDefaultTemplates(opts)
	if err != nil {
		return nil, err
	}
	if len(defaultTemplates) == 0 {
		return nil, errors.New("expected at least one template, 'list', or configured gi.templates")
	}

	return defaultTemplates, nil
}

func runGitignoreList(cmd *cobra.Command) error {
	content, err := fetchGitignoreURL(cmd, buildGitignoreListURL())
	if err != nil {
		return err
	}

	fmt.Fprint(cmd.OutOrStdout(), content)
	if !strings.HasSuffix(content, "\n") {
		fmt.Fprintln(cmd.OutOrStdout())
	}

	return nil
}

func normalizeGitignoreTemplates(args []string) []string {
	templates := make([]string, 0, len(args))
	seen := make(map[string]struct{}, len(args))

	for _, rawArg := range args {
		for rawTemplate := range strings.SplitSeq(rawArg, ",") {
			template := strings.ToLower(strings.TrimSpace(rawTemplate))
			if template == "" {
				continue
			}

			if len(template) == 1 {
				template = expandGitignoreShortcut(template)
			}

			if _, exists := seen[template]; exists {
				continue
			}

			seen[template] = struct{}{}
			templates = append(templates, template)
		}
	}

	return templates
}

func expandGitignoreShortcut(template string) string {
	switch template {
	case "g":
		return "go"
	case "m":
		return "macos"
	case "p":
		return "python"
	case "r":
		return "rust"
	case "v":
		return "visualstudiocode"
	default:
		return template
	}
}

func buildGitignoreTemplateURL(templates []string) string {
	return fmt.Sprintf("%s/%s", gitignoreAPIBaseURL, strings.Join(templates, ","))
}

func buildGitignoreListURL() string {
	return fmt.Sprintf("%s/%s", gitignoreAPIBaseURL, gitignoreListArg)
}

func fetchGitignoreURL(cmd *cobra.Command, url string) (string, error) {
	req, err := http.NewRequestWithContext(cmd.Context(), http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}

	client := &http.Client{Timeout: gitignoreHTTPTimeout}
	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request gitignore api: %w", err)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gitignore api returned %s: %s", res.Status, strings.TrimSpace(string(body)))
	}

	return string(body), nil
}

func writeGitignoreFile(path, content string) error {
	if strings.TrimSpace(path) == "" {
		return errors.New("output path cannot be empty")
	}

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	return nil
}

func gitignoreCommandLogger(cmd *cobra.Command, opts *rootOptions) ui.Logger {
	path, err := appconfig.ResolvePath(opts.configPath)
	if err != nil {
		defaultCfg := appconfig.DefaultConfig()
		return newGitignoreLogger(ui.ShouldUseColor(defaultCfg.UI.Color, cmd.OutOrStdout()), defaultCfg.UI.Theme)
	}

	cfg, err := appconfig.LoadForEdit(path)
	if err != nil {
		defaultCfg := appconfig.DefaultConfig()
		return newGitignoreLogger(ui.ShouldUseColor(defaultCfg.UI.Color, cmd.OutOrStdout()), defaultCfg.UI.Theme)
	}

	return newGitignoreLogger(ui.ShouldUseColor(cfg.UI.Color, cmd.OutOrStdout()), cfg.UI.Theme)
}

func gitignoreDefaultTemplates(opts *rootOptions) ([]string, error) {
	path, err := appconfig.ResolvePath(opts.configPath)
	if err != nil {
		return nil, err
	}

	cfg, err := appconfig.LoadForEdit(path)
	if err != nil {
		return nil, err
	}

	return cfg.GI.Templates, nil
}
