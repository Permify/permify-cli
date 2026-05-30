package permission

import (
	"os"

	"github.com/Permify/permify-cli/core/client"
	"github.com/Permify/permify-cli/core/config"
	v1 "github.com/Permify/permify-go/generated/base/v1"
	"github.com/charmbracelet/log"
)

func Client() v1.PermissionClient {
	opts := client.Options{
		Token:       config.CliConfig.Token,
		CertPath:    config.CliConfig.CertPath,
		CertKeyPath: config.CliConfig.CertKeyPath,
	}
	c, err := client.NewFromEnv(config.CliConfig.PermifyURL, opts)
	if err != nil {
		log.Error("Error initializing permify client. Check the configuration or rerun `permify configure`")
		os.Exit(-1)
	}
	return c.Permission
}
