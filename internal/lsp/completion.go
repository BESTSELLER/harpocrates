package lsp

import (
	"strings"

	"github.com/rs/zerolog/log"
)

// CompletionContext tells us where we are in the YAML document
type CompletionContext int

const (
	ContextUnknown CompletionContext = iota
	ContextSecretsList
	ContextKeysList
)

func (s *Server) provideCompletions(params CompletionParams) CompletionList {
	content, ok := s.documents[params.TextDocument.URI]
	if !ok {
		return CompletionList{}
	}

	lines := strings.Split(content, "\n")
	if params.Position.Line >= len(lines) {
		return CompletionList{}
	}
	currentLine := lines[params.Position.Line]

	// Extract prefix we have typed so far on this line
	prefix := ""
	if params.Position.Character <= len(currentLine) {
		prefix = currentLine[:params.Position.Character]
	}

	if strings.HasSuffix(strings.TrimSpace(prefix), ":") {
		return CompletionList{}
	}

	// Clean up prefix to just handle the current word for matching
	trimmedPrefix := strings.TrimLeft(prefix, " \t-")
	if idx := strings.LastIndexAny(trimmedPrefix, " :"); idx != -1 {
		trimmedPrefix = strings.TrimSpace(trimmedPrefix[idx+1:])
	}
	trimmedPrefix = strings.TrimLeft(trimmedPrefix, "'\"")

	needsDash := !strings.HasPrefix(strings.TrimSpace(prefix), "-")

	parsedCtx := parseContext(lines, params.Position.Line)

	if parsedCtx.Type == ContextSecretsList {
		if s.vaultClient == nil {
			return CompletionList{}
		}

		// Find the closest slash '/' to determine the "base path"
		basePath := ""
		if idx := strings.LastIndex(trimmedPrefix, "/"); idx != -1 {
			basePath = trimmedPrefix[:idx+1]
		}

		// The Vault V2 KV engine list method expects /metadata/ to list child secrets/directories.
		queryPath := strings.Replace(basePath, "/data/", "/metadata/", 1)

		var tokens []string
		cacheKey := "list:" + queryPath
		if cached, ok := s.vaultCache.Get(cacheKey); ok {
			tokens = cached.([]string)
		} else {
			// Try to list secrets
			var err error
			tokens, err = s.vaultClient.ListTokens(queryPath)
			if err != nil {
				log.Error().Err(err).Str("path", queryPath).Msg("ListTokens failed")
			}

			// At root level, also list secret engines
			if basePath == "" {
				engines, err := s.vaultClient.ListSecretEngines()
				if err != nil {
					log.Error().Err(err).Msg("ListSecretEngines failed")
				} else {
					tokens = append(tokens, engines...)
				}
			} else {
				// suggest engine sub-paths (like data/, roles/)
				subPath, err := s.vaultClient.GetEngineSubPath(basePath)
				if err != nil {
					log.Error().Err(err).Str("path", basePath).Msg("GetEngineSubPath failed")
				} else if subPath != "" {
					// Avoid duplicate if the engine already returned it in ListTokens
					exists := false
					for _, t := range tokens {
						// tokens might be relative like "data/", but subPath contains basePath too
						if basePath+t == subPath {
							exists = true
							break
						}
					}
					if !exists {
						// token in ListTokens is just the name (e.g. "data/"), so we must trim basePath
						tokens = append(tokens, strings.TrimPrefix(subPath, basePath))
					}
				}
			}
			s.vaultCache.Set(cacheKey, tokens)
		}

		var items []CompletionItem
		for _, token := range tokens {
			kind := CompletionItemKindValue
			if strings.HasSuffix(token, "/") {
				kind = CompletionItemKindFolder
			}

			// prefix matching to filter out items not matching the current word
			currentWord := strings.TrimPrefix(trimmedPrefix, basePath)
			if strings.HasPrefix(token, currentWord) {
				var cmd *Command
				if kind == CompletionItemKindFolder {
					cmd = &Command{
						Title:   "Trigger Suggest",
						Command: "editor.action.triggerSuggest",
					}
				}

				insertText := token
				if needsDash {
					insertText = "- " + insertText
				}

				wordStart := params.Position.Character - len(currentWord)

				items = append(items, CompletionItem{
					Label:      token,
					Kind:       kind,
					InsertText: insertText,
					TextEdit: &TextEdit{
						Range: Range{
							Start: Position{
								Line:      params.Position.Line,
								Character: wordStart,
							},
							End: params.Position,
						},
						NewText: insertText,
					},
					Command: cmd,
				})
			}
		}
		return CompletionList{Items: items}
	}

	if parsedCtx.Type == ContextKeysList && parsedCtx.ParentSecret != "" {
		if s.vaultClient == nil {
			return CompletionList{}
		}

		var secretData map[string]interface{}
		cacheKey := "read:" + parsedCtx.ParentSecret
		if cached, ok := s.vaultCache.Get(cacheKey); ok {
			secretData = cached.(map[string]interface{})
		} else {
			// Try to read secret keys
			var err error
			secretData, err = s.vaultClient.ReadSecret(parsedCtx.ParentSecret)
			if err != nil || secretData == nil {
				return CompletionList{}
			}
			s.vaultCache.Set(cacheKey, secretData)
		}

		var items []CompletionItem
		for key := range secretData {
			// Skip if already added
			if parsedCtx.Existing[key] {
				continue
			}

			// Prefix match
			if strings.HasPrefix(key, trimmedPrefix) {
				insertText := key
				if needsDash {
					insertText = "- " + insertText
				}

				wordStart := params.Position.Character - len(trimmedPrefix)

				items = append(items, CompletionItem{
					Label:      key,
					Kind:       CompletionItemKindField,
					InsertText: insertText,
					TextEdit: &TextEdit{
						Range: Range{
							Start: Position{
								Line:      params.Position.Line,
								Character: wordStart,
							},
							End: params.Position,
						},
						NewText: insertText,
					},
				})
			}
		}
		return CompletionList{Items: items}
	}

	return CompletionList{}
}
