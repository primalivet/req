package gql

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
)

type Body struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

type Command struct {
	Body     Body
	Endpoint url.URL
	Header   http.Header
}

type GQL struct {
	logger *slog.Logger
}

func New(logger *slog.Logger) *GQL {
	return &GQL{logger: logger}
}

func (gql *GQL) NewCommand(args []string) (*Command, error) {
	gql.logger.Debug("Args", "args", args)
	fs := flag.NewFlagSet("gql", flag.ExitOnError)
	query := fs.String("query", "", "String query for the GraphQL request")
	file := fs.String("file", "", "Path to a file containing the query for the GraphQL request")
	token := fs.String("token", "", "Bearer token to use for authentication")
	raw := fs.Bool("raw", false, "Show raw response body, including headers")
	endpoint := fs.String("endpoint", "", "Endpoint URL to send the request to")

	fs.Parse(args[1:])

	gql.logger.Debug(fmt.Sprintf("Flag -token %s", *token))
	gql.logger.Debug(fmt.Sprintf("Flag -query %s", *query))
	gql.logger.Debug(fmt.Sprintf("Flag -file %s", *file))
	gql.logger.Debug(fmt.Sprintf("Flag -raw %v", *raw))
	gql.logger.Debug(fmt.Sprintf("Flag -endpoint %s", *endpoint))

	lastArg := args[len(args)-1]
	gql.logger.Debug(fmt.Sprintf("Last arg %s", lastArg))

	cmd := &Command{
		Body:   Body{},
		Header: http.Header{},
	}

	if parsedEndpoint, err := url.ParseRequestURI(lastArg); err == nil {
		// if arg is endpoint then use it as endpoint (require either -query or -file flag)

		gql.logger.Debug("Last arg is URL", "lastArg", lastArg)
		cmd.Endpoint = *parsedEndpoint

		switch {
		case *query == "" && *file == "":
			gql.logger.Error("Must provide either query or file when using URL as last argument")
			os.Exit(1)
		case *query != "" && *file == "":
			cmd.Body.Query = *query
		case *query == "" && *file != "":
			content, err := os.ReadFile(*file)
			if err != nil {
				gql.logger.Error("Error reading query file", "error", err)
				os.Exit(1)
			}
			cmd.Body.Query = string(content)
		}
	} else {
		// lastArg was not the endpoint, so here we require -endpoint flag
		parsedEndpoint, err := url.ParseRequestURI(*endpoint)
		if err != nil {
			gql.logger.Error("Must provide endpoint when using file or query as last argument")
			os.Exit(1)
		}
		cmd.Endpoint = *parsedEndpoint

		// lastArg has to be either a file or a query
		if parsedFile, err := os.ReadFile(lastArg); err == nil {
			cmd.Body.Query = string(parsedFile)
		} else if lastArg != "" {
			cmd.Body.Query = lastArg
		} else {
			gql.logger.Error("Last argument is neither a valid file nor a valid query")
			os.Exit(1)
		}
	}

	return cmd, nil
}

func (gql *GQL) ToRequest(cmd *Command) (*http.Request, error) {
	json, err := json.Marshal(cmd.Body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, cmd.Endpoint.String(), bytes.NewBuffer(json))
	if err != nil {
		return nil, err
	}

	req.Header = cmd.Header.Clone()
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}
