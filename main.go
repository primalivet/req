package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
)

type GraphQLRequest struct {
	Query string `json:"query"`
}

func main() {
	debug := flag.Bool("debug", false, "Enable debug logging")

	flag.Parse()

	logLevel := slog.LevelInfo
	if debug != nil && *debug {
		logLevel = slog.LevelDebug
	}

	var logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	}))

	logger.Debug(fmt.Sprintf("Flag -debug %v", *debug))

	args := flag.Args()

	switch args[0] {
	case "gql":
		fs := flag.NewFlagSet("gql", flag.ExitOnError)
		query := fs.String("query", "", "String query for the GraphQL request")
		file := fs.String("file", "", "Path to a file containing the query for the GraphQL request")
		token := fs.String("token", "", "Bearer token to use for authentication")
		raw := fs.Bool("raw", false, "Show raw response body, including headers")

		fs.Parse(args[1:])

		logger.Debug(fmt.Sprintf("Flag -token %s", *token))
		logger.Debug(fmt.Sprintf("Flag -query %s", *query))
		logger.Debug(fmt.Sprintf("Flag -file %s", *file))
		logger.Debug(fmt.Sprintf("Flag -raw %v", *raw))

		url, err := url.Parse(args[len(args)-1])
		if err != nil {
			logger.Error("Invalid URL")
			os.Exit(1)
		}

		logger.Debug(fmt.Sprintf("URL %s", url))

		client := &http.Client{}

		var queryStr *string

		switch {
		case *query != "" && *file != "":
			logger.Error("Cannot specify both query and file")
			os.Exit(1)
		case *file != "":
			content, err := os.ReadFile(*file)
			if err != nil {
				logger.Error("Error reading query file", "error", err)
				os.Exit(1)
			}
			logger.Debug("Read file content", "file", content)
			value := string(content)
			queryStr = &value
		case *query != "":
			queryStr = query
		default:
			logger.Error("Must provide either query and file")
			os.Exit(1)
		}

		reqBody := &GraphQLRequest{
			Query: *queryStr,
		}

		reqBodyJSON, err := json.Marshal(reqBody)
		if err != nil {
			logger.Error("Error encoding request body", "error", err)
			os.Exit(1)
		}

		req, err := http.NewRequest(http.MethodPost, url.String(), bytes.NewBuffer(reqBodyJSON))
		if err != nil {
			logger.Error("Error creating request", "error", err)
			os.Exit(1)
		}

		req.Header.Set("Content-Type", "application/json")

		if *token != "" {
			req.Header.Set("Authorization", "Bearer "+*token)
		}

		resp, err := client.Do(req)
		if err != nil {
			logger.Error("Error making request", "error", err)
			os.Exit(1)
		}

		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Error("Error reading response body", "error", err)
			os.Exit(1)
		}

		if *raw {
			rawResp := bytes.Buffer{}
			rawResp.WriteString(fmt.Sprintf("%s %s\n", resp.Proto, resp.Status))
			for k, v := range resp.Header {
				rawResp.WriteString(fmt.Sprintf("%s: %s\n", k, v))
			}
			rawResp.WriteString("\n")
			rawResp.Write(body)
			fmt.Println(rawResp.String())
		}

	case "http":
		fs := flag.NewFlagSet("http", flag.ExitOnError)
		method := fs.String("method", "GET", "Method to do the request")
		token := fs.String("token", "", "Bearer token to use for authentication")
		raw := fs.Bool("raw", false, "Show raw response body, including headers")

		fs.Parse(args[1:])

		logger.Debug(fmt.Sprintf("Flag -method %s", *method))
		logger.Debug(fmt.Sprintf("Flag -token %s", *token))
		logger.Debug(fmt.Sprintf("Flag -raw %v", *raw))

		url, err := url.Parse(args[len(args)-1])
		if err != nil {
			logger.Error("Invalid URL", "error", err)
			os.Exit(1)
		}
		logger.Debug(fmt.Sprintf("URL %s", url))

		req, err := http.NewRequest(http.MethodGet, url.String(), nil)
		if err != nil {
			logger.Error("Error making request", "error", err)
			os.Exit(1)
		}

		client := http.Client{}

		resp, err := client.Do(req)
		if err != nil {
			logger.Error("Error making request", "error", err)
			os.Exit(1)
		}

		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Error("Error reading response body", "error", err)
			os.Exit(1)
		}

		if *raw {
			rawResp := bytes.Buffer{}
			rawResp.WriteString(fmt.Sprintf("%s %s\n", resp.Proto, resp.Status))
			for k, v := range resp.Header {
				rawResp.WriteString(fmt.Sprintf("%s: %s\n", k, v))
			}
			rawResp.WriteString("\n")
			rawResp.Write(body)
			fmt.Println(rawResp.String())
		}
	default:
		logger.Info("Unknown command")
	}

	// if showRaw {
	// } else {
	// 	var prettyJSON bytes.Buffer
	// 	if err := json.Indent(&prettyJSON, body, "", "  "); err != nil {
	// 		fmt.Println("Error pretty-printing JSON:", err)
	// 		os.Exit(1)
	// 	}
	//
	// 	if showHeaders {
	// 		for k, v := range resp.Header {
	// 			fmt.Printf("%s: %s\n", k, v)
	// 		}
	// 		fmt.Println()
	// 	}
	//
	// 	fmt.Println(prettyJSON.String())
	// }
}
