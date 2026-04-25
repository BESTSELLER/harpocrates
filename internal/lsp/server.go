package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// Server represents the LSP server
type Server struct {
	documents          *DocumentStore
	completionProvider *CompletionProvider
	tokenErr           error
}

// NewServer creates a new LSP server

func NewServer(vaultClient VaultClient, tokenErr error) *Server {
	documents := NewDocumentStore()

	return &Server{
		documents:          documents,
		completionProvider: NewCompletionProvider(documents, vaultClient, time.Minute*5),
		tokenErr:           tokenErr,
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
		s.handleInitialize(req)
	case "initialized":
		s.handleInitialized()
	case "textDocument/didOpen":
		s.handleDidOpen(req)
	case "textDocument/didChange":
		s.handleDidChange(req)
	case "textDocument/didSave":
		// Safely ignore didSave
	case "textDocument/completion":
		s.handleCompletion(req)
	case "shutdown":
		s.handleShutdown(req)
	case "exit":
		s.handleExit()
	default:
		if req.ID != nil {
			s.writeError(req.ID, -32601, "Method not found")
		}
	}
}

func (s *Server) handleInitialize(req Request) {
	if _, ok := parseParams[InitializeParams](req.Params, "failed to unmarshal initialize params"); !ok {
		return
	}

	result := InitializeResult{
		Capabilities: ServerCapabilities{
			TextDocumentSync: 1, // Full document sync
			CompletionProvider: &CompletionOptions{
				TriggerCharacters: []string{"/", "-", " ", "\n", "\r"},
			},
		},
	}

	s.writeResponse(req.ID, result)
}

func (s *Server) handleInitialized() {
	if s.tokenErr == nil {
		return
	}

	s.SendNotification("window/showMessage", ShowMessageParams{
		Type:    2, // Warning
		Message: "Harpocrates: Vault token validation failed. Autocomplete and validation may not work. Please make sure you are logged in with with the vault cli.",
	})
}

func (s *Server) handleDidOpen(req Request) {
	params, ok := parseParams[DidOpenTextDocumentParams](req.Params, "failed to unmarshal didOpen params")
	if !ok {
		return
	}

	s.documents.Open(params.TextDocument.URI, params.TextDocument.Text)
}

func (s *Server) handleDidChange(req Request) {
	params, ok := parseParams[DidChangeTextDocumentParams](req.Params, "failed to unmarshal didChange params")
	if !ok {
		return
	}
	if len(params.ContentChanges) == 0 {
		return
	}

	s.documents.Change(params.TextDocument.URI, params.ContentChanges[0].Text)
}

func (s *Server) handleCompletion(req Request) {
	params, ok := parseParams[CompletionParams](req.Params, "failed to unmarshal completion params")
	if !ok {
		return
	}

	s.writeResponse(req.ID, s.completionProvider.Provide(params))
}

func (s *Server) handleShutdown(req Request) {
	s.writeResponse(req.ID, nil)
}

func (s *Server) handleExit() {
	os.Exit(0)
}

func parseParams[T any](params json.RawMessage, errorMessage string) (T, bool) {
	var decoded T
	if err := json.Unmarshal(params, &decoded); err != nil {
		log.Error().Err(err).Msg(errorMessage)
		return decoded, false
	}
	return decoded, true
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

	s.writeMessage(body)
}

func (s *Server) writeError(id any, code int, message string) {
	resp := Response{
		RPC: "2.0",
		ID:  id,
		Error: &Error{
			Code:    code,
			Message: message,
		},
	}

	body, err := json.Marshal(resp)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal error response")
		return
	}

	s.writeMessage(body)
}

func (s *Server) SendNotification(method string, params any) {
	notif := Notification{
		RPC:    "2.0",
		Method: method,
		Params: params,
	}

	body, err := json.Marshal(notif)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal notification")
		return
	}

	s.writeMessage(body)
}

func (s *Server) writeMessage(body []byte) {
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(body))

	// Write atomically to avoid interleaving
	os.Stdout.Write([]byte(header)) //nolint:errcheck // If this fails here, there is no reason to handle it.
	os.Stdout.Write(body)           //nolint:errcheck // If this fails here, there is no reason to handle it.
}
