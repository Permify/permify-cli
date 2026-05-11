package permission

import (
	"os"

	"github.com/Permify/permify-cli/core/client"
	v1 "github.com/Permify/permify-go/generated/base/v1"
	"github.com/charmbracelet/log"
)

func Client() v1.PermissionClient {
	c, err := client.New()
	if err != nil {
		log.Error("Error initializing permify client. Check the configuration or rerun `permify configure`")
		os.Exit(-1)
	}
	return c.Permission
}
