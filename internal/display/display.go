package display

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
)

type Display struct {
	raw     bool
	headers bool
	logger  *slog.Logger
}

func New(logger *slog.Logger, raw *bool, headers *bool) *Display {
	return &Display{raw: *raw, headers: *headers, logger: logger}
}

func (d *Display) Raw(resp *http.Response) {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		d.logger.Error("Error reading response body", "error", err)
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
