package rest

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
)

type Command struct {
	Body any
	Endpoint url.URL
	Header http.Header
}

type REST struct {
	logger *slog.Logger
}

func New(logger *slog.Logger) *REST {
	return &REST{logger: logger}
}

func (rest *REST) NewCommand(args []string) (*Command, error) {
		fs := flag.NewFlagSet("http", flag.ExitOnError)
		method := fs.String("method", "GET", "Method to do the request")
		token := fs.String("token", "", "Bearer token to use for authentication")
		raw := fs.Bool("raw", false, "Show raw response body, including headers")

		fs.Parse(args[1:])

		rest.logger.Debug(fmt.Sprintf("Flag -method %s", *method))
		rest.logger.Debug(fmt.Sprintf("Flag -token %s", *token))
		rest.logger.Debug(fmt.Sprintf("Flag -raw %v", *raw))

		url, err := url.Parse(args[len(args)-1])
		if err != nil {
			rest.logger.Error("Invalid URL", "error", err)
			os.Exit(1)
		}
		rest.logger.Debug(fmt.Sprintf("URL %s", url))

	return nil, nil
}

func (rest *REST) ToRequest(cmd *Command) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, cmd.Endpoint.String(), nil)
	if err != nil {
		rest.logger.Error("Error making request", "error", err)
		os.Exit(1)
	}
	return req, nil
}
