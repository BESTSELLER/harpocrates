package lsp

import (
	"strings"
)

type ParserContext struct {
	Type         CompletionContext
	ParentSecret string
	Existing     map[string]bool
}

func parseContext(lines []string, targetLine int) ParserContext {
	result := ParserContext{
		Type:     ContextUnknown,
		Existing: make(map[string]bool),
	}

	if targetLine < 0 || targetLine >= len(lines) {
		return result
	}

	// Calculate target indent
	currentLine := lines[targetLine]
	targetIndent := getIndentCount(currentLine)

	// Special case: root level (indent == 0)
	if targetIndent == 0 {
		result.Type = ContextRoot
		// Collect existing root keys
		for i := 0; i < len(lines); i++ {
			if i == targetLine {
				continue
			}
			line := lines[i]
			if getIndentCount(line) == 0 && strings.Contains(line, ":") {
				key := extractKeyFromLine(line)
				if key != "" {
					result.Existing[key] = true
				}
			}
		}
		return result
	}

	// Scan up to find nearest block with a smaller indent
	blockLineIdx := -1
	blockIndent := -1
	currentIndentLimit := targetIndent
	var blockType string

	for i := targetLine - 1; i >= 0; i-- {
		line := lines[i]
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := getIndentCount(line)

		if indent < currentIndentLimit {
			trimmedLine := strings.TrimSpace(line)

			if strings.HasPrefix(trimmedLine, "secrets:") {
				result.Type = ContextSecretsList
				blockLineIdx = i
				blockIndent = indent
				blockType = "secrets:"
				break
			}

			if strings.HasPrefix(trimmedLine, "keys:") {
				result.Type = ContextKeysList
				blockLineIdx = i
				blockIndent = indent
				blockType = "keys:"

				// Find parent secret path
				for j := i - 1; j >= 0; j-- {
					pIndent := getIndentCount(lines[j])
					if pIndent < indent {
						pLine := strings.TrimSpace(lines[j])
						if strings.HasPrefix(pLine, "-") {
							result.ParentSecret = extractValFromList(pLine)
						}
						break
					}
				}
				break
			}

			// If we hit a line starting with "-", we might be in a secret or key object
			if strings.HasPrefix(trimmedLine, "-") {
				// Look ahead to find if "keys:" block is present in this secret's scope
				hasKeysBlock := false
				keysBlockIndent := -1
				for j := i + 1; j < targetLine; j++ {
					jIndent := getIndentCount(lines[j])
					jTrimmed := strings.TrimSpace(lines[j])
					if jIndent <= indent {
						break
					}
					if strings.HasPrefix(jTrimmed, "keys:") {
						hasKeysBlock = true
						keysBlockIndent = jIndent
						break
					}
				}

				if hasKeysBlock && targetIndent > keysBlockIndent {
					result.Type = ContextKeyObject
				} else {
					result.Type = ContextSecretObject
				}
				blockLineIdx = i
				blockIndent = indent
				result.ParentSecret = extractValFromList(trimmedLine)
				blockType = "object:"
				break
			}

			// If we hit any other block we are inside that block
			if strings.Contains(trimmedLine, ":") {
				break
			}

			currentIndentLimit = indent
		}
	}

	if result.Type == ContextUnknown {
		return result
	}

	// Scan down to collect existing items
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

		switch blockType {
		case "secrets:", "keys:":
			if strings.HasPrefix(trimmed, "-") {
				val := extractValFromList(trimmed)
				if val != "" {
					result.Existing[val] = true
				}
			}
		case "object:":
			// Only collect keys at the exact expected indent for objects
			if indent == blockIndent+2 && strings.Contains(line, ":") {
				key := extractKeyFromLine(line)
				if key != "" {
					result.Existing[key] = true
				}
			}
		}
	}

	return result
}

func getIndentCount(line string) int {
	count := 0
	for _, ch := range line {
		switch ch {
		case ' ':
			count++
		case '\t':
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
	val = strings.Trim(val, "'\"")
	return val
}

// extractKeyFromLine extracts the key name from a YAML line like "  fieldName: value".
func extractKeyFromLine(line string) string {
	trimmed := strings.TrimSpace(line)
	if idx := strings.Index(trimmed, ":"); idx != -1 {
		key := strings.TrimSpace(trimmed[:idx])
		return key
	}
	return ""
}
