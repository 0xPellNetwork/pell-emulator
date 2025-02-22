package commands

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/0xPellNetwork/pell-emulator/cmd/pell-emulator/chainflags"
	"github.com/0xPellNetwork/pell-emulator/config"
	"github.com/0xPellNetwork/pell-emulator/server"
)

var emulatorStartCmdFlagPort = &chainflags.IntFlag{
	Name:  "port",
	Usage: "port",
}

func init() {
	chainflags.EmulatorFlagRPCURL.AddToCmdFlag(EmulatorStartCmd)
	chainflags.EmulatorFlagWSURL.AddToCmdFlag(EmulatorStartCmd)
	chainflags.EmulatorFlagAutoUpdateConnector.AddToCmdFlag(EmulatorStartCmd)
	chainflags.EmulatorFlagDeployerKeyFile.AddToCmdFlag(EmulatorStartCmd)

	emulatorStartCmdFlagPort.AddToCmdFlag(EmulatorStartCmd)
}

var EmulatorStartCmd = &cobra.Command{
	Use:   "start",
	Short: "start emulator",
	Example: `
pell-emulator start \
	--home <home-dir> \
	--rpc-url <rpc-url> \
	--ws-url <ws-url> \
	--deployer-key-file <deployer-key-file> \
	--port <port, defaults to 9090> \
	--auto-update-connector <true/false, currently default false, 1/t/y/yes will be true>

pell-emulator start \
	--home ./_pd-proj_pell_pell-emulator/emulator-home-pelldvs-example \
	--rpc-url http://localhost:8545 \
	--ws-url ws://localhost:8545 \
	--deployer-key-file /path/to/deployer-key-file \
	--port 9090 \
	--auto-update-connector true
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Info("start emulator, params",
			"k", "v",
			"rpcURL", chainflags.EmulatorFlagRPCURL.GetValue(),
			"wsURL", chainflags.EmulatorFlagWSURL.GetValue(),
			"contractAddressFile", chainflags.EmulatorFlagContractAddressFile.GetValue(),
			"deployerKeyFile", chainflags.EmulatorFlagDeployerKeyFile.GetValue(),
			"autoUpdateConnector", chainflags.EmulatorFlagAutoUpdateConnector.GetBool(),
			"config", chainflags.EmulatorFlagConfigFile.GetValue(),
		)

		cfg := config.GetGlobalConfig()
		if isValidPort(emulatorStartCmdFlagPort.Value) {
			cfg.Port = emulatorStartCmdFlagPort.Value
		}
		if !isValidPort(cfg.Port) {
			cfg.Port = config.DefautlHTTPServerPort
		}

		logger.Info("cfg is", "cfg", cfg)

		rootCtx := cmd.Context()
		ctx, cancel := context.WithCancel(rootCtx)
		defer cancel()

		errCh := make(chan error, 1)
		srv, err := server.NewServer(rootCtx, cfg, logger, cfg.Port)
		if err != nil {
			logger.Error("Failed to create server", "error", err)
			return err
		}
		go func() {
			if err := srv.Start(ctx); err != nil {
				errCh <- fmt.Errorf("server error: %w", err)
				cancel() // cancel the context to stop the server
			}
		}()

		// Handle SIGINT and SIGTERM.
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		defer func() {
			signal.Stop(sigCh)
			close(sigCh)
		}()

		select {
		case sig := <-sigCh:
			logger.Info("Received shutdown signal", "signal", sig)
		case err := <-errCh:
			logger.Error("Service error", "error", err)
		case <-ctx.Done():
			logger.Info("Context canceled")
		}

		logger.Info("Starting graceful shutdown...")
		cancel()

		return nil
	},
}

func isValidPort(value int) bool {
	return value > 0 && value <= 65535
}
