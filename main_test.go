package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestLicenseParsing(t *testing.T) {
	jsonData := `{
		"message": "License query successful",
		"credential_expiration": "2026-12-31T23:59:59Z",
		"licenses": [
			{
				"pipeline": "5base",
				"quota_limit": -1,
				"usage": 0.5,
				"units": "Tb",
				"expiration": "2026-06-30"
			},
			{
				"pipeline": "Compression",
				"quota_limit": 100,
				"usage": 45.2,
				"units": "Tb",
				"expiration": "2026-06-30"
			},
			{
				"pipeline": "Genome",
				"quota_limit": 10,
				"usage": 0.27861,
				"units": "Tb",
				"expiration": "2026-06-30"
			},
			{
				"pipeline": "PipSeq",
				"quota_limit": -1,
				"usage": 1.2,
				"units": "Tb",
				"expiration": "2026-06-30"
			},
			{
				"pipeline": "Spatial",
				"quota_limit": -1,
				"usage": 0.8,
				"units": "Tb",
				"expiration": "2026-06-30"
			},
			{
				"pipeline": "ZeroLimit",
				"quota_limit": 0,
				"usage": 5.0,
				"units": "Tb",
				"expiration": "2026-06-30"
			}
		]
	}`

	var data Response
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify message and credential expiration parsed correctly
	if data.Message != "License query successful" {
		t.Errorf("Expected message 'License query successful', got '%s'", data.Message)
	}
	if data.CredentialExpiresOn != "2026-12-31T23:59:59Z" {
		t.Errorf("Expected credential_expiration '2026-12-31T23:59:59Z', got '%s'", data.CredentialExpiresOn)
	}

	// Verify 6 licenses total (including 3 with quota_limit: -1 and 1 with quota_limit: 0)
	if len(data.Licenses) != 6 {
		t.Errorf("Expected 6 licenses, got %d", len(data.Licenses))
	}

	// Verify Genome license: usage 0.27861 / limit 10 = 2.79%
	rows := buildRows(data.Licenses)

	// 3 rows: Compression, Genome, ZeroLimit (5base, PipSeq, Spatial skipped)
	if len(rows) != 3 {
		t.Fatalf("Expected 3 rows, got %d", len(rows))
	}
	if rows[0][0] != "Compression" || rows[1][0] != "Genome" || rows[2][0] != "ZeroLimit" {
		t.Errorf("Unexpected row order/content: %s, %s, %s", rows[0][0], rows[1][0], rows[2][0])
	}

	genome := data.Licenses[2]
	if genome.QuotaLimit != 10 {
		t.Errorf("Expected quota_limit 10, got %f", genome.QuotaLimit)
	}
	if genome.Usage != 0.27861 {
		t.Errorf("Expected usage 0.27861, got %f", genome.Usage)
	}
	if !strings.Contains(rows[1][3], "2.79%") {
		t.Errorf("Expected Genome percent '2.79%%', got %q", rows[1][3])
	}

	// Zero quota_limit yields "n/a" with no color codes
	if rows[2][3] != "n/a" {
		t.Errorf("Expected ZeroLimit percent 'n/a', got %q", rows[2][3])
	}

}

func TestInfoLine(t *testing.T) {
	data := Response{Message: "Active Licenses Found", CredentialExpiresOn: "2050-01-01T00:00:00+00:00"}
	if got := infoLine(false, data); got != "" {
		t.Errorf("expected no info line without --info, got %q", got)
	}
	want := "Message: Active Licenses Found | Credential Expiration: 2050-01-01T00:00:00+00:00"
	if got := infoLine(true, data); got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestBuildRowsColors(t *testing.T) {
	cases := []struct {
		usage, limit float64
		wantCode     string
	}{
		{45, 100, "\x1b[32m"}, // green < 70
		{75, 100, "\x1b[33m"}, // yellow 70-89
		{95, 100, "\x1b[31m"}, // red >= 90
	}
	for _, c := range cases {
		rows := buildRows([]License{{Pipeline: "X", QuotaLimit: c.limit, Usage: c.usage, Units: "Tb"}})
		if !strings.Contains(rows[0][3], c.wantCode) {
			t.Errorf("usage %v/%v: expected color code %q in %q", c.usage, c.limit, c.wantCode, rows[0][3])
		}
	}
}
