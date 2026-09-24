// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package main

import "time"

var (
	// version is the application version injected by the build.
	version = "dev"

	// commit is the source revision injected by the build.
	commit = "unknown"

	// buildTime is the UTC time when the binary was built.
	buildTime = time.Unix(0, 0)

	// repositoryURL is the source repository URL shown by version information.
	repositoryURL = "https://github.com/woozymasta/a2s"

	// _buildTime is the RFC3339 build timestamp injected by the linker.
	_buildTime string
)

func init() {
	if _buildTime == "" {
		return
	}

	value, err := time.Parse(time.RFC3339, _buildTime)
	if err == nil {
		buildTime = value.UTC()
	}
}
