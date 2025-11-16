// Copyright 2025 Steffen Busch

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// 	http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package i18n

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/caddyserver/caddy/v2"
	"go.uber.org/zap/zaptest"
)

func createTestDictFile(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	dictFile := filepath.Join(tmpDir, "test_dict.json")
	err := os.WriteFile(dictFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to create temp dict file: %v", err)
	}
	return dictFile
}

func TestI18nProvision(t *testing.T) {
	dictFile := createTestDictFile(t, `{
		"hello": {"de": "Hallo", "en": "Hello"},
		"goodbye": {"en": "Goodbye"}
	}`)

	i18n := &I18n{DictFile: dictFile}
	i18n.mu = new(sync.RWMutex)
	i18n.logger = zaptest.NewLogger(t)
	var stubCaddyCtx caddy.Context

	err := i18n.Provision(stubCaddyCtx)
	if err != nil {
		t.Fatalf("Provision failed: %v", err)
	}

	if i18n.translations == nil {
		t.Fatal("translations map not initialized")
	}
	if len(i18n.translations) != 2 {
		t.Errorf("expected 2 translation keys, got %d", len(i18n.translations))
	}
}

func TestI18nProvisionMissingFile(t *testing.T) {
	i18n := &I18n{DictFile: "/nonexistent/path/dict.json"}
	i18n.logger = zaptest.NewLogger(t)
	var stubCaddyCtx caddy.Context

	err := i18n.Provision(stubCaddyCtx)
	if err == nil {
		t.Fatal("expected error for missing dictionary file")
	}
}

func TestI18nProvisionEmptyDictFile(t *testing.T) {
	dictFile := createTestDictFile(t, `{}`)

	i18n := &I18n{DictFile: dictFile}
	i18n.logger = zaptest.NewLogger(t)
	var stubCaddyCtx caddy.Context

	err := i18n.Provision(stubCaddyCtx)
	if err != nil {
		t.Fatalf("Provision failed for empty dict: %v", err)
	}

	if len(i18n.translations) != 0 {
		t.Errorf("expected empty translations, got %d entries", len(i18n.translations))
	}
}

func TestI18nProvisionInvalidJSON(t *testing.T) {
	dictFile := createTestDictFile(t, `{invalid json}`)

	i18n := &I18n{DictFile: dictFile}
	i18n.logger = zaptest.NewLogger(t)
	var stubCaddyCtx caddy.Context

	err := i18n.Provision(stubCaddyCtx)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestI18nProvisionNoDictFile(t *testing.T) {
	i18n := &I18n{DictFile: ""}
	i18n.logger = zaptest.NewLogger(t)
	var stubCaddyCtx caddy.Context

	err := i18n.Provision(stubCaddyCtx)
	if err != nil {
		t.Fatalf("Provision should succeed with empty DictFile: %v", err)
	}

	if i18n.translations == nil {
		t.Fatal("translations map should be initialized")
	}
	if len(i18n.translations) != 0 {
		t.Errorf("expected 0 translations, got %d", len(i18n.translations))
	}
}

func TestI18nTranslateBasic(t *testing.T) {
	i18n := &I18n{
		translations: map[string]map[string]string{
			"hello":   {"de": "Hallo", "en": "Hello"},
			"welcome": {"de": "Willkommen", "en": "Welcome"},
		},
	}
	i18n.mu = new(sync.RWMutex)
	i18n.logger = zaptest.NewLogger(t)

	funcMap := i18n.CustomTemplateFunctions()
	translateFunc := funcMap["i18nTranslate"].(func(string, string, ...interface{}) (string, error))

	tests := []struct {
		key      string
		lang     string
		expected string
	}{
		{"hello", "de", "Hallo"},
		{"hello", "en", "Hello"},
		{"welcome", "de", "Willkommen"},
	}

	for _, tt := range tests {
		result, err := translateFunc(tt.key, tt.lang)
		if err != nil {
			t.Errorf("unexpected error for key %s: %v", tt.key, err)
		}
		if result != tt.expected {
			t.Errorf("key %s lang %s: expected %q, got %q", tt.key, tt.lang, tt.expected, result)
		}
	}
}

func TestI18nTranslateFallback(t *testing.T) {
	i18n := &I18n{
		translations: map[string]map[string]string{
			"hello": {"en": "Hello"},
		},
	}
	i18n.mu = new(sync.RWMutex)
	i18n.logger = zaptest.NewLogger(t)

	funcMap := i18n.CustomTemplateFunctions()
	translateFunc := funcMap["i18nTranslate"].(func(string, string, ...interface{}) (string, error))

	result, err := translateFunc("hello", "de")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "Hello" {
		t.Errorf("expected fallback to 'Hello', got %q", result)
	}
}

func TestI18nTranslateMissingKey(t *testing.T) {
	i18n := &I18n{
		translations: map[string]map[string]string{},
	}
	i18n.mu = new(sync.RWMutex)
	i18n.logger = zaptest.NewLogger(t)

	funcMap := i18n.CustomTemplateFunctions()
	translateFunc := funcMap["i18nTranslate"].(func(string, string, ...interface{}) (string, error))

	result, err := translateFunc("nonexistent", "en")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "nonexistent" {
		t.Errorf("expected key as fallback, got %q", result)
	}
}

func TestI18nTranslateMissingLanguage(t *testing.T) {
	i18n := &I18n{
		translations: map[string]map[string]string{
			"hello": {"fr": "Bonjour"},
		},
	}
	i18n.mu = new(sync.RWMutex)
	i18n.logger = zaptest.NewLogger(t)

	funcMap := i18n.CustomTemplateFunctions()
	translateFunc := funcMap["i18nTranslate"].(func(string, string, ...interface{}) (string, error))

	// Request non-existent language and no English fallback
	result, err := translateFunc("hello", "de")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// Should return the key as fallback when no English and no requested lang
	if result != "hello" {
		t.Errorf("expected key 'hello' as fallback, got %q", result)
	}
}

func TestI18nTranslateEmptyKey(t *testing.T) {
	i18n := &I18n{
		translations: map[string]map[string]string{},
	}
	i18n.mu = new(sync.RWMutex)
	i18n.logger = zaptest.NewLogger(t)

	funcMap := i18n.CustomTemplateFunctions()
	translateFunc := funcMap["i18nTranslate"].(func(string, string, ...interface{}) (string, error))

	result, err := translateFunc("", "en")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty string for empty key, got %q", result)
	}
}

func TestI18nInterpolateWithI18nPrefix(t *testing.T) {
	i18n := &I18n{
		translations: map[string]map[string]string{
			"account":             {"de": "Konto", "en": "Account"},
			"error.invalidAmount": {"de": "Ungültiger Betrag: {0}", "en": "Invalid amount: {0}"},
		},
	}
	i18n.mu = new(sync.RWMutex)
	i18n.logger = zaptest.NewLogger(t)

	funcMap := i18n.CustomTemplateFunctions()
	translateFunc := funcMap["i18nTranslate"].(func(string, string, ...interface{}) (string, error))

	result, err := translateFunc("error.invalidAmount", "en", "i18n:account")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "Invalid amount: Account" {
		t.Errorf("expected 'Invalid amount: Account', got %q", result)
	}

	result, err = translateFunc("error.invalidAmount", "de", "i18n:account")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "Ungültiger Betrag: Konto" {
		t.Errorf("expected 'Ungültiger Betrag: Konto', got %q", result)
	}
}

func TestI18nInterpolateWithoutPrefix(t *testing.T) {
	i18n := &I18n{
		translations: map[string]map[string]string{
			"error.user": {"de": "Fehler bei Benutzer: {0}", "en": "Error for user: {0}"},
		},
	}
	i18n.mu = new(sync.RWMutex)
	i18n.logger = zaptest.NewLogger(t)

	funcMap := i18n.CustomTemplateFunctions()
	translateFunc := funcMap["i18nTranslate"].(func(string, string, ...interface{}) (string, error))

	result, err := translateFunc("error.user", "en", "Hans Mueller")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "Error for user: Hans Mueller" {
		t.Errorf("expected 'Error for user: Hans Mueller', got %q", result)
	}
}

func TestI18nInterpolateMultipleArgs(t *testing.T) {
	i18n := &I18n{
		translations: map[string]map[string]string{
			"error":  {"de": "Fehler: {0} bei {1}", "en": "Error: {0} at {1}"},
			"system": {"de": "System", "en": "System"},
			"module": {"de": "Modul", "en": "Module"},
		},
	}
	i18n.mu = new(sync.RWMutex)
	i18n.logger = zaptest.NewLogger(t)

	funcMap := i18n.CustomTemplateFunctions()
	translateFunc := funcMap["i18nTranslate"].(func(string, string, ...interface{}) (string, error))

	result, err := translateFunc("error", "en", "i18n:system", "i18n:module")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "Error: System at Module" {
		t.Errorf("expected 'Error: System at Module', got %q", result)
	}
}

func TestI18nInterpolateNonStringArg(t *testing.T) {
	i18n := &I18n{
		translations: map[string]map[string]string{
			"amount": {"de": "Betrag: {0}", "en": "Amount: {0}"},
		},
	}
	i18n.mu = new(sync.RWMutex)
	i18n.logger = zaptest.NewLogger(t)

	funcMap := i18n.CustomTemplateFunctions()
	translateFunc := funcMap["i18nTranslate"].(func(string, string, ...interface{}) (string, error))

	result, err := translateFunc("amount", "en", 123.45)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "Amount: 123.45" {
		t.Errorf("expected 'Amount: 123.45', got %q", result)
	}
}

func TestI18nInterpolateInvalidPlaceholder(t *testing.T) {
	i18n := &I18n{
		translations: map[string]map[string]string{
			"msg": {"en": "Value: {0} and {5}"},
		},
	}
	i18n.mu = new(sync.RWMutex)
	i18n.logger = zaptest.NewLogger(t)

	funcMap := i18n.CustomTemplateFunctions()
	translateFunc := funcMap["i18nTranslate"].(func(string, string, ...interface{}) (string, error))

	// Placeholder {5} is out of range, should remain unchanged
	result, err := translateFunc("msg", "en", "hello")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "Value: hello and {5}" {
		t.Errorf("expected 'Value: hello and {5}', got %q", result)
	}
}

func TestI18nInterpolateNestedI18nPrefix(t *testing.T) {
	i18n := &I18n{
		translations: map[string]map[string]string{
			"account": {"de": "Konto", "en": "Account"},
			"msg":     {"de": "Typ: {0}", "en": "Type: {0}"},
		},
	}
	i18n.mu = new(sync.RWMutex)
	i18n.logger = zaptest.NewLogger(t)

	funcMap := i18n.CustomTemplateFunctions()
	translateFunc := funcMap["i18nTranslate"].(func(string, string, ...interface{}) (string, error))

	// Test nested translation with i18n: prefix
	result, err := translateFunc("msg", "de", "i18n:account")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "Typ: Konto" {
		t.Errorf("expected 'Typ: Konto', got %q", result)
	}
}

func TestI18nInterpolateUnknownI18nKey(t *testing.T) {
	i18n := &I18n{
		translations: map[string]map[string]string{
			"msg": {"en": "Value: {0}"},
		},
	}
	i18n.mu = new(sync.RWMutex)
	i18n.logger = zaptest.NewLogger(t)

	funcMap := i18n.CustomTemplateFunctions()
	translateFunc := funcMap["i18nTranslate"].(func(string, string, ...interface{}) (string, error))

	// i18n: prefix with non-existent key should return the key as fallback
	result, err := translateFunc("msg", "en", "i18n:nonexistent")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "Value: nonexistent" {
		t.Errorf("expected 'Value: nonexistent', got %q", result)
	}
}

func TestI18nInterpolateMixedArgs(t *testing.T) {
	i18n := &I18n{
		translations: map[string]map[string]string{
			"error":  {"en": "Error: {0} with code {1} for user {2}"},
			"system": {"en": "System"},
		},
	}
	i18n.mu = new(sync.RWMutex)
	i18n.logger = zaptest.NewLogger(t)

	funcMap := i18n.CustomTemplateFunctions()
	translateFunc := funcMap["i18nTranslate"].(func(string, string, ...interface{}) (string, error))

	// Mix of i18n: prefix and regular values
	result, err := translateFunc("error", "en", "i18n:system", 500, "alice")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "Error: System with code 500 for user alice" {
		t.Errorf("expected 'Error: System with code 500 for user alice', got %q", result)
	}
}

func TestI18nInterpolateWithFloat(t *testing.T) {
	i18n := &I18n{
		translations: map[string]map[string]string{
			"price": {"en": "Price: {0} EUR"},
		},
	}
	i18n.mu = new(sync.RWMutex)
	i18n.logger = zaptest.NewLogger(t)

	funcMap := i18n.CustomTemplateFunctions()
	translateFunc := funcMap["i18nTranslate"].(func(string, string, ...interface{}) (string, error))

	result, err := translateFunc("price", "en", 19.99)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "Price: 19.99 EUR" {
		t.Errorf("expected 'Price: 19.99 EUR', got %q", result)
	}
}

func TestI18nInterpolateWithInt(t *testing.T) {
	i18n := &I18n{
		translations: map[string]map[string]string{
			"count": {"en": "Items: {0}"},
		},
	}
	i18n.mu = new(sync.RWMutex)
	i18n.logger = zaptest.NewLogger(t)

	funcMap := i18n.CustomTemplateFunctions()
	translateFunc := funcMap["i18nTranslate"].(func(string, string, ...interface{}) (string, error))

	result, err := translateFunc("count", "en", 42)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "Items: 42" {
		t.Errorf("expected 'Items: 42', got %q", result)
	}
}

func TestI18nNoArgs(t *testing.T) {
	i18n := &I18n{
		translations: map[string]map[string]string{
			"hello": {"de": "Hallo Welt", "en": "Hello World"},
		},
	}
	i18n.mu = new(sync.RWMutex)
	i18n.logger = zaptest.NewLogger(t)

	funcMap := i18n.CustomTemplateFunctions()
	translateFunc := funcMap["i18nTranslate"].(func(string, string, ...interface{}) (string, error))

	result, err := translateFunc("hello", "de")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "Hallo Welt" {
		t.Errorf("expected 'Hallo Welt', got %q", result)
	}
}

func TestI18nFallbackToEnglish(t *testing.T) {
	i18n := &I18n{
		translations: map[string]map[string]string{
			"welcome": {"en": "Welcome"},
		},
	}
	i18n.mu = new(sync.RWMutex)
	i18n.logger = zaptest.NewLogger(t)

	funcMap := i18n.CustomTemplateFunctions()
	translateFunc := funcMap["i18nTranslate"].(func(string, string, ...interface{}) (string, error))

	// Request any language that doesn't exist, should fallback to English
	result, err := translateFunc("welcome", "it")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "Welcome" {
		t.Errorf("expected 'Welcome' (English fallback), got %q", result)
	}
}

func TestI18nCaddyModule(t *testing.T) {
	i18n := &I18n{}
	modInfo := i18n.CaddyModule()

	if modInfo.ID != "http.handlers.templates.functions.i18n" {
		t.Errorf("expected ID 'http.handlers.templates.functions.i18n', got %q", modInfo.ID)
	}
	if modInfo.New == nil {
		t.Fatal("expected New function to be set")
	}
}
