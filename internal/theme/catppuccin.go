// SPDX-FileCopyrightText: 2026 Elwan Mayencourt <mayencourt@elwan.ch>
// SPDX-License-Identifier: MIT

package theme

import (
	"fmt"
	"strings"

	catppuccin "github.com/catppuccin/go"
)

const (
	CatppuccinMocha     = "catppuccin-mocha"
	CatppuccinFrappe    = "catppuccin-frappe"
	CatppuccinMacchiato = "catppuccin-macchiato"
	CatppuccinLatte     = "catppuccin-latte"
)

func DefaultName() string {
	return CatppuccinMocha
}

func SupportedNames() []string {
	return []string{
		CatppuccinMocha,
		CatppuccinFrappe,
		CatppuccinMacchiato,
		CatppuccinLatte,
	}
}

func CanonicalName(name string) (string, error) {
	resolvedTheme, err := resolveTheme(name)
	if err != nil {
		return "", fmt.Errorf(
			"unsupported ui.theme %q: expected one of %s",
			name,
			strings.Join(SupportedNames(), ", "),
		)
	}

	return resolvedTheme.name, nil
}

func FlavourByName(name string) (catppuccin.Flavour, error) {
	resolvedTheme, err := resolveTheme(name)
	if err != nil {
		return nil, err
	}

	return resolvedTheme.flavour, nil
}

type resolvedTheme struct {
	name    string
	flavour catppuccin.Flavour
}

func resolveTheme(name string) (resolvedTheme, error) {
	canonicalName := strings.ToLower(strings.TrimSpace(name))

	switch canonicalName {
	case CatppuccinMocha:
		return resolvedTheme{name: CatppuccinMocha, flavour: catppuccin.Mocha}, nil
	case CatppuccinFrappe:
		return resolvedTheme{name: CatppuccinFrappe, flavour: catppuccin.Frappe}, nil
	case CatppuccinMacchiato:
		return resolvedTheme{name: CatppuccinMacchiato, flavour: catppuccin.Macchiato}, nil
	case CatppuccinLatte:
		return resolvedTheme{name: CatppuccinLatte, flavour: catppuccin.Latte}, nil
	default:
		return resolvedTheme{}, fmt.Errorf("unsupported ui.theme %q", name)
	}
}
