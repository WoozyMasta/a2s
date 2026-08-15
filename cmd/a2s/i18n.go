package main

import (
	"embed"
	"fmt"

	"github.com/woozymasta/flags"
)

// applicationCatalogFS contains all application-localized messages.
//
// Catalogs are embedded so the CLI does not depend on files beside the binary.
//
//go:embed i18n/*.json
var applicationCatalogFS embed.FS

// newI18nConfig loads the embedded application catalog
// and configures English as the final fallback after automatic locale detection.
func newI18nConfig() (flags.I18nConfig, error) {
	catalog, err := flags.NewJSONCatalogDirFS(applicationCatalogFS, "i18n")
	if err != nil {
		return flags.I18nConfig{}, fmt.Errorf("load embedded i18n catalogs: %w", err)
	}

	return flags.I18nConfig{
		UserCatalog:     catalog,
		FallbackLocales: []string{"en"},
	}, nil
}
