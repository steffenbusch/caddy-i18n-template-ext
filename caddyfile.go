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
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

// UnmarshalCaddyfile deserializes Caddyfile tokens into the I18n struct.
// It is called by Caddy when parsing Caddyfile configuration blocks.
//
// Syntax:
//
//	i18n {
//	    dict_file <path/to/dictionary.json>
//	}
//
// Parameters:
//   - dict_file: Path to the JSON file containing translation dictionaries (required)
//
// Example:
//
//	i18n {
//	    dict_file /etc/caddy/translations.json
//	}
func (i *I18n) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "dict_file":
				if !d.NextArg() {
					return d.ArgErr()
				}
				i.DictFile = d.Val()
				if d.NextArg() {
					return d.ArgErr()
				}

			default:
				return d.Errf("unrecognized i18n config property: %s", d.Val())
			}
		}
	}
	return nil
}

// Interface guard ensures that I18n implements caddyfile.Unmarshaler.
var _ caddyfile.Unmarshaler = (*I18n)(nil)
