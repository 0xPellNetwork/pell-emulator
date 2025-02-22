package main

import (
	"os"
	"path/filepath"

	"github.com/0xPellNetwork/pell-emulator/cmd/pell-emulator/commands"
	"github.com/0xPellNetwork/pell-emulator/libs/cli"
)

func main() {
	cmd := cli.PrepareBaseCmd(
		commands.RootCmd,
		"PELL_EMULATOR", os.ExpandEnv(filepath.Join("$HOME", ".pell-emulator")),
	)
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
