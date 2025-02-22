package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pell-emulator/version"
)

var verbose bool

// VersionCmd ...
var VersionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v"},
	Short:   "Show version info",
	Run: func(cmd *cobra.Command, args []string) {
		coreVsersion := version.CoreSemVer
		if version.GitCommitHash != "" {
			coreVsersion += "+" + version.GitCommitHash
		}

		if verbose {
			values, _ := json.MarshalIndent(struct {
				PellEmulator string `json:"pell_emulator"`
			}{
				PellEmulator: coreVsersion,
			}, "", "  ")
			fmt.Println(string(values))
		} else {
			fmt.Println(coreVsersion)
		}
	},
}

func init() {
	VersionCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show protocol and library versions")
}
