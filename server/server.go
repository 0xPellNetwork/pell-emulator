package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/0xPellNetwork/pell-emulator/config"
	"github.com/0xPellNetwork/pell-emulator/internal/chains"
	events2 "github.com/0xPellNetwork/pell-emulator/internal/events"
	"github.com/0xPellNetwork/pell-emulator/libs/log"
)

type Server struct {
	bindings *chains.ChainBindings
	logger   log.Logger
	port     int
}

func NewServer(
	ctx context.Context,
	cfg *config.Config,
	logger log.Logger,
	port int,
) (*Server, error) {
	bindings, err := chains.NewChainBindings(ctx, cfg, logger)
	if err != nil {
		logger.Error("Failed to create chain bindings", "error", err)
		return nil, err
	}
	return &Server{
		bindings: bindings,
		logger:   logger,
		port:     port,
	}, nil
}

func (s *Server) Start(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	// Start HTTP server
	g.Go(func() error {
		return s.startHTTPServer(ctx, s.port)
	})

	// Start emulator server
	emulatorCtx, emulatorCancel := context.WithCancel(ctx)
	g.Go(func() error {
		return s.startEmulator(emulatorCtx)
	})

	// Start ETH connection check
	clientClosedCh := make(chan struct{})
	g.Go(func() error {
		return s.startETHConnectionCheck(ctx, clientClosedCh)
	})

	// Wait for all goroutines to finish
	if err := g.Wait(); err != nil {
		emulatorCancel() // Stop the emulator service
		return err
	}

	select {
	case <-clientClosedCh:
		s.logger.Error("eth client closed, update state and stopping emulator")
		emulatorServerState.Disable("eth client closed")
		emulatorCancel()
	case <-ctx.Done():
		emulatorCancel()
		return ctx.Err()
	}

	<-ctx.Done()
	return nil
}

// start a goroutine to monitor the connection of rpcClient and wsClient, every 5 seconds
func (s *Server) startETHConnectionCheck(parrent context.Context, clientClosed chan struct{}) error {
	var blockNumber uint64
	var err error

	lg := s.logger.With("func", "CheckClientClosed")

	lg.Info("monitoring connection will be started soon")

	// init blockNumber
	blockNumber, err = s.bindings.WsClient.BlockNumber(parrent)
	if err != nil {
		lg.Error("init failed to get blockNumber from wsClient",
			"error", err,
		)
		return err
	}
	lg.Info("monitoring connection started", "blockNumber", blockNumber)

	var maxFailedTimes = 3
	var interval = 3 * time.Second

	go func(ctx context.Context) {
		failedTimes := 0
		for {
			select {
			case <-ctx.Done():
				lg.Info("parent done, exiting monitor")
				select {
				case clientClosed <- struct{}{}:
				default:
				}
				return
			default:
				// check connection
				blockNumber, err = s.bindings.WsClient.BlockNumber(ctx)
				if err != nil {
					lg.Error(
						"failed to get blockNumber from webSocketClient, may be connection lost",
						"error", err,
						"inverval", interval,
						"maxFailedTimes", maxFailedTimes,
						"failedTimes", failedTimes,
					)
					failedTimes++
				} else {
					failedTimes = 0 // reset failedTimes
				}

				if failedTimes > maxFailedTimes {
					lg.Error("wsClient.BlockNumber failed 5 times, exiting monitor")
					emulatorServerState.Disable("wsClient.BlockNumber failed 5 times")
					select {
					case clientClosed <- struct{}{}:
					default:
					}
					return
				}

				time.Sleep(interval)
				if time.Now().Second()%10 == 0 {
					lg.Debug("monitoring connection emulatorServerState",
						"interval", interval,
						"blockNumber", blockNumber,
						"fialedTimes", failedTimes,
					)
				}
			}

		}
	}(parrent)

	return nil
}

func (s *Server) startEmulator(ctx context.Context) error {
	events := events2.GetAllEvents(
		s.bindings.ChainID,
		s.bindings.RPCClient,
		s.bindings.RPCBindings,
		s.bindings.WsClient,
		s.bindings.WsBindings,
		s.bindings.TxMgr,
		s.logger,
	)
	s.logger.Info("events loaded", "count", len(events))

	for _, event := range events {
		err := event.Init(ctx)
		if err != nil {
			s.logger.Error("event init failed", "event", event, "error", err)
			return err
		}

		go func() {
			_ = event.Listen(ctx)
		}()
		s.logger.Info("event started")
	}

	emulatorServerState.Enable()

	fmt.Println()
	fmt.Println()
	fmt.Println("start listening for events...")

	// Wait for the context to be canceled
	<-ctx.Done()

	fmt.Println("Main function exiting...")
	fmt.Println()

	return nil
}

func (s *Server) startHTTPServer(ctx context.Context, port int) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		statusMutex.RLock()
		defer statusMutex.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(emulatorServerState)
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go func() {
		s.logger.Info("Starting HTTP server", "port", port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("HTTP server failed", "error", err)
			return
		}
	}()

	<-ctx.Done()
	s.logger.Info("Shutting down HTTP server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("HTTP server shutdown failed", "error", err)
		return err
	} else {
		s.logger.Info("HTTP server gracefully stopped")
	}

	return nil
}
