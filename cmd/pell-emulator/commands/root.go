package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/0xPellNetwork/pell-emulator/cmd/pell-emulator/chainflags"
	"github.com/0xPellNetwork/pell-emulator/cmd/pell-emulator/commands/mocks"
	"github.com/0xPellNetwork/pell-emulator/cmd/pell-emulator/utils"
	"github.com/0xPellNetwork/pell-emulator/config"
	"github.com/0xPellNetwork/pell-emulator/libs/cli"
	"github.com/0xPellNetwork/pell-emulator/libs/log"
)

var (
	logger = log.NewLogger(os.Stdout)
)

func init() {
	registerFlagsRootCmd(RootCmd)
}

func registerFlagsRootCmd(cmd *cobra.Command) {
	chainflags.LogLevelFlag.AddToCmdPersistentFlags(cmd)
	chainflags.LogFormatFlag.AddToCmdPersistentFlags(cmd)
}

func ParseConfig(configFile, chainFile string) (*config.Config, error) {
	configFile = strings.TrimSpace(configFile)
	chainFile = strings.TrimSpace(chainFile)

	if configFile == "" && chainFile == "" {
		logger.Info("No config file and chain file specified, use default Config")
		return config.DefaultConfig(), nil
	}

	var conf *config.Config
	var err error
	conf, err = config.LoadConfigFromFile(configFile)
	if err != nil {
		logger.Error(
			"Failed to load config from file, use default Config",
			"error", err,
			"file", configFile,
		)
		conf = config.DefaultConfig()
	}

	// load chain address from file
	if chainFile != "" && conf.ContractAddress == nil {
		if !utils.FileExists(chainFile) {
			logger.Error("Chain address file not exists", "file", chainFile)
			return nil, fmt.Errorf("chain address file was provied but the file don't exists: %s", chainFile)
		}
		address, err := config.LoadContractAddressFromFile(chainFile)
		if err != nil {
			logger.Error("Failed to load chain address from file, use default contract address",
				"error", err,
				"file", chainFile,
			)
			address = config.DefaultContractAddress()
		}
		conf.ContractAddress = address
	}

	return conf, nil
}

func overwriteConfigOnRootCmd(conf *config.Config) {
	if chainflags.LogLevelFlag.GetValue() != "" {
		conf.LogLevel = chainflags.LogLevelFlag.GetValue()
	}
	if chainflags.LogFormatFlag.GetValue() != "" {
		conf.LogFormat = chainflags.LogFormatFlag.GetValue()
	}

	if chainflags.EmulatorFlagRPCURL.Value != "" {
		conf.RPCURL = chainflags.EmulatorFlagRPCURL.Value
	}
	if chainflags.EmulatorFlagWSURL.Value != "" {
		conf.WSURL = chainflags.EmulatorFlagWSURL.Value
	}

	if chainflags.EmulatorFlagAutoUpdateConnector.GetValue() != "" {
		conf.AutoUpdateConnector = chainflags.EmulatorFlagAutoUpdateConnector.GetBool()
	}

	if chainflags.EmulatorFlagDeployerKeyFile.Value != "" {
		conf.DeployerKeyFile = chainflags.EmulatorFlagDeployerKeyFile.Value
	}
}

func validateFlagsOnRootCmd() error {
	// check if deployerKeyFilepath file exists
	if chainflags.EmulatorFlagDeployerKeyFile.Value != "" && !utils.FileExists(chainflags.EmulatorFlagDeployerKeyFile.Value) {
		logger.Error("deployer Key Filepath", "path", chainflags.EmulatorFlagDeployerKeyFile.Value)
		return fmt.Errorf("deployer Key Filepath not exists: %s", chainflags.EmulatorFlagDeployerKeyFile.Value)
	}

	return nil
}

// RootCmd is the root command for Pell Emulator core.
var RootCmd = &cobra.Command{
	Use:   "pell-emulator",
	Short: "pell emulator operations",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
		if cmd.Name() == VersionCmd.Name() {
			return nil
		}

		err = validateFlagsOnRootCmd()
		if err != nil {
			return err
		}

		if chainflags.LogFormatFlag.GetValue() == "json" {
			logger = log.NewLogger(os.Stdout, log.OutputJSONOption())
		}

		if viper.GetBool(cli.TraceFlag) {
			logger = log.NewTracingLogger(logger)
		}

		configFile := utils.GetHomeDir(cmd) + "/config/config.json"
		if chainflags.EmulatorFlagConfigFile.Value != "" {
			configFile = chainflags.EmulatorFlagConfigFile.Value
		}
		contractConfigFile := chainflags.EmulatorFlagContractAddressFile.Value

		logger.Info("config file", "path", configFile)

		conf, err := ParseConfig(configFile, contractConfigFile)
		if err != nil {
			return err
		}

		overwriteConfigOnRootCmd(conf)

		config.SetGlobalConfig(conf)

		return nil
	},
}

func init() {

	chainflags.EmulatorFlagConfigFile.AddToCmdPersistentFlags(RootCmd)
	chainflags.EmulatorFlagContractAddressFile.AddToCmdPersistentFlags(RootCmd)

	chainflags.EmulatorFlagRPCURL.AddToCmdPersistentFlags(RootCmd)
	chainflags.EmulatorFlagWSURL.AddToCmdPersistentFlags(RootCmd)

	RootCmd.AddCommand(EmulatorInitCmd)
	RootCmd.AddCommand(EmulatorStartCmd)
	RootCmd.AddCommand(EmulatorUpdateConnectorCmd)

	RootCmd.AddCommand(mocks.EmulatorMocksCmd)
	RootCmd.AddCommand(VersionCmd)
	RootCmd.AddCommand(cli.NewCompletionCmd(RootCmd, true))
}
