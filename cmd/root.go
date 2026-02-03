package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/mengbo/pacproxy/internal/config"
	"github.com/mengbo/pacproxy/pac"
	"github.com/mengbo/pacproxy/proxy"
	"github.com/spf13/cobra"
)

var (
	// Version is set during build
	Version = "dev"

	cfg = &config.Config{}
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pacproxy",
	Short: "A PAC-aware HTTP/HTTPS proxy",
	Long: `pacproxy is a lightweight local HTTP/HTTPS proxy that applies PAC 
(Proxy Auto-Config) rules to dynamically route traffic either directly 
or through upstream HTTP/SOCKS5 proxies.`,
	Version: Version,
	RunE:    run,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&cfg.ListenAddr, "listen", "l", "127.0.0.1:1080", "Listen address")
	rootCmd.Flags().StringVarP(&cfg.PACURL, "pac-url", "p", "http://clash-server:9090/ui/proxy.pac", "PAC file URL")
	rootCmd.Flags().StringVar(&cfg.LogLevel, "log-level", "info", "Log level (debug, info, warn, error)")
}

func run(cmd *cobra.Command, args []string) error {
	// Validate config
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Setup logging
	if err := setupLogging(cfg.LogLevel); err != nil {
		return fmt.Errorf("setting up logging: %w", err)
	}

	slog.Info("Starting pacproxy",
		"version", Version,
		"listen", cfg.ListenAddr,
		"pac-url", cfg.PACURL)

	// Create PAC fetcher
	fetcher := pac.NewFetcher(0) // 0 = use default timeout

	// Create router
	router, err := proxy.NewRouter(cfg.PACURL, fetcher)
	if err != nil {
		return fmt.Errorf("creating router: %w", err)
	}

	// Create handler
	handler := proxy.NewHandler(router)

	// Create server
	server, err := proxy.NewServer(cfg.ListenAddr, handler)
	if err != nil {
		return fmt.Errorf("creating server: %w", err)
	}

	// Start server
	if err := server.Start(); err != nil {
		return fmt.Errorf("starting server: %w", err)
	}

	slog.Info("pacproxy is running", "address", server.Addr())

	server.PrintUsageHelp()

	// Wait for shutdown signal
	server.WaitForShutdown()

	// Graceful shutdown
	if err := server.Stop(); err != nil {
		return fmt.Errorf("stopping server: %w", err)
	}

	slog.Info("pacproxy stopped")
	return nil
}

func setupLogging(level string) error {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		return fmt.Errorf("invalid log level: %s", level)
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey || a.Key == slog.LevelKey {
				return slog.Attr{}
			}
			return a
		},
	}

	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	return nil
}
