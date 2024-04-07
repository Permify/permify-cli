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
	return &cobra.Command{
		Use:              "configure",
		Short:            "configure permctl",
		PersistentPreRun: persistentPreRun,
		RunE:             runE,
	}
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

	url, err := tui.StringPrompt("enter permify url", "", config.CliConfig.PermifyURL)
	if err != nil {
		return err
	}

	resp, err := client.New(url)

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
	config.CliConfig.PermifyURL = url
	config.CliConfig.Tenant = tenantIds[tenant]
	err = config.Write()
	if err != nil {
		logger.Log.Error(err)
	}
	logger.Log.Info("successfully configured ", "config file", configFile)
	return nil
}
