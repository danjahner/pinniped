// Copyright 2020 the Pinniped contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	clientauthenticationv1beta1 "k8s.io/client-go/pkg/apis/clientauthentication/v1beta1"

	idpv1alpha1 "go.pinniped.dev/generated/1.19/apis/idp/v1alpha1"
	"go.pinniped.dev/internal/client"
	"go.pinniped.dev/internal/constable"
	"go.pinniped.dev/internal/here"
)

//nolint: gochecknoinits
func init() {
	rootCmd.AddCommand(newExchangeCredentialCmd(os.Args, os.Stdout, os.Stderr).cmd)
}

type exchangeCredentialCommand struct {
	// runFunc is called by the cobra.Command.Run hook. It is included here for
	// testability.
	runFunc func(stdout, stderr io.Writer)

	// cmd is the cobra.Command for this CLI command. It is included here for
	// testability.
	cmd *cobra.Command
}

func newExchangeCredentialCmd(args []string, stdout, stderr io.Writer) *exchangeCredentialCommand {
	c := &exchangeCredentialCommand{
		runFunc: runExchangeCredential,
	}

	c.cmd = &cobra.Command{
		Run: func(cmd *cobra.Command, _ []string) {
			c.runFunc(stdout, stderr)
		},
		Args:  cobra.NoArgs, // do not accept positional arguments for this command
		Use:   "exchange-credential",
		Short: "Exchange a credential for a cluster-specific access credential",
		Long: here.Doc(`
			Exchange a credential which proves your identity for a time-limited,
			cluster-specific access credential.

			Designed to be conveniently used as an credential plugin for kubectl.
			See the help message for 'pinniped get-kubeconfig' for more
			information about setting up a kubeconfig file using Pinniped.

			Requires all of the following environment variables, which are
			typically set in the kubeconfig:
			  - PINNIPED_TOKEN: the token to send to Pinniped for exchange
			  - PINNIPED_NAMESPACE: the namespace of the identity provider to authenticate
			    against
			  - PINNIPED_IDP_TYPE: the type of identity provider to authenticate
			    against (e.g., "webhook")
			  - PINNIPED_IDP_NAME: the name of the identity provider to authenticate
			    against
			  - PINNIPED_CA_BUNDLE: the CA bundle to trust when calling
				Pinniped's HTTPS endpoint
			  - PINNIPED_K8S_API_ENDPOINT: the URL for the Pinniped credential
				exchange API

			For more information about credential plugins in general, see
			https://kubernetes.io/docs/reference/access-authn-authz/authentication/#client-go-credential-plugins
		`),
	}

	c.cmd.SetArgs(args)
	c.cmd.SetOut(stdout)
	c.cmd.SetErr(stderr)

	return c
}

type envGetter func(string) (string, bool)
type tokenExchanger func(
	ctx context.Context,
	namespace string,
	idp corev1.TypedLocalObjectReference,
	token string,
	caBundle string,
	apiEndpoint string,
) (*clientauthenticationv1beta1.ExecCredential, error)

const (
	ErrMissingEnvVar  = constable.Error("failed to get credential: environment variable not set")
	ErrInvalidIDPType = constable.Error("invalid IDP type")
)

func runExchangeCredential(stdout, _ io.Writer) {
	err := exchangeCredential(os.LookupEnv, client.ExchangeToken, stdout, 30*time.Second)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}

func exchangeCredential(envGetter envGetter, tokenExchanger tokenExchanger, outputWriter io.Writer, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	namespace, varExists := envGetter("PINNIPED_NAMESPACE")
	if !varExists {
		return envVarNotSetError("PINNIPED_NAMESPACE")
	}

	idpType, varExists := envGetter("PINNIPED_IDP_TYPE")
	if !varExists {
		return envVarNotSetError("PINNIPED_IDP_TYPE")
	}

	idpName, varExists := envGetter("PINNIPED_IDP_NAME")
	if !varExists {
		return envVarNotSetError("PINNIPED_IDP_NAME")
	}

	token, varExists := envGetter("PINNIPED_TOKEN")
	if !varExists {
		return envVarNotSetError("PINNIPED_TOKEN")
	}

	caBundle, varExists := envGetter("PINNIPED_CA_BUNDLE")
	if !varExists {
		return envVarNotSetError("PINNIPED_CA_BUNDLE")
	}

	apiEndpoint, varExists := envGetter("PINNIPED_K8S_API_ENDPOINT")
	if !varExists {
		return envVarNotSetError("PINNIPED_K8S_API_ENDPOINT")
	}

	idp := corev1.TypedLocalObjectReference{Name: idpName}
	switch strings.ToLower(idpType) {
	case "webhook":
		idp.APIGroup = &idpv1alpha1.SchemeGroupVersion.Group
		idp.Kind = "WebhookIdentityProvider"
	default:
		return fmt.Errorf(`%w: %q, supported values are "webhook"`, ErrInvalidIDPType, idpType)
	}

	cred, err := tokenExchanger(ctx, namespace, idp, token, caBundle, apiEndpoint)
	if err != nil {
		return fmt.Errorf("failed to get credential: %w", err)
	}

	err = json.NewEncoder(outputWriter).Encode(cred)
	if err != nil {
		return fmt.Errorf("failed to marshal response to stdout: %w", err)
	}

	return nil
}

func envVarNotSetError(varName string) error {
	return fmt.Errorf("%w: %s", ErrMissingEnvVar, varName)
}
