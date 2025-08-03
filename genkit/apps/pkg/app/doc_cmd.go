package app

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"strings"
)

const (
	flagDoc          = "doc"
	flagDocShorthand = "d"
)

func docCommand(name string) *cobra.Command {
	return &cobra.Command{
		Use:   "doc",
		Short: "Generate markdown and man document of all commands.",
		Long:  "Generate markdown and man document of all commands.",

		Run: func(c *cobra.Command, args []string) {
			fmt.Printf("Generating command docs...\n")
		},
	}
}

// addDocFlag adds flags for a specific application to the specified FlagSet
// object.
// nolint: deadcode,unused,varcheck
func addDocFlag(name string, fs *pflag.FlagSet) {
	fs.BoolP(flagDoc, flagDocShorthand, false, fmt.Sprintf("Generate docs for %s.", color.GreenString(strings.Split(name, " ")[0])))
}

// addDocumentCommandFlag adds flags for a specific command of application to the
// specified FlagSet object.
func addDocumentCommandFlag(name string, fs *pflag.FlagSet) {
	fs.BoolP(
		flagDoc,
		flagDocShorthand,
		false,
		fmt.Sprintf("Generate docs for %s.", color.GreenString(strings.Split(name, " ")[0])),
	)
}
