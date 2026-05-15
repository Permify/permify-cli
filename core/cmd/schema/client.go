package schema

import (
	"os"

	"github.com/Permify/permify-cli/core/client"
	v1 "github.com/Permify/permify-go/generated/base/v1"
	"github.com/charmbracelet/log"
)

func Client() v1.SchemaClient {
	c, err := client.NewFromCLIConfig()
	if err != nil {
		log.Error("Error initializing permify client. Check the configuration or rerun `permify configure`")
		os.Exit(-1)
	}
	return c.Schema
}
