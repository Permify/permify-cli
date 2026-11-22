// Package cli initializes the permctl cli by mounting all commands
package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/Permify/permify-cli/core/config"
	"github.com/Permify/permify-cli/core/logger"
	"github.com/Permify/permify-cli/templates"
	"github.com/Permify/permify-cli/version"
	"github.com/spf13/cobra"
)

// Cli holds the bootstrapped permctl command
type Cli struct {
	Name              string
	ShortDescription  string
	DefaultConfigPath string
	Cmd               *cobra.Command
}

// New initializes a permctl root command with configuration
func New(name, shortDescription, defaultConfigPath string) *Cli {
	c := &Cli{
		Name:              name,
		ShortDescription:  shortDescription,
		DefaultConfigPath: defaultConfigPath,
	}
	c.Cmd = &cobra.Command{
		Use:              name,
		Short:            c.ShortDescription,
		PersistentPreRun: PersistentPreRun,
		Run:              c.Run,
		Version:          version.Version,
	}

	longDescription := templates.LongDescription("permctl", c.Cmd)
	examples := templates.Examples("permctl", c.Cmd)
	c.Cmd.Long = longDescription
	c.Cmd.Example = examples
	//disable help sub command
	c.Cmd.SetHelpCommand(&cobra.Command{Hidden: true})
	c.Cmd.AddCommand(ConfigureCmd())
	c.Cmd.PersistentFlags().Bool("debug", false, "verbose logging")
	c.Cmd.PersistentFlags().String("config", defaultConfigPath, fmt.Sprintf("%s config file", c.Name))
	c.Cmd.PersistentFlags().String("profile", "default", "profile name for config")
	c.Cmd.PersistentFlags().String("schema", "", "schema version to use")
	c.Cmd.PersistentFlags().String("endpoint", "", "permify gRPC endpoint")
	c.Cmd.PersistentFlags().String("token", "", "permify auth token")
	c.Cmd.PersistentFlags().String("cert-path", "", "path to TLS certificate file")
	c.Cmd.PersistentFlags().String("cert-key", "", "path to TLS certificate key file")

	return c
}

// PersistentPreRun defines the actions to be done before running the cli
func PersistentPreRun(cmd *cobra.Command, args []string) {
	debugEnabled, _ := cmd.Flags().GetBool("debug")
	os.Setenv(PermifyDebugEnv, fmt.Sprintf("%t", debugEnabled))
	logger.Update(debugEnabled)
	err := initializeConfig(cmd, args)
	if err != nil {
		logger.Log.Fatal(err)
	}
}

func initializeConfig(cmd *cobra.Command, _ []string) error {
	configFile, err := cmd.Flags().GetString("config")
	if err != nil {
		return err
	}
	profile, err := cmd.Flags().GetString("profile")
	if err != nil {
		return err
	}

	err = config.IsConfigured(configFile, profile)
	if err != nil {
		logger.Log.Error(err)
		logger.Log.Print("permctl is not configured. Please run `permctl configure`")
		os.Exit(1)
	}
	err = config.Load(configFile, profile)
	if err != nil {
		logger.Log.Fatal(err)
	}
	return applyRuntimeOverrides(cmd)
}

func applyRuntimeOverrides(cmd *cobra.Command) error {
	endpoint, token, certPath, certKey := config.GetCredentials()

	if flagEndpoint, err := cmd.Flags().GetString("endpoint"); err == nil && strings.TrimSpace(flagEndpoint) != "" {
		endpoint = strings.TrimSpace(flagEndpoint)
	}
	if flagToken, err := cmd.Flags().GetString("token"); err == nil && strings.TrimSpace(flagToken) != "" {
		token = strings.TrimSpace(flagToken)
	}
	if flagCertPath, err := cmd.Flags().GetString("cert-path"); err == nil && strings.TrimSpace(flagCertPath) != "" {
		certPath = strings.TrimSpace(flagCertPath)
	}
	if flagCertKey, err := cmd.Flags().GetString("cert-key"); err == nil && strings.TrimSpace(flagCertKey) != "" {
		certKey = strings.TrimSpace(flagCertKey)
	}

	config.CliConfig.PermifyURL = endpoint
	config.CliConfig.Token = token
	config.CliConfig.CertPath = certPath
	config.CliConfig.CertKey = certKey
	config.CliConfig.SslEnabled = config.ComputeSslEnabled(
		config.CliConfig.PermifyURL,
		config.CliConfig.CertPath,
		config.CliConfig.CertKey,
	)

	return config.CliConfig.Validate()
}

// Run - function to run on root cli command
func (c Cli) Run(cmd *cobra.Command, _ []string) {
	cmd.Help()
}

// Execute - run the root command
func (c Cli) Execute() {
	err := c.Cmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
