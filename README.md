# Caddy i18n Template Extension

A Caddy v2 template extension module for internationalization (i18n) support. This module enables translation of content based on language codes with support for nested translations, language fallbacks, and argument interpolation.

## Features

- **Dictionary-Based Translations**: Load translations from JSON files
- **Language Fallbacks**: Automatically falls back to English if requested language is unavailable
- **Nested Translations**: Use translation keys as arguments with `i18n:` prefix
- **Argument Interpolation**: Replace placeholders `{0}`, `{1}`, etc. with provided values
- **Thread-Safe**: Protected concurrent access to translations with RWMutex
- **Logging**: Informational and warning logs for debugging

## Installation

Build the module as part of your Caddy binary using `xcaddy`:

```bash
xcaddy build --with github.com/steffenbusch/caddy-i18n-template-ext
```

## Configuration

### Caddyfile Syntax

```caddyfile
:8080 {
    root ./demo/html
    templates {
        extensions {
            i18n {
                dict_file ./demo/translations.json
            }
        }
    }
    file_server
}
```

### JSON Dictionary Format

```json
{
  "hello": {
    "de": "Hallo",
    "en": "Hello"
  },
  "error.invalidAmount": {
    "de": "Ungültiger Betrag: {0}",
    "en": "Invalid amount: {0}"
  },
  "finance.account": {
    "de": "Konto",
    "en": "Account"
  }
}
```

## Usage

### Basic Translation

```html
{{ i18nTranslate "hello" "de" }}
<!-- Output: Hallo -->
```

### With Literal Arguments

```html
{{ i18nTranslate "error.invalidAmount" "en" "500.99" }}
<!-- Output: Invalid amount: 500.99 -->
```

### With Nested Translations

```html
{{ i18nTranslate "error.invalidAmount" "de" "i18n:finance.account" }}
<!-- Output: Ungültiger Betrag: Konto -->
```

### With Multiple Arguments

```html
{{ i18nTranslate "error.internalError" "en" "i18n:system" "i18n:module" }}
<!-- Template: "Error: {0} at {1}" -->
<!-- Output: Error: System at Module -->
```

### Using Variables

```html
{{- $lang := "de" -}}
{{ i18nTranslate "welcome" $lang }}
```

## Language Fallback Behavior

1. **First**: Try to find the translation for the requested language
2. **Second**: Fall back to English ("en") if the requested language is unavailable
3. **Third**: Return the translation key itself if neither the requested language nor English exists

Each fallback is logged for debugging purposes.

## Error Handling

- Missing dictionary files return an error during provisioning
- Invalid JSON in dictionary files is reported with detailed error messages
- Unknown translation keys are logged as warnings and the key is returned as fallback
- Placeholder indices outside the argument range remain unchanged in the output

## Example Complete Configuration

```caddyfile
{
    http_port 8080
}

:8080 {
    route /api/* {
        templates {
            functions {
                i18n {
                    dict_file /etc/caddy/translations.json
                }
            }
        }
        respond `{{ i18nTranslate "welcome" "de" }}`
    }
}
```

## License

This project is licensed under the Apache License, Version 2.0. See the [LICENSE](LICENSE) file for details.

## Acknowledgements

- [Caddy](https://caddyserver.com) for providing a powerful and extensible web server.
