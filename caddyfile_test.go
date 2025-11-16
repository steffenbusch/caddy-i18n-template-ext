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
	"strings"
	"testing"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

func TestUnmarshalCaddyfileBasic(t *testing.T) {
	input := `i18n {
		dict_file /path/to/dict.json
	}`

	d := caddyfile.NewTestDispenser(input)
	i18n := &I18n{}

	err := i18n.UnmarshalCaddyfile(d)
	if err != nil {
		t.Fatalf("UnmarshalCaddyfile failed: %v", err)
	}

	if i18n.DictFile != "/path/to/dict.json" {
		t.Errorf("expected DictFile '/path/to/dict.json', got %q", i18n.DictFile)
	}
}

func TestUnmarshalCaddyfileRelativePath(t *testing.T) {
	input := `i18n {
		dict_file ./translations.json
	}`

	d := caddyfile.NewTestDispenser(input)
	i18n := &I18n{}

	err := i18n.UnmarshalCaddyfile(d)
	if err != nil {
		t.Fatalf("UnmarshalCaddyfile failed: %v", err)
	}

	if i18n.DictFile != "./translations.json" {
		t.Errorf("expected DictFile './translations.json', got %q", i18n.DictFile)
	}
}

func TestUnmarshalCaddyfileEmpty(t *testing.T) {
	input := `i18n {
	}`

	d := caddyfile.NewTestDispenser(input)
	i18n := &I18n{}

	err := i18n.UnmarshalCaddyfile(d)
	if err != nil {
		t.Fatalf("UnmarshalCaddyfile should succeed with empty block: %v", err)
	}

	if i18n.DictFile != "" {
		t.Errorf("expected empty DictFile, got %q", i18n.DictFile)
	}
}

func TestUnmarshalCaddyfileMissingValue(t *testing.T) {
	input := `i18n {
		dict_file
	}`

	d := caddyfile.NewTestDispenser(input)
	i18n := &I18n{}

	err := i18n.UnmarshalCaddyfile(d)
	if err == nil {
		t.Fatal("expected error for missing dict_file value")
	}
}

func TestUnmarshalCaddyfileTooManyArgs(t *testing.T) {
	input := `i18n {
		dict_file /path/to/dict.json extra
	}`

	d := caddyfile.NewTestDispenser(input)
	i18n := &I18n{}

	err := i18n.UnmarshalCaddyfile(d)
	if err == nil {
		t.Fatal("expected error for too many arguments")
	}
}

func TestUnmarshalCaddyfileUnknownProperty(t *testing.T) {
	input := `i18n {
		unknown_prop value
	}`

	d := caddyfile.NewTestDispenser(input)
	i18n := &I18n{}

	err := i18n.UnmarshalCaddyfile(d)
	if err == nil {
		t.Fatal("expected error for unknown property")
	}
	if !strings.Contains(err.Error(), "unrecognized i18n config property: unknown_prop") {
		t.Errorf("expected error containing 'unrecognized i18n config property: unknown_prop', got: %v", err)
	}
}

func TestUnmarshalCaddyfileMultipleProperties(t *testing.T) {
	input := `i18n {
		dict_file /path/to/dict.json
		unknown_prop value
	}`

	d := caddyfile.NewTestDispenser(input)
	i18n := &I18n{}

	err := i18n.UnmarshalCaddyfile(d)
	if err == nil {
		t.Fatal("expected error for unknown property after valid one")
	}
}

func TestUnmarshalCaddyfileAbsolutePath(t *testing.T) {
	input := `i18n {
		dict_file /etc/caddy/translations.json
	}`

	d := caddyfile.NewTestDispenser(input)
	i18n := &I18n{}

	err := i18n.UnmarshalCaddyfile(d)
	if err != nil {
		t.Fatalf("UnmarshalCaddyfile failed: %v", err)
	}

	if i18n.DictFile != "/etc/caddy/translations.json" {
		t.Errorf("expected DictFile '/etc/caddy/translations.json', got %q", i18n.DictFile)
	}
}

func TestUnmarshalCaddyfilePathWithSpaces(t *testing.T) {
	input := `i18n {
		dict_file "/path/to/my translations.json"
	}`

	d := caddyfile.NewTestDispenser(input)
	i18n := &I18n{}

	err := i18n.UnmarshalCaddyfile(d)
	if err != nil {
		t.Fatalf("UnmarshalCaddyfile failed: %v", err)
	}

	if i18n.DictFile != "/path/to/my translations.json" {
		t.Errorf("expected DictFile '/path/to/my translations.json', got %q", i18n.DictFile)
	}
}

func TestUnmarshalCaddyfileInterfaceGuard(t *testing.T) {
	var i interface{} = &I18n{}
	_, ok := i.(caddyfile.Unmarshaler)
	if !ok {
		t.Fatal("I18n should implement caddyfile.Unmarshaler")
	}
}
