// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/a2s/pkg/a3sb"
	"github.com/woozymasta/flags"
)

func TestEmbeddedI18nCatalogsHaveIdenticalKeys(t *testing.T) {
	var english map[string]string
	data, err := applicationCatalogFS.ReadFile("i18n/en.json")
	if err != nil {
		t.Fatalf("ReadFile(en.json) error = %v", err)
	}
	if err := json.Unmarshal(data, &english); err != nil {
		t.Fatalf("Unmarshal(en.json) error = %v", err)
	}

	for _, locale := range []string{"ru", "de", "it", "es", "cs", "zh"} {
		t.Run(locale, func(t *testing.T) {
			data, err := applicationCatalogFS.ReadFile("i18n/" + locale + ".json")
			if err != nil {
				t.Fatalf("ReadFile(%s.json) error = %v", locale, err)
			}
			var catalog map[string]string
			if err := json.Unmarshal(data, &catalog); err != nil {
				t.Fatalf("Unmarshal(%s.json) error = %v", locale, err)
			}
			if len(catalog) != len(english) {
				t.Errorf("catalog has %d keys, English catalog has %d", len(catalog), len(english))
			}
			for key := range english {
				if _, ok := catalog[key]; !ok {
					t.Errorf("catalog is missing key %q", key)
				}
			}
			for key := range catalog {
				if _, ok := english[key]; !ok {
					t.Errorf("catalog has unexpected key %q", key)
				}
			}
		})
	}
}

func TestCommandHelpLocalizesApplicationMetadata(t *testing.T) {
	tests := map[string]struct {
		command string
		option  string
		arg     string
	}{
		"en": {
			"Query server metadata with A2S_INFO.",
			"Output format",
			"Server host (with optional port, e.g., 127.0.0.1:27016)",
		},
		"ru": {
			"Запросить метаданные сервера через A2S_INFO.",
			"Формат вывода",
			"Хост сервера (порт можно указать через двоеточие, например 127.0.0.1:27016)",
		},
		"de": {
			"Servermetadaten mit A2S_INFO abfragen.",
			"Ausgabeformat",
			"Server-Host (mit optionalem Port, z. B. 127.0.0.1:27016)",
		},
		"it": {
			"Interroga i metadati del server con A2S_INFO.",
			"Formato di output",
			"Host del server (con porta opzionale, ad es. 127.0.0.1:27016)",
		},
		"es": {
			"Consultar los metadatos del servidor con A2S_INFO.",
			"Formato de salida",
			"Host del servidor (con puerto opcional, p. ej. 127.0.0.1:27016)",
		},
		"cs": {
			"Dotázat se na metadata serveru pomocí A2S_INFO.",
			"Výstupní formát",
			"Hostitel serveru (s volitelným portem, např. 127.0.0.1:27016)",
		},
		"zh": {
			"使用 A2S_INFO 查询服务器元数据。",
			"输出格式",
			"服务器主机（可选端口，例如 127.0.0.1:27016）",
		},
	}

	for locale, want := range tests {
		t.Run(locale, func(t *testing.T) {
			config, err := newI18nConfig()
			if err != nil {
				t.Fatalf("newI18nConfig() error = %v", err)
			}
			config.Locale = locale

			parser, err := newParser(&Options{}, config)
			if err != nil {
				t.Fatalf("newParser() error = %v", err)
			}
			_, err = parser.ParseArgs([]string{"info", "--help"})
			if !isParserError(err, flags.ErrHelp) {
				t.Fatalf("help parse error = %v, want ErrHelp", err)
			}
			help := strings.Join(strings.Fields(err.Error()), " ")
			if !strings.Contains(help, strings.Join(strings.Fields(want.command), " ")) {
				t.Fatalf("help does not contain %q: %v", want.command, err)
			}
			if !strings.Contains(help, strings.Join(strings.Fields(want.option), " ")) {
				t.Fatalf("help does not contain %q: %v", want.option, err)
			}
			if !strings.Contains(help, strings.Join(strings.Fields(want.arg), " ")) {
				t.Fatalf("help does not contain %q: %v", want.arg, err)
			}
		})
	}
}

func TestUnsupportedLocaleUsesEnglishCommandHelp(t *testing.T) {
	config, err := newI18nConfig()
	if err != nil {
		t.Fatalf("newI18nConfig() error = %v", err)
	}
	config.Locale = "xx-YY"

	parser, err := newParser(&Options{}, config)
	if err != nil {
		t.Fatalf("newParser() error = %v", err)
	}
	_, err = parser.ParseArgs([]string{"info", "--help"})
	if !isParserError(err, flags.ErrHelp) {
		t.Fatalf("help parse error = %v, want ErrHelp", err)
	}
	if !strings.Contains(err.Error(), "Query server metadata with A2S_INFO.") {
		t.Fatalf("unsupported locale did not use English command help: %v", err)
	}
}

func TestInfoTableLocalizesLabelsWithoutChangingValues(t *testing.T) {
	tests := map[string]string{
		"en": "Server name:",
		"ru": "Имя сервера:",
		"de": "Servername:",
	}

	for locale, wantLabel := range tests {
		t.Run(locale, func(t *testing.T) {
			var output bytes.Buffer
			app := localizedTestApplication(t, locale, &output, &output)
			info := &a2s.Info{
				Format:     a2s.InfoFormat(a2s.ResponseInfo),
				Name:       "Server value Ω",
				Players:    3,
				MaxPlayers: 10,
			}

			if err := renderInfoTable(
				app,
				info,
				a2s.QueryMeta{Duration: 25 * time.Millisecond},
				"127.0.0.1:27015",
				"table",
				NewFormatter("table", &output, app.Localizer),
			); err != nil {
				t.Fatalf("renderInfoTable() error = %v", err)
			}

			if !strings.Contains(output.String(), wantLabel) {
				t.Fatalf("output = %q, want localized label %q", output.String(), wantLabel)
			}
			if !strings.Contains(output.String(), "Server value Ω") {
				t.Fatalf("output = %q, server value was changed", output.String())
			}
		})
	}
}

func TestInfoTableLocalizesBooleanValuesAndDayZLabels(t *testing.T) {
	var output bytes.Buffer
	app := localizedTestApplication(t, "ru", &output, &output)
	gameID := uint64(221100)
	info := &a2s.Info{
		Format:   a2s.InfoFormat(a2s.ResponseInfo),
		GameID:   &gameID,
		EDF:      a2s.EDF(0x20),
		Keywords: []string{"privHive"},
	}

	if err := renderInfoTable(
		app,
		info,
		a2s.QueryMeta{},
		"127.0.0.1:27015",
		"table",
		NewFormatter("table", &output, app.Localizer),
	); err != nil {
		t.Fatalf("renderInfoTable() error = %v", err)
	}

	text := output.String()
	if !strings.Contains(text, "Приватная БД:") {
		t.Fatalf("output = %q, want localized private hive label", text)
	}
	if !strings.Contains(text, "Время отклика:") {
		t.Fatalf("output = %q, want localized response time label", text)
	}
	if !strings.Contains(text, "да") || !strings.Contains(text, "нет") {
		t.Fatalf("output = %q, want localized boolean values", text)
	}
	if strings.Contains(text, "true") || strings.Contains(text, "false") {
		t.Fatalf("output = %q, found unlocalized boolean value", text)
	}
}

func TestPlayersTableLocalizesLabelsAndEmptyMessage(t *testing.T) {
	var output bytes.Buffer
	app := localizedTestApplication(t, "ru", &output, &output)
	players := []a2s.Player{{Name: "Player value", Score: 7, Index: 1}}

	if err := renderPlayersTable(
		app,
		players,
		"127.0.0.1:27015",
		NewFormatter("table", &output, app.Localizer),
	); err != nil {
		t.Fatalf("renderPlayersTable() error = %v", err)
	}
	if !strings.Contains(strings.ToUpper(output.String()), strings.ToUpper("Имя")) ||
		!strings.Contains(output.String(), "Player value") {
		t.Fatalf("output = %q, player labels or value missing", output.String())
	}

	output.Reset()
	if err := renderPlayersTable(
		app,
		nil,
		"127.0.0.1:27015",
		NewFormatter("table", &output, app.Localizer),
	); err != nil {
		t.Fatalf("renderPlayersTable(empty) error = %v", err)
	}
	if !strings.Contains(output.String(), "Сервер пуст") {
		t.Fatalf("empty output = %q, want localized empty message", output.String())
	}
}

func TestRulesAndA3SBSectionsLocalizeLabelsWithoutChangingRuleValues(t *testing.T) {
	var output bytes.Buffer
	app := localizedTestApplication(t, "ru", &output, &output)
	if err := renderRulesTable(
		app,
		map[string]any{"enabled": true, "hostname": "Rule value Ω"},
		"127.0.0.1:27015",
		NewFormatter("table", &output, app.Localizer),
	); err != nil {
		t.Fatalf("renderRulesTable() error = %v", err)
	}
	if !strings.Contains(strings.ToUpper(output.String()), strings.ToUpper("Правило")) ||
		!strings.Contains(output.String(), "Rule value Ω") ||
		!strings.Contains(output.String(), "да") {
		t.Fatalf("rules output = %q, labels or value missing", output.String())
	}

	output.Reset()
	if err := renderA3SBRules(
		app,
		&a3sb.Rules{
			Version:    3,
			Island:     "Altis",
			Difficulty: &a3sb.Difficulty{},
		},
		NewFormatter("md", &output, app.Localizer),
		"127.0.0.1:27015",
	); err != nil {
		t.Fatalf("renderA3SBRules() error = %v", err)
	}
	if !strings.Contains(output.String(), "Информация о сервере") ||
		!strings.Contains(output.String(), "Настройки сложности") ||
		!strings.Contains(output.String(), "Карта:") ||
		!strings.Contains(output.String(), "Altis") {
		t.Fatalf("A3SB output = %q, localized sections or value missing", output.String())
	}
}

func TestRuntimeErrorsLocalizeContextAndPreserveCause(t *testing.T) {
	for _, locale := range []string{"ru", "de", "cs", "zh"} {
		t.Run(locale, func(t *testing.T) {
			app := localizedTestApplication(t, locale, &bytes.Buffer{}, &bytes.Buffer{})
			err := friendlyQueryError(
				app,
				"error.server_info",
				"failed to get server info",
				context.DeadlineExceeded,
				3*time.Second,
			)
			if errors.Is(err, context.DeadlineExceeded) == false {
				t.Fatal("localized query error lost its cause")
			}
			if strings.Contains(err.Error(), "context deadline exceeded") {
				t.Fatalf("error = %q, low-level timeout leaked", err)
			}
			if strings.Contains(err.Error(), "failed to get server info") {
				t.Fatalf("error = %q, English context was not localized", err)
			}
		})
	}
}

func TestRuntimeFailureKeepsMachineOutputClean(t *testing.T) {
	var stdout, stderr bytes.Buffer
	app := localizedTestApplication(t, "ru", &stdout, &stderr)
	if err := executeInfo(app, &InfoCommand{}, ClientOptions{}); err == nil {
		t.Fatal("executeInfo() error = nil, want client creation error")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty output on failure", stdout.String())
	}
}

func TestMachineReadableOutputIsLocaleIndependent(t *testing.T) {
	info := &a2s.Info{Name: "Server value", Keywords: []string{"rule", "value"}}
	rules := map[string]any{"hostname": "Rule value", "players": 2}

	var baselineInfo, baselineRules []byte
	for _, locale := range []string{"en", "ru", "de", "it", "es", "cs", "zh"} {
		t.Run(locale, func(t *testing.T) {
			app := localizedTestApplication(t, locale, &bytes.Buffer{}, &bytes.Buffer{})
			var infoOutput, rulesOutput bytes.Buffer
			if err := printInfoJSON(
				info,
				NewFormatter("json", &infoOutput, app.Localizer),
			); err != nil {
				t.Fatalf("printInfoJSON() error = %v", err)
			}
			if err := printRules(
				app,
				rules,
				nil,
				NewFormatter("json", &rulesOutput, app.Localizer),
			); err != nil {
				t.Fatalf("printRules() error = %v", err)
			}

			if locale == "en" {
				baselineInfo = infoOutput.Bytes()
				baselineRules = rulesOutput.Bytes()
				return
			}
			if !bytes.Equal(infoOutput.Bytes(), baselineInfo) {
				t.Fatalf("JSON info differs from English output: %s", infoOutput.String())
			}
			if !bytes.Equal(rulesOutput.Bytes(), baselineRules) {
				t.Fatalf("JSON rules differ from English output: %s", rulesOutput.String())
			}
		})
	}
}

func localizedTestApplication(t *testing.T, locale string, out, errOut *bytes.Buffer) *Application {
	t.Helper()
	config, err := newI18nConfig()
	if err != nil {
		t.Fatalf("newI18nConfig() error = %v", err)
	}
	config.Locale = locale
	app := NewApplication(out, errOut)
	app.Localizer = flags.NewLocalizer(config)
	return app
}
