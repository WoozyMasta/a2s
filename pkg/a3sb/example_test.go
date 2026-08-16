// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/a2s

package a3sb_test

import (
	"fmt"

	"github.com/woozymasta/a2s/pkg/a2s"
	"github.com/woozymasta/a2s/pkg/a3sb"
)

func ExampleClient() {
	baseClient, err := a2s.New("127.0.0.1", 2303)
	if err != nil {
		panic(err)
	}
	defer baseClient.Close()

	client := &a3sb.Client{Client: baseClient}
	fmt.Println(client.Addr())
	// Output: 127.0.0.1:2303
}
