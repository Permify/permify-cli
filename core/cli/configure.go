package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/Permify/permify-cli/core/client"
	"github.com/Permify/permify-cli/core/config"
	"github.com/Permify/permify-cli/core/logger"
	"github.com/Permify/permify-cli/tui"
	v1 "github.com/Permify/permify-go/generated/base/v1"
	"github.com/spf13/cobra"
)

// ConfigureCmd provides the configure command on permctl
func ConfigureCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "configure",
		Short:            "configure permctl",
		PersistentPreRun: persistentPreRun,
		RunE:             runE,
	}
	cmd.Flags().String("url", "", "permify server URL or endpoint")
	cmd.Flags().String("endpoint", "", "alias for --url")
	cmd.Flags().String("tenant", "", "default tenant ID")
	cmd.Flags().String("token", "", "bearer authentication token")
	cmd.Flags().String("cert-path", "", "path to TLS certificate or CA bundle")
	cmd.Flags().String("cert-key", "", "path to TLS private key")
	cmd.Flags().Bool("insecure", false, "disable TLS verification")
	return cmd
}

func persistentPreRun(cmd *cobra.Command, args []string) {
	err := persistentPreRunE(cmd, args)
	if err != nil {
		logger.Log.Error(err)
		os.Exit(1)
	}
}

func persistentPreRunE(cmd *cobra.Command, args []string) error {
	err := validateFlags(cmd, args)
	if err != nil {
		return err
	}
	debugEnabled, _ := cmd.Flags().GetBool("debug")
	logger.Update(debugEnabled)
	configFile, err := cmd.Flags().GetString("config")
	if err != nil {
		return err
	}
	profile, err := cmd.Flags().GetString("profile")
	if err != nil {
		return err
	}
	_, err = os.Stat(configFile)
	if err != nil {
		logger.Log.Debug("Initializing new config ", "path", configFile)
		err = config.New(configFile, profile)
		if err != nil {
			logger.Log.Error("Error initializing", "error", err)
			os.Exit(0)
		}
	} else {
		logger.Log.Info("Updating existing config", "path", configFile)
		return config.Load(configFile, profile)
	}
	return nil
}

func validateFlags(cmd *cobra.Command, args []string) error {
	cmd.DisableFlagParsing = false
	err := cmd.ParseFlags(args)
	// if error, print error and help info
	if err != nil {
		fmt.Println(err)
		cmd.Help()
		fmt.Println()
		os.Exit(0)
	}
	// if flag is for help show help message
	if cmd.Flag("help").Value.String() == "true" {
		cmd.Help()
		fmt.Println()
		os.Exit(0)
	}
	err = cmd.ValidateFlagGroups()
	if err != nil {
		return err
	}

	err = cmd.ValidateRequiredFlags()
	if err != nil {
		return err
	}

	err = cmd.ValidateArgs(args)
	if err != nil {
		return err
	}
	return err
}

func runE(cmd *cobra.Command, _ []string) error {
	configFile, _ := cmd.Flags().GetString("config")

	url, _ := cmd.Flags().GetString("url")
	if url == "" {
		url, _ = cmd.Flags().GetString("endpoint")
	}
	if url == "" {
		var err error
		url, err = tui.StringPrompt("enter permify url", "", config.CliConfig.GetEndpoint())
		if err != nil {
			return err
		}
	}

	token, _ := cmd.Flags().GetString("token")
	certPath, _ := cmd.Flags().GetString("cert-path")
	certKey, _ := cmd.Flags().GetString("cert-key")
	insecure, _ := cmd.Flags().GetBool("insecure")
	tenantFlag, _ := cmd.Flags().GetString("tenant")

	config.CliConfig.SetEndpoint(url)
	if token != "" {
		config.CliConfig.Token = token
	}
	if certPath != "" {
		config.CliConfig.CertPath = certPath
	}
	if certKey != "" {
		config.CliConfig.CertKey = certKey
	}
	if insecure {
		config.CliConfig.Insecure = true
	}

	// Validate credentials accessibility before connecting or saving
	if err := config.CliConfig.ValidateCredentials(); err != nil {
		logger.Log.Error(err)
		return err
	}

	// If tenant was passed directly via flag, use it; otherwise discover tenants via API
	if tenantFlag != "" {
		config.CliConfig.Tenant = tenantFlag
	} else {
		resp, err := client.New(url)
		if err != nil {
			logger.Log.Error("failed to create client: ", err)
			return err
		}

		// Todo: Implement pagination
		tenants, err := resp.Tenancy.List(context.Background(), &v1.TenantListRequest{})
		if err != nil {
			logger.Log.Fatal(err)
		}

		tenantNames := []string{}
		tenantIds := map[string]string{}
		for _, tenant := range tenants.Tenants {
			nameID := fmt.Sprintf("%s {%s}", tenant.Name, tenant.Id)
			tenantNames = append(tenantNames, nameID)
			tenantIds[nameID] = tenant.Id
		}

		tenant, err := tui.Choice("Select a tenant: ", tenantNames)
		if err != nil {
			logger.Log.Error(err)
		}
		config.CliConfig.Tenant = tenantIds[tenant]
	}

	err := config.Write()
	if err != nil {
		logger.Log.Error(err)
		return err
	}
	logger.Log.Info("successfully configured ", "config file", configFile)
	return nil
}
