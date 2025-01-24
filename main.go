package main

import (
	"flag"
	"fmt"
	"github.com/primalivet/req/internal/display"
	"github.com/primalivet/req/internal/gql"
	"github.com/primalivet/req/internal/rest"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
)

type Method string

const (
	OPTION Method = http.MethodOptions
	PATCH  Method = http.MethodPatch
	DELETE Method = http.MethodDelete
	GET    Method = http.MethodGet
	POST   Method = http.MethodPost
	PUT    Method = http.MethodPut
)

type HTTPRequestDeps struct {
	Method   Method
	Body     io.Reader
	Endpoint url.URL
	Header   http.Header
}

func makeLogger(debug *bool) *slog.Logger {
	logLevel := slog.LevelInfo
	if debug != nil && *debug {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	}))
	return logger
}

func main() {
	debug := flag.Bool("debug", false, "Enable debug logging")
	raw := flag.Bool("raw", false, "Display raw response")
	headers := flag.Bool("headers", false, "Display response headers")
	flag.Parse()
	logger := makeLogger(debug)
	logger.Debug(fmt.Sprintf("Flag -debug %v", *debug))
	args := flag.Args()

	client := &http.Client{}
	display := display.New(logger, raw, headers)
	command := args[0]

	switch command {
	case "gql":
		handler := gql.New(logger)
		cmd, err := handler.NewCommand(args)
		if err != nil {
			logger.Error("Error creating command", "error", err)
			os.Exit(1)
		}
		req, err := handler.ToRequest(cmd)
		if err != nil {
			logger.Error("Error creating request", "error", err)
			os.Exit(1)
		}
		resp, err := client.Do(req)
		if err != nil {
			logger.Error("Error making request", "error", err)
			os.Exit(1)
		}
		display.Raw(resp)
	case "http":
		handler := rest.New(logger)
		cmd, err := handler.NewCommand(args)
		if err != nil {
			logger.Error("Error creating command", "error", err)
			os.Exit(1)
		}
		req, err := handler.ToRequest(cmd)
		if err != nil {
			logger.Error("Error creating request", "error", err)
			os.Exit(1)
		}
		resp, err := client.Do(req)
		if err != nil {
			logger.Error("Error making request", "error", err)
			os.Exit(1)
		}
		display.Raw(resp)
	default:
		logger.Info("Unknown command")
	}
}
