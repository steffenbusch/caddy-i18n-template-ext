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
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"text/template"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp/templates"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(I18n{})
}

// I18n implements a simple internationalization (i18n) template extension for Caddy v2.
// It loads translation dictionaries from a JSON file and provides template functions
// for dictionary-based translation lookups with support for nested translations and
// language fallbacks.
//
// Example JSON structure:
//
//	{
//	  "hello": {
//	    "de": "Hallo",
//	    "en": "Hello"
//	  },
//	  "error.invalidAmount": {
//	    "de": "Ungültiger Betrag: {0}",
//	    "en": "Invalid amount: {0}"
//	  }
//	}
//
// Template usage:
//
//	{{ i18nTranslate "hello" "de" }}
//	{{ i18nTranslate "error.invalidAmount" "en" "i18n:finance.account" }}
type I18n struct {
	// DictFile is the path to the translations dictionary file in JSON format.
	// Structure: map[translationKey]map[languageCode]translatedText
	// Example: "/etc/caddy/translations.json"
	DictFile string `json:"dict_file,omitempty"`

	// translations holds the in-memory translation dictionary.
	// Structure: map[translationKey]map[languageCode]translatedText
	translations map[string]map[string]string

	// mu protects concurrent access to the translations map.
	mu *sync.RWMutex

	// logger is the Caddy logger instance for logging warnings and info messages.
	logger *zap.Logger
}

// CaddyModule returns the Caddy module information for registration.
func (I18n) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.templates.functions.i18n",
		New: func() caddy.Module { return new(I18n) },
	}
}

// Provision initializes the i18n extension by loading the translations dictionary
// from the configured JSON file. It is called during Caddy's provisioning phase.
func (i *I18n) Provision(ctx caddy.Context) error {
	i.logger = ctx.Logger()

	if i.mu == nil {
		i.mu = &sync.RWMutex{}
	}

	// Initialize the translations map
	i.translations = make(map[string]map[string]string)

	// Load translations from the dictionary file if configured
	if i.DictFile != "" {
		if err := i.loadDictionary(); err != nil {
			return fmt.Errorf("failed to load i18n dictionary: %w", err)
		}
		i.logger.Info("i18n dictionary loaded successfully", zap.String("dict_file", i.DictFile))
	}

	return nil
}

// CustomTemplateFunctions returns a FuncMap with the i18nTranslate template function.
// This function is used within Caddy templates to translate messages based on language codes.
//
// Function signature: i18nTranslate(key string, lang string, args ...interface{}) string
//
// Parameters:
//   - key: The translation dictionary key (e.g., "welcome" or "error.invalidAmount")
//   - lang: The language code (e.g., "de", "en", "fr")
//   - args: Optional positional arguments for interpolation in the translation template.
//     Arguments prefixed with "i18n:" are translated recursively.
//
// Behavior:
//   - If key doesn't exist: Returns key as fallback, logs warning
//   - If language doesn't exist: Falls back to "en", logs info
//   - If "en" also doesn't exist: Returns key as fallback, logs warning
//   - Replaces {0}, {1}, etc. in translation with provided arguments
//
// Example:
//
//	{{ i18nTranslate "error.invalidAmount" "de" "500.99" }}
//	{{ i18nTranslate "error.account" "en" "i18n:finance.account" }}
func (i *I18n) CustomTemplateFunctions() template.FuncMap {
	return template.FuncMap{
		"i18nTranslate": func(key, lang string, args ...interface{}) (string, error) {
			i.mu.RLock()
			defer i.mu.RUnlock()

			// Check if the translation key exists
			entry, ok := i.translations[key]
			if !ok {
				// Log a warning and return the key itself as a sensible fallback
				if i.logger != nil {
					i.logger.Warn("translation key not found, using key as fallback", zap.String("key", key))
				}
				return key, nil
			}

			// If requested language exists, use it
			val, ok := entry[lang]
			if !ok {
				// Try English as fallback language
				val, ok = entry["en"]
				if !ok {
					// Final fallback: log warning and return key
					if i.logger != nil {
						i.logger.Warn(
							"no translation for requested language or 'en', using key as fallback",
							zap.String("key", key),
							zap.String("requested_lang", lang),
						)
					}
					return key, nil
				}
				if i.logger != nil {
					i.logger.Info(
						"requested language not found, falling back to 'en'",
						zap.String("key", key),
						zap.String("requested_lang", lang),
					)
				}
			}

			// Replace positional arguments {0}, {1}, etc. with provided arguments
			if len(args) > 0 {
				val = i.interpolateTranslations(val, lang, args)
			}

			return val, nil
		},
	}
}

// interpolateTranslations replaces placeholders in the template string with argument values.
// Placeholders are in the form {0}, {1}, etc., indexed from 0.
//
// Argument handling:
//   - Arguments starting with "i18n:" prefix are treated as translation keys and translated recursively
//   - Other string arguments are used as-is
//   - Non-string arguments are converted to strings using fmt.Sprint
//
// Example:
//
//	Template: "Error: {0} at {1}"
//	Args: []interface{}{"i18n:system", "i18n:module"}
//	Result: "Error: System at Module" (after translation)
func (i *I18n) interpolateTranslations(tmpl string, lang string, args []interface{}) string {
	// Regex to find placeholders like {0}, {1}, etc.
	re := regexp.MustCompile(`\{(\d+)\}`)

	result := re.ReplaceAllStringFunc(tmpl, func(match string) string {
		// Extract the number from {N}
		numStr := strings.Trim(match, "{}")
		idx, err := strconv.Atoi(numStr)
		if err != nil || idx >= len(args) {
			return match // Return unchanged if invalid index
		}

		arg := args[idx]

		// If the argument is a string, check if it should be translated
		if str, ok := arg.(string); ok {
			// Check for i18n: prefix indicating a translation key
			if strings.HasPrefix(str, "i18n:") {
				translationKey := strings.TrimPrefix(str, "i18n:")
				entry, exists := i.translations[translationKey]
				if exists {
					// Try requested language first
					if val, ok := entry[lang]; ok {
						return val
					}
					// Fallback to English
					if val, ok := entry["en"]; ok {
						return val
					}
				}
				// If no translation found, return the key as fallback
				return translationKey
			}
			// No i18n: prefix, return string as-is
			return str
		}

		// For other types, convert to string representation
		return fmt.Sprint(arg)
	})

	return result
}

// loadDictionary reads and parses the JSON translation dictionary file.
// The file must contain a JSON object with the structure:
// map[translationKey]map[languageCode]translatedText
func (i *I18n) loadDictionary() error {
	file, err := os.Open(i.DictFile)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("dictionary file not found: %s", i.DictFile)
		}
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	if err := decoder.Decode(&i.translations); err != nil {
		return fmt.Errorf("failed to parse JSON dictionary: %w", err)
	}

	return nil
}

// Interface guards ensure that I18n implements the required interfaces.
var (
	_ caddy.Provisioner         = (*I18n)(nil)
	_ templates.CustomFunctions = (*I18n)(nil)
)
