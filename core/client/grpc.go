// Package client handles the permify client to connect with the server.
//
// The New function accepts an endpoint argument but also consults the local
// credential store (see core/credentials). Any value left blank in the
// argument falls back to the stored value, so callers can keep calling
// client.New(config.CliConfig.PermifyURL) unchanged while still benefiting
// from token authentication and mTLS configured via `permctl login`.
package client

import (
	"errors"
	"fmt"
	"os"

	"github.com/Permify/permify-cli/core/credentials"
	"github.com/Permify/permify-cli/core/logger"
	permify "github.com/Permify/permify-go/v1"
	"google.golang.org/grpc"
	grpccreds "google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// profileEnv lets the active profile be overridden without plumbing it
// through every call site. Falls back to credentials.DefaultProfile.
const profileEnv = "PERMIFY_PROFILE"

// New initializes a new permify client. The endpoint argument takes
// precedence; if empty, the endpoint stored under the active profile is
// used. Stored token and mTLS material (if present) are always applied.
//
// New returns an error when:
//   - no endpoint is supplied either by argument or by storage
//   - stored mTLS certificate material is referenced but unreadable
//   - the underlying grpc.Dial fails
func New(endpoint string) (*permify.Client, error) {
	profile := os.Getenv(profileEnv)
	stored, err := credentials.Load(profile)
	if err != nil && !errors.Is(err, credentials.ErrNotFound) {
		// A real error (insecure perms, corrupt YAML, ...) is surfaced.
		// ErrNotFound is intentionally swallowed so the CLI keeps working
		// without a credential file.
		return nil, fmt.Errorf("load stored credentials: %w", err)
	}
	if stored == nil {
		stored = &credentials.Credentials{}
	}

	if endpoint == "" {
		endpoint = stored.Endpoint
	}
	if endpoint == "" {
		return nil, errors.New("permify endpoint is empty: pass one to client.New or run `permctl login --endpoint <host:port>`")
	}

	dialOpts, err := buildDialOptions(stored)
	if err != nil {
		return nil, err
	}

	logger.Log.Debug("dialing permify", "endpoint", endpoint, "tls", stored.CertPath != "", "token", stored.Token != "")

	return permify.NewClient(
		permify.Config{Endpoint: endpoint},
		dialOpts...,
	)
}

// buildDialOptions translates stored credentials into grpc.DialOptions:
//
//   - cert_path + cert_key  => mTLS transport credentials
//   - cert_path only        => TLS with server cert as CA (rare but supported)
//   - neither               => insecure transport (preserves prior behavior)
//   - token                 => per-RPC bearer token; secure flavor over TLS,
//     non-secure flavor over plaintext
func buildDialOptions(c *credentials.Credentials) ([]grpc.DialOption, error) {
	var opts []grpc.DialOption

	tlsConfigured := c.CertPath != "" || c.CertKey != ""

	if tlsConfigured {
		if c.CertPath == "" || c.CertKey == "" {
			return nil, errors.New("incomplete TLS credentials: both cert_path and cert_key must be set")
		}
		creds, err := grpccreds.NewClientTLSFromFile(c.CertPath, "")
		if err != nil {
			return nil, fmt.Errorf("load TLS credentials from %s: %w", c.CertPath, err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	if c.Token != "" {
		md := map[string]string{"authorization": "Bearer " + c.Token}
		if tlsConfigured {
			opts = append(opts, grpc.WithPerRPCCredentials(secureTokenCredentials(md)))
		} else {
			opts = append(opts, grpc.WithPerRPCCredentials(nonSecureTokenCredentials(md)))
		}
	}

	return opts, nil
}
