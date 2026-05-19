package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"inventory-movement-processing/common"
	"inventory-movement-processing/pkg/components/configc"
	"inventory-movement-processing/pkg/components/ginc"
	"inventory-movement-processing/pkg/components/ginc/middleware"
	sctx "inventory-movement-processing/pkg/service_context"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	os.Setenv("APP_ENV", sctx.PrdEnv)
	var dateFlag string
	flag.StringVar(&dateFlag, "date", "", "Target date for the report (YYYY-MM-DD). Defaults to current date if omitted.")

	serviceCtx := sctx.NewServiceContext(
		sctx.WithName("report"),
		sctx.WithComponent(ginc.NewGin("gin")),
		sctx.WithComponent(configc.NewConfigComponent("config")),
	)

	if err := serviceCtx.Load(); err != nil {
		log.Fatalln(err)
	}

	ginComp := serviceCtx.MustGet("gin").(common.HTTPServer)
	configComp := serviceCtx.MustGet("config").(middleware.Config)
	port := ginComp.GetPort()

	baseURL := fmt.Sprintf("http://localhost:%d/api/v1/reports/daily", port)
	authToken := configComp.GetManagerAPIKey()
	targetDate := dateFlag
	if targetDate == "" {
		targetDate = time.Now().Format("2006-01-02")
	}
	limit := 10

	reqURL := fmt.Sprintf("%s?date=%s&limit=%d", baseURL, targetDate, limit)
	fmt.Printf("Preparing to call API: %s\n", reqURL)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error calling API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("API returned error with HTTP status: %d %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, body, "", "  ")
	if err != nil {
		log.Fatalf("Error formatting JSON: %v", err)
	}

	fileName := fmt.Sprintf("daily_inventory_report_%s.json", targetDate)
	err = os.WriteFile(fileName, prettyJSON.Bytes(), 0644)
	if err != nil {
		log.Fatalf("Error saving report file: %v", err)
	}

	fmt.Printf("Successfully authenticated and generated report. Data saved to file: %s\n", fileName)
}
