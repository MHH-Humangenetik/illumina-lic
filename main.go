package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/pterm/pterm"
)

type License struct {
	Pipeline   string  `json:"pipeline"`
	QuotaLimit float64 `json:"quota_limit"`
	Usage      float64 `json:"usage"`
	Units      string  `json:"units"`
	Expiration string  `json:"expiration"`
}

type Response struct {
	Licenses            []License `json:"licenses"`
	Message             string    `json:"message"`
	CredentialExpiresOn string    `json:"credential_expiration"`
}

func main() {
	endpoint := flag.String("endpoint", "https://license.dragen.illumina.com/api/v2/query_quota", "License query endpoint URL")
	envPath := flag.String("env", ".env", "Path to .env file")
	showInfo := flag.Bool("info", false, "Show message and credential expiration info line")
	flag.Parse()

	// ponytail: ignore missing .env; vars may be exported in the shell
	_ = godotenv.Load(*envPath)

	username := os.Getenv("LICENSE_USERNAME")
	password := os.Getenv("LICENSE_PASSWORD")
	if username == "" || password == "" {
		pterm.Error.Println("LICENSE_USERNAME and LICENSE_PASSWORD must be set in .env")
		os.Exit(1)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", *endpoint, nil)
	if err != nil {
		pterm.Error.Printf("Failed to create request: %v\n", err)
		os.Exit(1)
	}
	req.SetBasicAuth(username, password)

	spinner := pterm.DefaultSpinner.WithText("Fetching license data...")
	spinner.Start()

	resp, err := client.Do(req)
	if err != nil {
		spinner.Fail(fmt.Sprintf("Request failed: %v", err))
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		spinner.Fail(fmt.Sprintf("HTTP %d: %s %s", resp.StatusCode, resp.Status, body))
		os.Exit(1)
	}

	spinner.Success("License data retrieved")

	var data Response
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		pterm.Error.Printf("Failed to parse JSON: %v\n", err)
		os.Exit(1)
	}

	if line := infoLine(*showInfo, data); line != "" {
		pterm.Info.Println(line)
		pterm.Println()
	}

	var tableData pterm.TableData
	tableData = append(tableData, []string{"Pipeline", "Limit", "Used", "% Used", "Expires"})
	tableData = append(tableData, buildRows(data.Licenses)...)

	pterm.DefaultTable.WithHasHeader().WithHeaderStyle(pterm.NewStyle(pterm.Bold)).WithHeaderRowSeparator("─").WithBoxed().WithData(tableData).Render()
}

func infoLine(show bool, data Response) string {
	if !show {
		return ""
	}
	return fmt.Sprintf("Message: %s | Credential Expiration: %s", data.Message, data.CredentialExpiresOn)
}

func buildRows(licenses []License) pterm.TableData {
	var rows pterm.TableData
	for _, lic := range licenses {
		if lic.QuotaLimit == -1 {
			continue
		}

		percent := "n/a"
		if lic.QuotaLimit > 0 {
			pct := (lic.Usage / lic.QuotaLimit) * 100
			color := pterm.FgGreen
			if pct >= 90 {
				color = pterm.FgRed
			} else if pct >= 70 {
				color = pterm.FgYellow
			}
			percent = color.Sprint(fmt.Sprintf("%.2f%%", pct))
		}

		rows = append(rows, []string{
			lic.Pipeline,
			fmt.Sprintf("%.2f %s", lic.QuotaLimit, lic.Units),
			fmt.Sprintf("%.4f %s", lic.Usage, lic.Units),
			percent,
			lic.Expiration,
		})
	}
	return rows
}
