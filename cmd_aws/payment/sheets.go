package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func getSheetsService(ctx context.Context) (*sheets.Service, error) {
	creds := loadAppConfig().GoogleCredentialsJSON
	if creds == "" {
		return nil, fmt.Errorf("GOOGLE_CREDENTIALS_JSON not set")
	}
	cfg, err := google.JWTConfigFromJSON([]byte(creds), sheets.SpreadsheetsScope)
	if err != nil {
		return nil, err
	}
	client := cfg.Client(ctx)
	return sheets.NewService(ctx, option.WithHTTPClient(client))
}

func updateGoogleSheets(row []string, sheet string, wg *sync.WaitGroup) {
	defer wg.Done()
	id := spreadsheetID()
	if id == "" {
		log.Printf("Google Sheets skipped: no spreadsheet id")
		return
	}
	srv, err := getSheetsService(context.Background())
	if err != nil {
		log.Printf("Unable to retrieve GoogleSheets client: %v", err)
		return
	}
	vr := &sheets.ValueRange{Values: [][]interface{}{toInterfaceRow(row)}}
	rangeStr := sheet + "!A:Z"
	_, err = srv.Spreadsheets.Values.Append(id, rangeStr, vr).
		ValueInputOption("USER_ENTERED").
		InsertDataOption("INSERT_ROWS").
		Do()
	if err != nil {
		log.Printf("Unable to upload data to Google Sheets: %v", err)
		return
	}
	log.Printf("Row Updated into Google Sheets sheet=%s", sheet)
}

func toInterfaceRow(row []string) []interface{} {
	out := make([]interface{}, len(row))
	for i, v := range row {
		out[i] = v
	}
	return out
}

// Ensure google JWT JSON is valid when provided (startup sanity, optional).
func validateGoogleCredsJSON() {
	raw := os.Getenv("GOOGLE_CREDENTIALS_JSON")
	if raw == "" {
		return
	}
	var probe map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &probe); err != nil {
		log.Printf("GOOGLE_CREDENTIALS_JSON is not valid JSON: %v", err)
	}
}

func sheetTitleSafe(s string) string {
	return strings.ReplaceAll(s, "'", "")
}
