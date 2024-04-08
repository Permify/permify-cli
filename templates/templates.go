// Package templates holds the help templates for the cli
package templates

import (
	"bytes"
	"embed"
	"os"
	"path"
	"text/template"

	"github.com/Permify/permify-cli/core/logger"
	"github.com/charmbracelet/glamour"
	"github.com/spf13/cobra"
)

//go:embed *
var TempFs embed.FS //assign the variable TempFs to embed.FS, FS means (File System)
var customStyle = `
{
	"strong": {
		"bold": true,
		"color": "1"
	}
}`

func getTemplate(dir, filename string, cmd *cobra.Command) (string, error) {
	templatePath := path.Join(dir, filename+".md")
	tpl, err := template.ParseFS(TempFs, templatePath)
	if err != nil {
		return "", err
	}

	var b bytes.Buffer
	err = tpl.Execute(&b, cmd)
	if err != nil {
		return "", err
	}
	termRender, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithEmoji(),
		glamour.WithPreservedNewLines(),
		glamour.WithStylesFromJSONBytes([]byte(customStyle)),
		glamour.WithWordWrap(160),
	)
	if err != nil {
		return "", err
	}

	color, err := termRender.RenderBytes(b.Bytes())
	if err != nil {
		return "", err
	}
	return string(color), nil
}

func LongDescription(dir string, cmd *cobra.Command) string {
	longDescriptionstring, err := getTemplate(dir, "long_description", cmd)
	if err != nil {
		logger.Log.Error("failed to get long description for", "dir", dir)
		logger.Log.Fatal(err)

	}
	return longDescriptionstring
}

func Examples(dir string, cmd *cobra.Command) string {
	examples, err := getTemplate(dir, "examples", cmd)
	if err != nil {
		logger.Log.Error("failed to get examples for", "dir", dir)
		logger.Log.Fatal(err)
	}
	return examples
}

func apiGetTemplate(cmd *cobra.Command, filename string) (string, error) {
	apiType := cmd.Parent().Name()
	action := cmd.Name()
	templatePath := path.Join("permify", apiType, action, filename+".md")
	_, err := TempFs.ReadFile(templatePath)
	if os.IsNotExist(err) {
		return "", err
	}
	tpl, err := template.ParseFS(TempFs, templatePath)
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	err = tpl.Execute(&b, cmd)
	if err != nil {
		return "", err
	}
	termRender, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithEmoji(),
		glamour.WithPreservedNewLines(),
		glamour.WithStylesFromJSONBytes([]byte(customStyle)),
		glamour.WithWordWrap(160),
	)
	if err != nil {
		return "", err
	}
	color, err := termRender.RenderBytes(b.Bytes())
	if err != nil {
		return "", err
	}
	return string(color), nil
}

func GetDescription(cmd *cobra.Command) (string, error) {
	return apiGetTemplate(cmd, "long_description")
}

func GetExamples(cmd *cobra.Command) (string, error) {
	return apiGetTemplate(cmd, "examples")
}
