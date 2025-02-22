package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pell-emulator/cmd/pell-emulator/chainflags"
	"github.com/0xPellNetwork/pell-emulator/cmd/pell-emulator/utils"
	"github.com/0xPellNetwork/pell-emulator/config"
)

var EmulatorInitCmd = &cobra.Command{
	Use:   "init",
	Short: "init pell emulator config",
	Long: `init will create a default config file in the home directory, defaults to ~/.pell-emulator/config/config.json
`,
	Example: `
pell-emulator init --home <home-dir>, defaults to ~/.pell-emulator>
`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// private validator
		dftConfig := config.DefaultConfig()
		cfgFile := chainflags.EmulatorFlagConfigFile.Value
		homeDir := utils.GetHomeDir(cmd)
		if cfgFile == "" {
			cfgFile = os.ExpandEnv(homeDir + "/config/config.json")
		}

		// check if file exists
		if _, err := os.Stat(cfgFile); err == nil {
			fmt.Println("config file exists, skip init: " + cfgFile)
			return nil
		}

		dir, filename := filepath.Split(cfgFile)
		logger.Info("config file", "dir", dir, "filename", filename)

		logger.Info("ensure dir", "dir", dir)
		err := os.MkdirAll(dir, 0700)
		if err != nil {
			logger.Error("failed to create dir", "err", err)
			return err
		}

		jsonBytes, err := json.MarshalIndent(dftConfig, "", "  ")
		if err != nil {
			logger.Error("failed to marshal config", "err", err)
			return err
		}
		err = os.WriteFile(cfgFile, jsonBytes, 0600)
		if err != nil {
			logger.Error("failed to write config file", "err", err)
			return err
		}

		fmt.Println()
		fmt.Println("init pell_emulator config success to " + cfgFile)
		fmt.Println()

		return nil
	},
}
