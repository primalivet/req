package main

import (
	"bytes"
	"flag"
	"fmt"
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

func makeHTTPRequest(deps HTTPRequestDeps) (*http.Request, error) {
	req, err := http.NewRequest(string(deps.Method), deps.Endpoint.String(), deps.Body)
	if err != nil {
		return nil, err
	}
	req.Header = deps.Header.Clone()
	return req, nil
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
	flag.Parse()

	logger := makeLogger(debug)

	logger.Debug(fmt.Sprintf("Flag -debug %v", *debug))

	args := flag.Args()

	switch args[0] {
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

		client := &http.Client{}

		resp, err := client.Do(req)
		if err != nil {
			logger.Error("Error making request", "error", err)
			os.Exit(1)
		}

		printResponse(logger, resp)

	case "http":

		client := http.Client{}

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
		printResponse(logger, resp)

	default:
		logger.Info("Unknown command")
	}
}

func printResponse(logger *slog.Logger, resp *http.Response) {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Error reading response body", "error", err)
		os.Exit(1)
	}

	rawResp := bytes.Buffer{}
	rawResp.WriteString(fmt.Sprintf("%s %s\n", resp.Proto, resp.Status))
	for k, v := range resp.Header {
		rawResp.WriteString(fmt.Sprintf("%s: %s\n", k, v))
	}
	rawResp.WriteString("\n")
	rawResp.Write(body)
	fmt.Println(rawResp.String())
}
