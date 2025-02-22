package utils

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pell-emulator/libs/cli"
)

func GetHomeDir(cmd *cobra.Command) string {
	var home string
	var err error
	home = GetEnvAny("PELLEMULATORHOME", "PELL_EMULATOR_HOME")
	if home == "" {
		home, err = cmd.Flags().GetString(cli.HomeFlag)
		if err != nil {
			home = os.ExpandEnv("$HOME/.pell-emulator")
		}
	}

	return home
}

func GetEnvAny(names ...string) string {
	for _, name := range names {
		if value := os.Getenv(name); value != "" {
			return value
		}
	}
	return ""
}

func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}
