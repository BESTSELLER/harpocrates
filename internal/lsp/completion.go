package lsp

import (
	"sort"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// CompletionContext tells us where we are in the YAML document
type CompletionContext int

const (
	ContextUnknown CompletionContext = iota
	ContextSecretsList
	ContextKeysList
)

type CompletionProvider struct {
	documents       *DocumentStore
	vaultClient     VaultClient
	secretListCache *TTLMap[[]string]
	secretReadCache *TTLMap[map[string]any]
}

type completionRequest struct {
	params        CompletionParams
	lines         []string
	trimmedPrefix string
	needsDash     bool
	prefix        string
}

func NewCompletionProvider(documents *DocumentStore, vaultClient VaultClient, cacheTTL time.Duration) *CompletionProvider {
	return &CompletionProvider{
		documents:       documents,
		vaultClient:     vaultClient,
		secretListCache: NewTTLMap[[]string](cacheTTL),
		secretReadCache: NewTTLMap[map[string]any](cacheTTL),
	}
}

func (p *CompletionProvider) Provide(params CompletionParams) CompletionList {
	request, ok := p.newCompletionRequest(params)
	if !ok {
		return CompletionList{}
	}

	parsedCtx := parseContext(request.lines, params.Position.Line)
	switch parsedCtx.Type {
	case ContextSecretsList:
		return p.completeSecrets(request)
	case ContextKeysList:
		return p.completeKeys(request, parsedCtx)
	default:
		return CompletionList{}
	}
}

func (p *CompletionProvider) newCompletionRequest(params CompletionParams) (completionRequest, bool) {
	content, ok := p.documents.Get(params.TextDocument.URI)
	if !ok {
		return completionRequest{}, false
	}

	lines := strings.Split(content, "\n")
	if params.Position.Line >= len(lines) {
		return completionRequest{}, false
	}
	currentLine := lines[params.Position.Line]

	prefix := ""
	if params.Position.Character <= len(currentLine) {
		prefix = currentLine[:params.Position.Character]
	}

	if strings.HasSuffix(strings.TrimSpace(prefix), ":") {
		return completionRequest{}, false
	}

	if strings.HasSuffix(prefix, " ") {
		trimmed := strings.TrimSpace(prefix)
		if trimmed != "" && trimmed != "-" {
			return completionRequest{}, false
		}
	}

	trimmedPrefix := strings.TrimLeft(prefix, " \t-")
	if idx := strings.LastIndexAny(trimmedPrefix, " :"); idx != -1 {
		trimmedPrefix = strings.TrimSpace(trimmedPrefix[idx+1:])
	}
	trimmedPrefix = strings.TrimLeft(trimmedPrefix, "'\"")

	needsDash := !strings.HasPrefix(strings.TrimSpace(prefix), "-")

	return completionRequest{
		params:        params,
		lines:         lines,
		trimmedPrefix: trimmedPrefix,
		needsDash:     needsDash,
		prefix:        prefix,
	}, true
}

func (p *CompletionProvider) completeSecrets(request completionRequest) CompletionList {
	if p.vaultClient == nil {
		return CompletionList{}
	}

	basePath := ""
	if idx := strings.LastIndex(request.trimmedPrefix, "/"); idx != -1 {
		basePath = request.trimmedPrefix[:idx+1]
	}

	tokens := p.listSecretTokens(basePath)
	currentWord := strings.TrimPrefix(request.trimmedPrefix, basePath)

	var items []CompletionItem
	for _, token := range tokens {
		if !strings.HasPrefix(token, currentWord) {
			continue
		}

		kind := CompletionItemKindValue
		if strings.HasSuffix(token, "/") {
			kind = CompletionItemKindFolder
		}

		var cmd *Command
		if kind == CompletionItemKindFolder {
			cmd = &Command{
				Title:   "Trigger Suggest",
				Command: "editor.action.triggerSuggest",
			}
		}

		items = append(items, newCompletionItem(token, kind, request, currentWord, cmd))
	}
	return CompletionList{Items: items}
}

func (p *CompletionProvider) completeKeys(request completionRequest, parsedCtx ParserContext) CompletionList {
	if p.vaultClient == nil || parsedCtx.ParentSecret == "" {
		return CompletionList{}
	}

	secretData, ok := p.readSecret(parsedCtx.ParentSecret)
	if !ok {
		return CompletionList{}
	}

	keys := make([]string, 0, len(secretData))
	for key := range secretData {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var items []CompletionItem
	for _, key := range keys {
		if parsedCtx.Existing[key] || !strings.HasPrefix(key, request.trimmedPrefix) {
			continue
		}

		items = append(items, newCompletionItem(key, CompletionItemKindField, request, request.trimmedPrefix, nil))
	}
	return CompletionList{Items: items}
}

func (p *CompletionProvider) listSecretTokens(basePath string) []string {
	queryPath := strings.Replace(basePath, "/data/", "/metadata/", 1)
	cacheKey := "list:" + queryPath
	if tokens, ok := p.secretListCache.Get(cacheKey); ok {
		return tokens
	}

	tokens, err := p.vaultClient.ListTokens(queryPath)
	if err != nil {
		log.Error().Err(err).Str("path", queryPath).Msg("ListTokens failed")
	}

	if basePath == "" {
		engines, err := p.vaultClient.ListSecretEngines()
		if err != nil {
			log.Error().Err(err).Msg("ListSecretEngines failed")
		} else {
			tokens = append(tokens, engines...)
		}
	} else {
		tokens = p.withEngineSubPath(tokens, basePath)
	}

	p.secretListCache.Set(cacheKey, tokens)
	return tokens
}

func (p *CompletionProvider) withEngineSubPath(tokens []string, basePath string) []string {
	subPath, err := p.vaultClient.GetEngineSubPath(basePath)
	if err != nil {
		log.Error().Err(err).Str("path", basePath).Msg("GetEngineSubPath failed")
		return tokens
	}
	if subPath == "" {
		return tokens
	}

	for _, token := range tokens {
		if basePath+token == subPath {
			return tokens
		}
	}
	return append(tokens, strings.TrimPrefix(subPath, basePath))
}

func (p *CompletionProvider) readSecret(path string) (map[string]any, bool) {
	cacheKey := "read:" + path
	if secretData, ok := p.secretReadCache.Get(cacheKey); ok {
		return secretData, true
	}

	secretData, err := p.vaultClient.ReadSecret(path)
	if err != nil || secretData == nil {
		return nil, false
	}

	p.secretReadCache.Set(cacheKey, secretData)
	return secretData, true
}

func newCompletionItem(label string, kind int, request completionRequest, currentWord string, cmd *Command) CompletionItem {
	insertText := label
	wordStart := request.params.Position.Character - len(currentWord)

	if request.needsDash {
		insertText = "- " + insertText
	} else if wordStart > 0 && len(request.prefix) >= wordStart {
		if strings.HasSuffix(request.prefix[:wordStart], "-") {
			insertText = "- " + insertText
			wordStart--
		}
	}

	return CompletionItem{
		Label:      label,
		Kind:       kind,
		InsertText: insertText,
		FilterText: insertText,
		TextEdit: &TextEdit{
			Range: Range{
				Start: Position{
					Line:      request.params.Position.Line,
					Character: wordStart,
				},
				End: request.params.Position,
			},
			NewText: insertText,
		},
		Command: cmd,
	}
}
