// Copyright 2020 the Pinniped contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"
)

//nolint: gochecknoglobals
var loginCmd = &cobra.Command{
	Use:          "login",
	Short:        "login",
	Long:         "Login to a Pinniped server",
	SilenceUsage: true, // do not print usage message when commands fail
}

//nolint: gochecknoinits
func init() {
	rootCmd.AddCommand(loginCmd)
}
