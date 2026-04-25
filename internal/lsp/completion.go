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
	ContextRoot
	ContextSecretObject
	ContextKeyObject
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
	fieldName     string
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

	if request.fieldName != "" {
		switch parsedCtx.Type {
		case ContextRoot:
			return p.completeValue(request, parsedCtx, GetRootFieldVals())
		case ContextSecretObject, ContextSecretsList:
			return p.completeValue(request, parsedCtx, GetSecretFieldVals())
		case ContextKeyObject, ContextKeysList:
			return p.completeValue(request, parsedCtx, GetKeyFieldVals())
		}
		return CompletionList{}
	}

	switch parsedCtx.Type {
	case ContextSecretsList:
		return p.completeSecrets(request)
	case ContextKeysList:
		return p.completeKeys(request, parsedCtx)
	case ContextRoot:
		return p.completeRoot(request, parsedCtx)
	case ContextSecretObject:
		return p.completeSecretObject(request, parsedCtx)
	case ContextKeyObject:
		return p.completeKeyObject(request, parsedCtx)
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

	trimmedPrefix := ""
	fieldName := ""
	needsDash := !strings.HasPrefix(strings.TrimSpace(prefix), "-")

	// Check if we're typing a value after a colon
	colonIdx := strings.LastIndex(prefix, ":")
	if colonIdx != -1 {
		// Typing a value after a field name and colon
		beforeColon := prefix[:colonIdx]
		fieldNamePart := strings.TrimSpace(beforeColon)

		// Extract the field name (last word before colon)
		if idx := strings.LastIndexAny(fieldNamePart, " \t-"); idx != -1 {
			fieldName = strings.TrimSpace(fieldNamePart[idx+1:])
		} else {
			fieldName = strings.TrimLeft(fieldNamePart, " \t-")
		}

		// Get the value part (text after colon)
		afterColon := prefix[colonIdx+1:]
		trimmedPrefix = strings.TrimLeft(afterColon, " \t")
		trimmedPrefix = strings.TrimLeft(trimmedPrefix, "'\"")
	} else {
		// Not typing a value yet, still completing field names
		if strings.HasSuffix(prefix, " ") {
			trimmed := strings.TrimSpace(prefix)
			if trimmed != "" && trimmed != "-" {
				return completionRequest{}, false
			}
		}

		trimmedPrefix = strings.TrimLeft(prefix, " \t-")
		if idx := strings.LastIndexAny(trimmedPrefix, " :"); idx != -1 {
			trimmedPrefix = strings.TrimSpace(trimmedPrefix[idx+1:])
		}
		trimmedPrefix = strings.TrimLeft(trimmedPrefix, "'\"")
	}

	return completionRequest{
		params:        params,
		lines:         lines,
		trimmedPrefix: trimmedPrefix,
		needsDash:     needsDash,
		prefix:        prefix,
		fieldName:     fieldName,
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

func (p *CompletionProvider) completeRoot(request completionRequest, parsedCtx ParserContext) CompletionList {
	return p.completeSchemaFields(request, parsedCtx, rootFields)
}

func (p *CompletionProvider) completeSecretObject(request completionRequest, parsedCtx ParserContext) CompletionList {
	return p.completeSchemaFields(request, parsedCtx, secretFields)
}

func (p *CompletionProvider) completeKeyObject(request completionRequest, parsedCtx ParserContext) CompletionList {
	return p.completeSchemaFields(request, parsedCtx, keyFields)
}

func (p *CompletionProvider) completeSchemaFields(request completionRequest, parsedCtx ParserContext, fields []string) CompletionList {
	var items []CompletionItem
	for _, field := range fields {
		if parsedCtx.Existing[field] || !strings.HasPrefix(field, request.trimmedPrefix) {
			continue
		}

		items = append(items, newSchemaFieldCompletionItem(field, request, request.trimmedPrefix))
	}
	return CompletionList{Items: items}
}

func (p *CompletionProvider) completeValue(request completionRequest, parsedCtx ParserContext, fieldVals map[string][]string) CompletionList {
	vals, ok := fieldVals[request.fieldName]
	if !ok || len(vals) == 0 {
		return CompletionList{}
	}

	var items []CompletionItem
	for _, val := range vals {
		if !strings.HasPrefix(val, request.trimmedPrefix) {
			continue
		}
		items = append(items, newValueCompletionItem(val, request, request.trimmedPrefix))
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

func newSchemaFieldCompletionItem(label string, request completionRequest, currentWord string) CompletionItem {
	insertText := label + ": "
	wordStart := request.params.Position.Character - len(currentWord)

	return CompletionItem{
		Label:      label,
		Kind:       CompletionItemKindField,
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
		Command: &Command{
			Title:   "Trigger Suggest",
			Command: "editor.action.triggerSuggest",
		},
	}
}

func newValueCompletionItem(label string, request completionRequest, currentWord string) CompletionItem {
	insertText := label
	if strings.HasSuffix(request.prefix, ":") {
		insertText = " " + label
	}
	wordStart := request.params.Position.Character - len(currentWord)

	return CompletionItem{
		Label:      label,
		Kind:       CompletionItemKindValue,
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
		Command: nil,
	}
}
