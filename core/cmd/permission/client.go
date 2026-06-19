package permission

import (
	"os"

	"github.com/Permify/permify-cli/core/client"
	"github.com/Permify/permify-cli/core/config"
	v1 "github.com/Permify/permify-go/generated/base/v1"
	"github.com/charmbracelet/log"
)

func Client() v1.PermissionClient {
	token, certPath, certKey := config.GetCredentials()
	creds := client.Credentials{
		Token:    token,
		CertPath: certPath,
		CertKey:  certKey,
	}
	c, err := client.New(config.CliConfig.PermifyURL, creds)
	if err != nil {
		log.Error("Error initializing permify client. Check the configuration or rerun `permify configure`")
		os.Exit(-1)	
	}
	return c.Permission
}