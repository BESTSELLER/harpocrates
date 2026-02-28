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

	ctx, parentSecret := s.determineContext(content, params.Position.Line)

	if ctx == ContextSecretsList {
		if s.vaultClient == nil {
			return CompletionList{}
		}

		// Try to list secrets
		// Assume trimming prefix correctly leaves exact vault path like secret/data/foo
		tokens, err := s.vaultClient.ListTokens(trimmedPrefix)
		if err != nil {
			log.Error().Err(err).Str("path", trimmedPrefix).Msg("ListTokens failed")
		}

		var items []CompletionItem
		for _, token := range tokens {
			kind := CompletionItemKindValue
			if strings.HasSuffix(token, "/") {
				kind = CompletionItemKindFolder
			}
			items = append(items, CompletionItem{
				Label:      token,
				Kind:       kind,
				InsertText: token,
			})
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

		var items []CompletionItem
		for key := range secretData {
			// Prefix match
			if strings.HasPrefix(key, trimmedPrefix) {
				items = append(items, CompletionItem{
					Label:      key,
					Kind:       CompletionItemKindField,
					InsertText: key,
				})
			}
		}
		return CompletionList{Items: items}
	}

	return CompletionList{}
}

// determineContext figures out if we are under `secrets:` or `keys:`
func (s *Server) determineContext(content string, targetLine int) (ctx CompletionContext, parentSecret string) {
	ctx = ContextUnknown

	var node yaml.Node
	err := yaml.Unmarshal([]byte(content), &node)

	// Either AST parsing succeeded entirely or partially.
	// Try parsing AST first.
	if err == nil {
		path := findContextPath(&node, targetLine+1, nil) // targetLine is 0-indexed in LSP, yaml Node is 1-indexed
		if inSecretsList(path) {
			return ContextSecretsList, ""
		}
		if inKeysList(path) {
			// We are in keys, find the parent secret path
			return ContextKeysList, extractParentSecret(path)
		}
	}

	// Fallback to line scanning heuristics if AST is broken due to typing
	lines := strings.Split(content, "\n")

	if targetLine < 0 || targetLine >= len(lines) {
		return
	}

	// Count indent of the current line
	currentIndent := getIndentCount(lines[targetLine])

	// Scan up to find nearest `keys:` or `secrets:` with a smaller indent
	for i := targetLine - 1; i >= 0; i-- {
		line := lines[i]
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := getIndentCount(line)

		if indent < currentIndent {
			trimmedLine := strings.TrimSpace(line)
			if strings.HasPrefix(trimmedLine, "secrets:") {
				ctx = ContextSecretsList
				return
			}
			if strings.HasPrefix(trimmedLine, "keys:") {
				ctx = ContextKeysList
				// Now find parent secret path
				for j := i - 1; j >= 0; j-- {
					if getIndentCount(lines[j]) < indent {
						// This should be the list item `- secret/data/m...:`
						pLine := strings.TrimSpace(lines[j])
						if strings.HasPrefix(pLine, "-") {
							parentSecret = extractValFromList(pLine)
						}
						break
					}
				}
				return
			}
			// Update current indent limit
			currentIndent = indent
		}
	}

	return
}

func getIndentCount(line string) int {
	count := 0
	for _, ch := range line {
		if ch == ' ' {
			count++
		} else if ch == '\t' { // Tabs not standard for yaml, but fallback support
			count += 2
		} else {
			break
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
