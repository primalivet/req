package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
)

type GraphQLRequest struct {
	Query     string `json:"query"`
}

func main() {
	var (
		file string
		vars string
		endpoint string
		token string
		showHeaders bool
		showRaw bool
	)

	flag.StringVar(&file, "file", "", "Path to the graphql query file (required)")
	flag.StringVar(&vars, "vars", "", "Path to the graphql variables file")
	flag.StringVar(&endpoint, "endpoint", "", "GraphQL endpoint to run the query against, can be set with the ENDPOINT environment variable")
	flag.StringVar(&token, "token", "", "Bearer token to use for authentication")
	flag.BoolVar(&showHeaders, "show-headers", false, "Show response headers")
	flag.BoolVar(&showRaw, "show-raw", false, "Show raw response body, including headers")
	flag.Parse()

	if file == "" {
		fmt.Println("Missing required arguments: file")
		flag.Usage()
		os.Exit(1)
	}

	if endpoint == "" {
		endpoint = os.Getenv("ENDPOINT")
		if endpoint == "" {
			fmt.Println("Missing required arguments: endpoint")
			flag.Usage()
			os.Exit(1)
		}
	}

	query, err := os.ReadFile(file)
	if err != nil {
		fmt.Println("Error reading query file:", err)
		os.Exit(1)
	}

	reqBody := &GraphQLRequest{
		Query: string(query),
	}

	reqBodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Println("Error encoding request body:", err)
		os.Exit(1)
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(reqBodyJSON))
	if err != nil {
		fmt.Println("Error creating request:", err)
		os.Exit(1)
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer " + token)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		os.Exit(1)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		os.Exit(1)
	}

	if showRaw {
		rawResp := bytes.Buffer{}
		rawResp.WriteString(fmt.Sprintf("%s %s\n", resp.Proto, resp.Status))
		for k, v := range resp.Header {
			rawResp.WriteString(fmt.Sprintf("%s: %s\n", k, v))
		}
		rawResp.WriteString("\n")
		rawResp.Write(body)
		fmt.Println(rawResp.String())
	} else {
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, body, "", "  "); err != nil {
			fmt.Println("Error pretty-printing JSON:", err)
			os.Exit(1)
		}

		if showHeaders {
			for k, v := range resp.Header {
				fmt.Printf("%s: %s\n", k, v)
			}
			fmt.Println()
		}

		fmt.Println(prettyJSON.String())
	}
}
