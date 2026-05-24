package cli

import (
	"errors"
	"fmt"

	"github.com/Permify/permify-cli/core/credentials"
	"github.com/Permify/permify-cli/core/logger"
	"github.com/spf13/cobra"
)

// authPersistentPreRun is a stripped-down replacement for the root
// PersistentPreRun: it sets up logging based on --debug but skips the
// permctl config-file check so login/logout/whoami remain usable before
// `permctl configure` has ever been run.
func authPersistentPreRun(cmd *cobra.Command, _ []string) {
	debugEnabled, _ := cmd.Flags().GetBool("debug")
	logger.Update(debugEnabled)
}

// LoginCmd persists endpoint / token / mTLS material under the active
// profile. It writes to the credentials file with 0600 permissions; only
// the fields supplied on the command line are updated, so the command can
// also be used to rotate a single value (e.g. `permctl login --token new`).
func LoginCmd() *cobra.Command {
	var (
		endpoint string
		token    string
		certPath string
		certKey  string
	)
	cmd := &cobra.Command{
		Use:   "login",
		Short: "store permify credentials (endpoint, token, mTLS cert/key)",
		Long: `Persist credentials used by permctl when talking to a Permify server.

Credentials are stored as YAML at $PERMIFY_CREDENTIALS_FILE if set,
otherwise at $HOME/.permctl-credentials.yaml, with 0600 file permissions.
Entries are keyed by profile (default: "default").`,
		// Override the root PersistentPreRun: login must work before a
		// permctl config file exists.
		PersistentPreRun: authPersistentPreRun,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if endpoint == "" && token == "" && certPath == "" && certKey == "" {
				return errors.New("at least one of --endpoint, --token, --cert-path, --cert-key must be provided")
			}
			profile, _ := cmd.Flags().GetString("profile")

			// Merge with any existing entry so single-flag rotation works.
			existing, err := credentials.Load(profile)
			if err != nil && !errors.Is(err, credentials.ErrNotFound) {
				return err
			}
			if existing == nil {
				existing = &credentials.Credentials{}
			}
			if endpoint != "" {
				existing.Endpoint = endpoint
			}
			if token != "" {
				existing.Token = token
			}
			if certPath != "" {
				existing.CertPath = certPath
			}
			if certKey != "" {
				existing.CertKey = certKey
			}

			if err := credentials.Save(profile, existing); err != nil {
				return err
			}
			path, _ := credentials.GetCredentialsPath()
			logger.Log.Info("credentials stored", "profile", profile, "file", path)
			return nil
		},
	}
	cmd.Flags().StringVar(&endpoint, "endpoint", "", "permify gRPC endpoint (host:port)")
	cmd.Flags().StringVar(&token, "token", "", "bearer token for authenticated calls")
	cmd.Flags().StringVar(&certPath, "cert-path", "", "path to client TLS certificate (PEM)")
	cmd.Flags().StringVar(&certKey, "cert-key", "", "path to client TLS key (PEM)")
	return cmd
}

// LogoutCmd deletes stored credentials for the active profile. If no other
// profiles remain in the file, the file itself is removed.
func LogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:              "logout",
		Short:            "remove stored permify credentials for the active profile",
		PersistentPreRun: authPersistentPreRun,
		RunE: func(cmd *cobra.Command, _ []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			if err := credentials.Delete(profile); err != nil {
				return err
			}
			logger.Log.Info("credentials removed", "profile", profile)
			return nil
		},
	}
}

// WhoamiCmd prints the stored credentials for the active profile, masking
// the token so the value is safe to share in screenshots/bug reports.
func WhoamiCmd() *cobra.Command {
	return &cobra.Command{
		Use:              "whoami",
		Short:            "show stored permify credentials for the active profile",
		PersistentPreRun: authPersistentPreRun,
		RunE: func(cmd *cobra.Command, _ []string) error {
			profile, _ := cmd.Flags().GetString("profile")
			creds, err := credentials.Load(profile)
			if err != nil {
				if errors.Is(err, credentials.ErrNotFound) {
					path, _ := credentials.GetCredentialsPath()
					return fmt.Errorf("no credentials stored for profile %q (file: %s) - run `permctl login` to set them", profile, path)
				}
				return err
			}
			path, _ := credentials.GetCredentialsPath()
			fmt.Printf("profile:   %s\n", profile)
			fmt.Printf("file:      %s\n", path)
			fmt.Printf("endpoint:  %s\n", display(creds.Endpoint))
			fmt.Printf("token:     %s\n", maskToken(creds.Token))
			fmt.Printf("cert_path: %s\n", display(creds.CertPath))
			fmt.Printf("cert_key:  %s\n", display(creds.CertKey))
			return nil
		},
	}
}

// display returns "(unset)" for empty values so the whoami output makes the
// difference between unset and intentional empty obvious.
func display(s string) string {
	if s == "" {
		return "(unset)"
	}
	return s
}

// maskToken keeps the first 4 and last 4 characters of long tokens and
// replaces the middle with asterisks; short tokens are fully masked so they
// cannot be reconstructed from the output.
func maskToken(t string) string {
	if t == "" {
		return "(unset)"
	}
	if len(t) <= 8 {
		return "****"
	}
	return t[:4] + "..." + t[len(t)-4:]
}
