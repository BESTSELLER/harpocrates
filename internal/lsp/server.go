package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/BESTSELLER/harpocrates/vault"
	"github.com/rs/zerolog/log"
)

// Server represents the LSP server
type Server struct {
	documents   map[string]string // URI to document content
	vaultClient *vault.API
}

// NewServer creates a new LSP server
func NewServer(vaultClient *vault.API) *Server {
	return &Server{
		documents:   make(map[string]string),
		vaultClient: vaultClient,
	}
}

// Start stars the LSP blocking loop
func (s *Server) Start() {
	reader := bufio.NewReader(os.Stdin)

	for {
		// Read headers
		var contentLength int
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					return // Client disconnected
				}
				log.Error().Err(err).Msg("failed to read from stdin")
				return
			}

			line = strings.TrimSpace(line)
			if line == "" {
				break // End of headers
			}

			if strings.HasPrefix(line, "Content-Length:") {
				parts := strings.Split(line, ":")
				if len(parts) == 2 {
					contentLength, _ = strconv.Atoi(strings.TrimSpace(parts[1]))
				}
			}
		}

		if contentLength == 0 {
			continue
		}

		body := make([]byte, contentLength)
		if _, err := io.ReadFull(reader, body); err != nil {
			log.Error().Err(err).Msg("failed to read body")
			continue
		}

		var req Request
		if err := json.Unmarshal(body, &req); err != nil {
			log.Error().Err(err).Msg("failed to parse message")
			continue
		}

		s.handleMessage(req)
	}
}

func (s *Server) handleMessage(req Request) {
	switch req.Method {
	case "initialize":
		var params InitializeParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			log.Error().Err(err).Msg("failed to unmarshal initialize params")
			return
		}

		result := InitializeResult{
			Capabilities: ServerCapabilities{
				TextDocumentSync: 1, // Full document sync
				CompletionProvider: &CompletionOptions{
					TriggerCharacters: []string{"/"},
				},
			},
		}

		s.writeResponse(req.ID, result)

	case "textDocument/didOpen":
		var params DidOpenTextDocumentParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			log.Error().Err(err).Msg("failed to unmarshal didOpen params")
			return
		}
		s.documents[params.TextDocument.URI] = params.TextDocument.Text

	case "textDocument/didChange":
		var params DidChangeTextDocumentParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			log.Error().Err(err).Msg("failed to unmarshal didChange params")
			return
		}
		if len(params.ContentChanges) > 0 {
			s.documents[params.TextDocument.URI] = params.ContentChanges[0].Text
		}

	case "textDocument/completion":
		var params CompletionParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			log.Error().Err(err).Msg("failed to unmarshal completion params")
			return
		}

		items := s.provideCompletions(params)

		s.writeResponse(req.ID, items)
	}
}

func (s *Server) writeResponse(id any, result any) {
	resp := Response{
		RPC:    "2.0",
		ID:     id,
		Result: result,
	}

	body, err := json.Marshal(resp)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal response")
		return
	}

	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(body))

	// Write atomically to avoid interleaving
	os.Stdout.Write([]byte(header)) //nolint:errcheck If this fails here, there is no reason to handle it.
	os.Stdout.Write(body)           //nolint:errcheck If this fails here, there is no reason to handle it.
}
