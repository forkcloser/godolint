package main

import (
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	t.Parallel()

	source, count, err := render([]byte(`{
		"licenseListVersion": "9.9.9",
		"licenses": [
			{"licenseId": "MIT", "isDeprecatedLicenseId": false},
			{"licenseId": "GPL-3.0", "isDeprecatedLicenseId": true},
			{"isDeprecatedLicenseId": false},
			{"licenseId": "0BSD", "isDeprecatedLicenseId": false}
		]
	}`))
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	if count != 2 {
		t.Errorf("%d identifiers, want 2", count)
	}

	text := string(source)

	for _, want := range []string{"SPDX License List 9.9.9", "\t\"0BSD\": {},\n\t\"MIT\":  {},\n"} {
		if !strings.Contains(text, want) {
			t.Errorf("the table does not hold %q:\n%s", want, text)
		}
	}

	if strings.Contains(text, "GPL-3.0") {
		t.Errorf("the table holds a deprecated identifier:\n%s", text)
	}

	if strings.Contains(text, `"":`) {
		t.Errorf("the table holds an entry with no identifier:\n%s", text)
	}
}

func TestRenderRefusesAListWithNothingInIt(t *testing.T) {
	t.Parallel()

	for _, data := range []string{
		`{"licenseListVersion": "9.9.9", "licenses": []}`,
		`{"licenses": [{"licenseId": "MIT"}]}`,
		`not json`,
	} {
		if _, _, err := render([]byte(data)); err == nil {
			t.Errorf("render(%s) gave no error", data)
		}
	}
}
