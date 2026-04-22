package lsp

import (
	"strings"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
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

	// Clean up prefix to just handle the current word for matching
	trimmedPrefix := strings.TrimLeft(prefix, " \t-")
	if idx := strings.LastIndexAny(trimmedPrefix, " :"); idx != -1 {
		trimmedPrefix = strings.TrimSpace(trimmedPrefix[idx+1:])
	}
	trimmedPrefix = strings.TrimLeft(trimmedPrefix, "'\"")

	needsDash := !strings.HasPrefix(strings.TrimSpace(prefix), "-")

	ctx, parentSecret := s.determineContext(content, params.Position.Line)

	if ctx == ContextSecretsList {
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

		// Try to list secrets
		tokens, err := s.vaultClient.ListTokens(queryPath)
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

				items = append(items, CompletionItem{
					Label:      token,
					Kind:       kind,
					InsertText: insertText,
					TextEdit: &TextEdit{
						Range: Range{
							Start: params.Position,
							End:   params.Position,
						},
						NewText: insertText,
					},
					Command: cmd,
				})
			}
		}
		return CompletionList{Items: items}
	}

	if ctx == ContextKeysList && parentSecret != "" {
		if s.vaultClient == nil {
			return CompletionList{}
		}

		// Try to read secret keys
		secretData, err := s.vaultClient.ReadSecret(parentSecret)
		if err != nil || secretData == nil {
			return CompletionList{}
		}

		existingKeys := getExistingItemsInBlock(lines, params.Position.Line)

		var items []CompletionItem
		for key := range secretData {
			// Skip if already added
			if existingKeys[key] {
				continue
			}

			// Prefix match
			if strings.HasPrefix(key, trimmedPrefix) {
				insertText := key
				if needsDash {
					insertText = "- " + insertText
				}

				items = append(items, CompletionItem{
					Label:      key,
					Kind:       CompletionItemKindField,
					InsertText: insertText,
					TextEdit: &TextEdit{
						Range: Range{
							Start: params.Position,
							End:   params.Position,
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

// determineContext figures out if we are under `secrets:` or `keys:`
func (s *Server) determineContext(content string, targetLine int) (CompletionContext, string) {
	var parentSecret string
	var node yaml.Node
	err := yaml.Unmarshal([]byte(content), &node)

	// Either AST parsing succeeded entirely or partially.
	// Try parsing AST first to extract parent secret if possible.
	if err == nil {
		path := findContextPath(&node, targetLine+1, nil) // targetLine is 0-indexed in LSP, yaml Node is 1-indexed
		if inKeysList(path) {
			parentSecret = extractParentSecret(path)
		}
	}

	lines := strings.Split(content, "\n")
	if targetLine < 0 || targetLine >= len(lines) {
		return ContextUnknown, ""
	}

	// Validate the AST context (or find it if AST failed) using indentation
	currentIndent := getIndentCount(lines[targetLine])

	// Scan up to find nearest block with a smaller indent
	for i := targetLine - 1; i >= 0; i-- {
		line := lines[i]
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := getIndentCount(line)

		if indent < currentIndent {
			trimmedLine := strings.TrimSpace(line)

			// If we hit the expected block and the cursor is truly indented inside it
			if strings.HasPrefix(trimmedLine, "secrets:") {
				return ContextSecretsList, ""
			}
			if strings.HasPrefix(trimmedLine, "keys:") {
				// Find parent secret path
				for j := i - 1; j >= 0; j-- {
					if getIndentCount(lines[j]) < indent {
						pLine := strings.TrimSpace(lines[j])
						if strings.HasPrefix(pLine, "-") {
							return ContextKeysList, extractValFromList(pLine)
						}
						break
					}
				}
				return ContextKeysList, parentSecret
			}

			// If we hit any other block (like a list item "- secret/data/foo:" or "format:"),
			// we are inside that block, NOT directly inside secrets: or keys:
			if strings.HasPrefix(trimmedLine, "-") || strings.Contains(trimmedLine, ":") {
				// We reached a different parent block.
				// The cursor is inside THIS block, not directly in `secrets:` or `keys:`
				break
			}

			// Update current indent limit
			currentIndent = indent
		}
	}

	return ContextUnknown, ""
}

func getExistingItemsInBlock(lines []string, targetLine int) map[string]bool {
	existing := make(map[string]bool)
	if targetLine < 0 || targetLine >= len(lines) {
		return existing
	}

	blockLineIdx := -1
	blockIndent := -1
	currentIndent := getIndentCount(lines[targetLine])

	// 1. Scan up to find the nearest keys: or secrets:
	for i := targetLine; i >= 0; i-- {
		line := lines[i]
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := getIndentCount(line)

		if indent < currentIndent || i == targetLine {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "keys:") || strings.HasPrefix(trimmed, "secrets:") {
				blockLineIdx = i
				blockIndent = indent
				break
			}
			currentIndent = indent
		}
	}

	if blockLineIdx == -1 {
		return existing
	}

	// 2. Scan down to collect existing items
	for i := blockLineIdx + 1; i < len(lines); i++ {
		if i == targetLine {
			continue // Skip the line currently being typed
		}
		line := lines[i]
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := getIndentCount(line)

		// Found the end of the block
		if indent <= blockIndent {
			break
		}

		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "-") {
			val := extractValFromList(trimmed)
			val = strings.TrimSpace(val)
			val = strings.Trim(val, "'\"")
			existing[val] = true
		}
	}

	return existing
}

func getIndentCount(line string) int {
	count := 0
	for _, ch := range line {
		switch ch {
		case ' ':
			count++
		case '\t': // Tabs not standard for yaml, but fallback support
			count += 2
		default:
			return count
		}
	}
	return count
}

func extractValFromList(line string) string {
	val := strings.TrimPrefix(line, "-")
	val = strings.TrimSpace(val)
	val = strings.TrimSuffix(val, ":")
	return val
}

// AST Path Finding Logic

func findContextPath(node *yaml.Node, targetLine int, currentPath []string) []string {
	if node == nil || node.Line > targetLine {
		return currentPath
	}

	switch node.Kind {
	case yaml.DocumentNode:
		if len(node.Content) > 0 {
			return findContextPath(node.Content[0], targetLine, currentPath)
		}

	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valNode := node.Content[i+1]

			nextKeyLine := 9999999
			if i+2 < len(node.Content) {
				nextKeyLine = node.Content[i+2].Line
			}

			if targetLine >= keyNode.Line && targetLine < nextKeyLine {
				newPath := append(currentPath, keyNode.Value)
				// If the value is a map, its elements own the line
				if valNode.Line <= targetLine {
					return findContextPath(valNode, targetLine, newPath)
				}
				return newPath
			}
		}

	case yaml.SequenceNode:
		for i, item := range node.Content {
			nextItemLine := 9999999
			if i+1 < len(node.Content) {
				nextItemLine = node.Content[i+1].Line
			}

			if targetLine >= item.Line && targetLine < nextItemLine {
				// use "[]" to signify a list element, followed by the item map keys
				newPath := append(currentPath, "[]")

				// To extract parent secrets properly we need string representation
				// In yaml `- secret/data/path:` the string is stored weirdly so we dump the raw
				if len(item.Content) > 0 && item.Content[0].Value != "" {
					newPath[len(newPath)-1] = item.Content[0].Value
				}

				return findContextPath(item, targetLine, newPath)
			}
		}
	}

	return currentPath
}

func inSecretsList(path []string) bool {
	if len(path) == 0 {
		return false
	}
	// "secrets", "[]"
	return path[0] == "secrets" && (len(path) == 1 || len(path) == 2)
}

func inKeysList(path []string) bool {
	if len(path) < 3 {
		return false
	}
	// "secrets", "<parent-secret>", "keys", ...
	for i := 0; i < len(path)-1; i++ {
		if path[i] == "keys" {
			return true
		}
	}
	return false
}

func extractParentSecret(path []string) string {
	if len(path) >= 3 && path[0] == "secrets" {
		return path[1]
	}
	return ""
}
