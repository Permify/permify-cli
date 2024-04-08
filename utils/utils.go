// Package utils provides helper functions for permctl
package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/Permify/permify-cli/core/logger"
	"github.com/Permify/permify-cli/templates"
	"github.com/Permify/permify-cli/tui"
	v1 "github.com/Permify/permify-go/generated/base/v1"
	"github.com/alecthomas/chroma/quick"
	"github.com/spf13/cobra"
)

// PrettyPrint readable json to terminal
func PrettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	err := quick.Highlight(os.Stdout, string(s), "json", "terminal256", "monokai")
	if err != nil {
		logger.Log.Error("failed to highlight", "error", err)
	}
	return string(s)
}

// CheckIfUnknownSubcommand halts execution on passing unknown subcommand. Which is not usually an error
func CheckIfUnknownSubcommand(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		return
	}
	parents := []string{}
	cmd.VisitParents(func(parent *cobra.Command) {
		if parent != nil && parent.Name() != "" {
			parents = append(parents, parent.Name())
		}
	})
	parents = reverse(parents)
	parents = append(parents, cmd.Name())
	msg := fmt.Sprintf(`Error: unknown command "%s" for "%s" `, args[0], strings.Join(parents, " "))

	err := tui.Critical(msg)
	if suggestions := cmd.SuggestionsFor(args[0]); len(suggestions) > 0 {
		err += "\n\nDid you mean this?\n"
		for _, s := range suggestions {
			err += fmt.Sprintf("\t%v\n", s)
		}
	}
	err += "\n\n"
	helpMsg := tui.Blue(fmt.Sprintf("'%s --help'", strings.Join(parents, " ")))
	err += fmt.Sprintf("Run %s for usage. \n", helpMsg)

	fmt.Println(err)
	fmt.Println()
	os.Exit(1)
}

func reverse(s []string) []string {
	var reversed []string
	for i := len(s) - 1; i >= 0; i-- {
		reversed = append(reversed, s[i])
	}
	return reversed
}

func trimTrailingWhitespaces(s string) string {
	return strings.TrimRight(s, " \t\n")
}

// CmdHelp shows help templates on commands
func CmdHelp(cmd *cobra.Command, _ []string) {
	longDescription, err := templates.GetDescription(cmd)
	if err != nil {
		logger.Log.Error("failed to get long description")
		logger.Log.Fatal(err)
	}
	examples, err := templates.GetExamples(cmd)
	if err != nil {
		logger.Log.Error("failed to get examples")
		logger.Log.Fatal(err)
	}
	cmd.Example = examples
	cmd.Long = longDescription
	t := cmd.HelpTemplate()
	tmpl, err := template.New("help").Funcs(template.FuncMap{
		"trimTrailingWhitespaces": trimTrailingWhitespaces,
	}).Parse(t)
	if err != nil {
		logger.Log.Error("Parsing HelpTemplate failed")
		logger.Log.Fatal(err)
	}
	err = tmpl.Execute(cmd.OutOrStdout(), cmd)
	if err != nil {
		logger.Log.Error("Executing HelpTemplate failed")
		logger.Log.Fatal(err)
	}
}

func ParseEntity(entityStr string) (*v1.Entity, error) {
	entityRegex := regexp.MustCompile(`^[\w\d]+:[\w\d]+$`)
	if !entityRegex.MatchString(entityStr) {
		return nil, errors.New("Entity string should match pattern <type>:<id>")
	}
	entity := strings.Split(entityStr, ":")
	return &v1.Entity{
		Type: entity[0],
		Id: entity[1],
	}, nil
}

func ParseSubject(subjectStr string) (*v1.Subject, error) {
	subjectRegex := regexp.MustCompile(`^([\w\d]+):([\w\d]+)(?:#([\w\d]+))?$`)
	if !subjectRegex.MatchString(subjectStr) {
		return nil, errors.New("Subject string should match pattern <type>:<id>#relation (relation is optional)")
	}
	subject := strings.FieldsFunc(subjectStr, func(r rune) bool {
        return r == ':' || r == '#'
    })
	if len(subject) == 3 {
		return &v1.Subject{
			Type: subject[0],
			Id: subject[1],
			Relation: subject[2],
		}, nil
	}
	return &v1.Subject{
		Type: subject[0],
		Id: subject[1],
	}, nil
}

func ReadFileToString(filePath string) (string, error) {
	if !strings.HasSuffix(filepath.Base(filePath), ".perm") {
        return "", fmt.Errorf("only perm schema files accepted")
    }
    data, err := os.ReadFile(filePath)
    if err != nil {
        return "", err
    }
    fileContents := string(data)
    return fileContents, nil
}