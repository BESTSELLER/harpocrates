package lsp

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseContext(t *testing.T) {
	tests := []struct {
		name             string
		document         string
		targetLine       int
		wantType         CompletionContext
		wantParentSecret string
		wantExisting     map[string]bool
	}{
		{
			name: "secrets list item",
			document: strings.Join([]string{
				"secrets:",
				"  - app/data/config",
				"  - app/data/",
			}, "\n"),
			targetLine:   2,
			wantType:     ContextSecretsList,
			wantExisting: map[string]bool{"app/data/config": true},
		},
		{
			name: "keys list item",
			document: strings.Join([]string{
				"secrets:",
				"  - app/data/config:",
				"      keys:",
				"        - username",
				"        - pass",
			}, "\n"),
			targetLine:       4,
			wantType:         ContextKeysList,
			wantParentSecret: "app/data/config",
			wantExisting:     map[string]bool{"username": true},
		},
		{
			name: "quoted parent and existing value",
			document: strings.Join([]string{
				"secrets:",
				"  - 'app/data/config':",
				"      keys:",
				"        - \"username\"",
				"        - ",
			}, "\n"),
			targetLine:       4,
			wantType:         ContextKeysList,
			wantParentSecret: "app/data/config",
			wantExisting:     map[string]bool{"username": true},
		},
		{
			name: "ignores blank lines",
			document: strings.Join([]string{
				"secrets:",
				"",
				"  - shared/data/config",
				"",
				"  - shared/data/",
			}, "\n"),
			targetLine:   4,
			wantType:     ContextSecretsList,
			wantExisting: map[string]bool{"shared/data/config": true},
		},
		{
			name: "secret object inside nested option block",
			document: strings.Join([]string{
				"secrets:",
				"  - app/data/config:",
				"      format: env",
			}, "\n"),
			targetLine:       2,
			wantType:         ContextSecretObject,
			wantParentSecret: "app/data/config",
			wantExisting:     map[string]bool{},
		},
		{
			name:         "out of range",
			document:     "secrets:\n  - app/data/config",
			targetLine:   5,
			wantType:     ContextUnknown,
			wantExisting: map[string]bool{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseContext(strings.Split(tt.document, "\n"), tt.targetLine)
			if got.Type != tt.wantType {
				t.Fatalf("Type = %v, want %v", got.Type, tt.wantType)
			}
			if got.ParentSecret != tt.wantParentSecret {
				t.Fatalf("ParentSecret = %q, want %q", got.ParentSecret, tt.wantParentSecret)
			}
			if !reflect.DeepEqual(got.Existing, tt.wantExisting) {
				t.Fatalf("Existing = %#v, want %#v", got.Existing, tt.wantExisting)
			}
		})
	}
}

func TestGetIndentCount(t *testing.T) {
	tests := map[string]int{
		"no-indent": 0,
		"  two":     2,
		"\ttab":     2,
		" \tmixed":  3,
	}

	for input, want := range tests {
		if got := getIndentCount(input); got != want {
			t.Fatalf("getIndentCount(%q) = %d, want %d", input, got, want)
		}
	}
}

func TestExtractValFromList(t *testing.T) {
	tests := map[string]string{
		"- app/data/config":     "app/data/config",
		"- app/data/config:":    "app/data/config",
		"- 'app/data/config'":   "app/data/config",
		"- \"app/data/config\"": "app/data/config",
	}

	for input, want := range tests {
		if got := extractValFromList(input); got != want {
			t.Fatalf("extractValFromList(%q) = %q, want %q", input, got, want)
		}
	}
}
