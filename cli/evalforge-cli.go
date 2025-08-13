package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/fatih/color"
)

const (
	DEFAULT_API_URL = "http://localhost:8088"
	VERSION         = "1.0.0"
)

// Config holds CLI configuration
type Config struct {
	APIUrl   string
	APIKey   string
	Token    string
	Format   string
	Verbose  bool
}

// Command represents a CLI command
type Command struct {
	Name        string
	Description string
	Execute     func(config *Config, args []string) error
}

var commands = map[string]Command{
	"login": {
		Name:        "login",
		Description: "Authenticate with EvalForge",
		Execute:     cmdLogin,
	},
	"projects": {
		Name:        "projects",
		Description: "Manage projects",
		Execute:     cmdProjects,
	},
	"traces": {
		Name:        "traces",
		Description: "View and query traces",
		Execute:     cmdTraces,
	},
	"analytics": {
		Name:        "analytics",
		Description: "View analytics and metrics",
		Execute:     cmdAnalytics,
	},
	"evaluate": {
		Name:        "evaluate",
		Description: "Run evaluations",
		Execute:     cmdEvaluate,
	},
	"export": {
		Name:        "export",
		Description: "Export data",
		Execute:     cmdExport,
	},
	"ingest": {
		Name:        "ingest",
		Description: "Ingest events",
		Execute:     cmdIngest,
	},
	"cost": {
		Name:        "cost",
		Description: "Analyze costs and get optimization recommendations",
		Execute:     cmdCost,
	},
	"abtest": {
		Name:        "abtest",
		Description: "Manage A/B tests",
		Execute:     cmdABTest,
	},
	"compare": {
		Name:        "compare",
		Description: "Compare model performance",
		Execute:     cmdCompare,
	},
	"status": {
		Name:        "status",
		Description: "Check system status",
		Execute:     cmdStatus,
	},
}

// Color scheme
var (
	successColor = color.New(color.FgGreen, color.Bold)
	errorColor   = color.New(color.FgRed, color.Bold)
	infoColor    = color.New(color.FgCyan)
	warningColor = color.New(color.FgYellow)
	headerColor  = color.New(color.FgBlue, color.Bold)
)

func main() {
	config := &Config{}

	// Global flags
	flag.StringVar(&config.APIUrl, "api", getEnv("EVALFORGE_API_URL", DEFAULT_API_URL), "EvalForge API URL")
	flag.StringVar(&config.APIKey, "key", getEnv("EVALFORGE_API_KEY", ""), "API Key for authentication")
	flag.StringVar(&config.Token, "token", getEnv("EVALFORGE_TOKEN", ""), "JWT token for authentication")
	flag.StringVar(&config.Format, "format", "table", "Output format (table, json, csv)")
	flag.BoolVar(&config.Verbose, "verbose", false, "Verbose output")

	// Custom usage
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `
EvalForge CLI v%s
A command-line interface for the EvalForge observability platform

Usage:
  evalforge [flags] <command> [arguments]

Commands:
`, VERSION)
		for _, cmd := range commands {
			fmt.Fprintf(os.Stderr, "  %-12s %s\n", cmd.Name, cmd.Description)
		}
		fmt.Fprintf(os.Stderr, "\nFlags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  evalforge login -email user@example.com -password mypassword\n")
		fmt.Fprintf(os.Stderr, "  evalforge projects list\n")
		fmt.Fprintf(os.Stderr, "  evalforge traces -project 1 -limit 10\n")
		fmt.Fprintf(os.Stderr, "  evalforge analytics summary -project 1\n")
		fmt.Fprintf(os.Stderr, "  evalforge cost analyze -project 1 -days 30\n")
	}

	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	command := flag.Arg(0)
	args := flag.Args()[1:]

	// Load token from file if exists
	if config.Token == "" {
		config.Token = loadToken()
	}

	// Execute command
	if cmd, exists := commands[command]; exists {
		if err := cmd.Execute(config, args); err != nil {
			errorColor.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		errorColor.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		flag.Usage()
		os.Exit(1)
	}
}

// Helper functions
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func loadToken() string {
	homeDir, _ := os.UserHomeDir()
	tokenFile := homeDir + "/.evalforge/token"
	data, _ := os.ReadFile(tokenFile)
	return strings.TrimSpace(string(data))
}

func saveToken(token string) error {
	homeDir, _ := os.UserHomeDir()
	configDir := homeDir + "/.evalforge"
	os.MkdirAll(configDir, 0700)
	tokenFile := configDir + "/token"
	return os.WriteFile(tokenFile, []byte(token), 0600)
}

// HTTP helper functions
func httpRequest(config *Config, method, path string, body interface{}) ([]byte, error) {
	url := config.APIUrl + path

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if config.Token != "" {
		req.Header.Set("Authorization", "Bearer "+config.Token)
	}
	if config.APIKey != "" {
		req.Header.Set("X-API-Key", config.APIKey)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		var errResp map[string]interface{}
		json.Unmarshal(respBody, &errResp)
		if msg, ok := errResp["error"].(string); ok {
			return nil, fmt.Errorf("%s (HTTP %d)", msg, resp.StatusCode)
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// Command implementations
func cmdLogin(config *Config, args []string) error {
	fs := flag.NewFlagSet("login", flag.ExitOnError)
	email := fs.String("email", "", "Email address")
	password := fs.String("password", "", "Password")
	fs.Parse(args)

	if *email == "" || *password == "" {
		return fmt.Errorf("email and password are required")
	}

	body := map[string]string{
		"email":    *email,
		"password": *password,
	}

	resp, err := httpRequest(config, "POST", "/api/auth/login", body)
	if err != nil {
		return err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	token, ok := result["token"].(string)
	if !ok {
		return fmt.Errorf("invalid response: no token")
	}

	// Save token
	if err := saveToken(token); err != nil {
		return fmt.Errorf("failed to save token: %v", err)
	}

	successColor.Println("✓ Successfully logged in")
	return nil
}

func cmdProjects(config *Config, args []string) error {
	if len(args) == 0 {
		args = []string{"list"}
	}

	switch args[0] {
	case "list":
		return listProjects(config)
	case "create":
		return createProject(config, args[1:])
	case "get":
		return getProject(config, args[1:])
	default:
		return fmt.Errorf("unknown subcommand: %s", args[0])
	}
}

func listProjects(config *Config) error {
	resp, err := httpRequest(config, "GET", "/api/projects", nil)
	if err != nil {
		return err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	projects, ok := result["projects"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid response format")
	}

	if config.Format == "json" {
		fmt.Println(string(resp))
		return nil
	}

	// Table format
	headerColor.Println("\nProjects:")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tDESCRIPTION\tCREATED")
	fmt.Fprintln(w, "---\t----\t-----------\t-------")

	for _, p := range projects {
		project := p.(map[string]interface{})
		fmt.Fprintf(w, "%v\t%v\t%v\t%v\n",
			project["id"],
			project["name"],
			project["description"],
			formatTime(project["created_at"]),
		)
	}
	w.Flush()

	infoColor.Printf("\nTotal: %d projects\n", len(projects))
	return nil
}

func createProject(config *Config, args []string) error {
	fs := flag.NewFlagSet("create", flag.ExitOnError)
	name := fs.String("name", "", "Project name")
	description := fs.String("description", "", "Project description")
	fs.Parse(args)

	if *name == "" {
		return fmt.Errorf("project name is required")
	}

	body := map[string]string{
		"name":        *name,
		"description": *description,
	}

	resp, err := httpRequest(config, "POST", "/api/projects", body)
	if err != nil {
		return err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	project := result["project"].(map[string]interface{})
	successColor.Printf("✓ Created project '%s' (ID: %v)\n", project["name"], project["id"])
	
	if apiKey, ok := project["api_key"].(string); ok && apiKey != "" {
		infoColor.Printf("API Key: %s\n", apiKey)
	}

	return nil
}

func getProject(config *Config, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("project ID is required")
	}

	projectID := args[0]
	resp, err := httpRequest(config, "GET", "/api/projects/"+projectID, nil)
	if err != nil {
		return err
	}

	if config.Format == "json" {
		fmt.Println(string(resp))
		return nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	project := result["project"].(map[string]interface{})
	
	headerColor.Printf("\nProject: %s\n", project["name"])
	fmt.Printf("ID:          %v\n", project["id"])
	fmt.Printf("Description: %v\n", project["description"])
	fmt.Printf("API Key:     %v\n", project["api_key"])
	fmt.Printf("Created:     %v\n", formatTime(project["created_at"]))

	return nil
}

func cmdTraces(config *Config, args []string) error {
	fs := flag.NewFlagSet("traces", flag.ExitOnError)
	projectID := fs.String("project", "", "Project ID")
	limit := fs.Int("limit", 10, "Number of traces to retrieve")
	fs.Parse(args)

	if *projectID == "" {
		return fmt.Errorf("project ID is required")
	}

	path := fmt.Sprintf("/api/projects/%s/traces?limit=%d", *projectID, *limit)
	resp, err := httpRequest(config, "GET", path, nil)
	if err != nil {
		return err
	}

	if config.Format == "json" {
		fmt.Println(string(resp))
		return nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	traces, ok := result["events"].([]interface{})
	if !ok || len(traces) == 0 {
		infoColor.Println("No traces found")
		return nil
	}

	headerColor.Println("\nTraces:")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TRACE ID\tOPERATION\tSTATUS\tDURATION(ms)\tCOST\tMODEL")
	fmt.Fprintln(w, "--------\t---------\t------\t------------\t----\t-----")

	for _, t := range traces {
		trace := t.(map[string]interface{})
		fmt.Fprintf(w, "%v\t%v\t%v\t%v\t$%.4f\t%v\n",
			truncate(getString(trace, "trace_id"), 12),
			getString(trace, "operation_type"),
			getString(trace, "status"),
			trace["duration_ms"],
			getFloat(trace, "cost"),
			getString(trace, "model"),
		)
	}
	w.Flush()

	infoColor.Printf("\nTotal: %d traces\n", len(traces))
	return nil
}

func cmdAnalytics(config *Config, args []string) error {
	if len(args) == 0 {
		args = []string{"summary"}
	}

	fs := flag.NewFlagSet("analytics", flag.ExitOnError)
	projectID := fs.String("project", "", "Project ID")
	fs.Parse(args[1:])

	if *projectID == "" {
		return fmt.Errorf("project ID is required")
	}

	switch args[0] {
	case "summary":
		return getAnalyticsSummary(config, *projectID)
	case "costs":
		return getAnalyticsCosts(config, *projectID)
	case "latency":
		return getAnalyticsLatency(config, *projectID)
	case "errors":
		return getAnalyticsErrors(config, *projectID)
	default:
		return fmt.Errorf("unknown analytics type: %s", args[0])
	}
}

func getAnalyticsSummary(config *Config, projectID string) error {
	path := fmt.Sprintf("/api/projects/%s/analytics/summary", projectID)
	resp, err := httpRequest(config, "GET", path, nil)
	if err != nil {
		return err
	}

	if config.Format == "json" {
		fmt.Println(string(resp))
		return nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	summary := result["summary"].(map[string]interface{})
	
	headerColor.Println("\nAnalytics Summary:")
	fmt.Printf("Total Events:    %v\n", summary["total_events"])
	fmt.Printf("Success Rate:    %.2f%%\n", getFloat(summary, "success_rate")*100)
	fmt.Printf("Avg Latency:     %.2fms\n", getFloat(summary, "average_latency"))
	fmt.Printf("Total Cost:      $%.4f\n", getFloat(summary, "total_cost"))
	fmt.Printf("Active Models:   %v\n", summary["active_models"])
	fmt.Printf("Error Count:     %v\n", summary["error_count"])

	return nil
}

func getAnalyticsCosts(config *Config, projectID string) error {
	path := fmt.Sprintf("/api/projects/%s/analytics/costs", projectID)
	resp, err := httpRequest(config, "GET", path, nil)
	if err != nil {
		return err
	}

	if config.Format == "json" {
		fmt.Println(string(resp))
		return nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	costs := result["costs"].(map[string]interface{})
	
	headerColor.Println("\nCost Analytics:")
	fmt.Printf("Total Cost:      $%.4f\n", getFloat(costs, "total"))
	fmt.Printf("Average Cost:    $%.6f\n", getFloat(costs, "average"))
	fmt.Printf("Daily Average:   $%.4f\n", getFloat(costs, "daily_average"))
	
	if byModel, ok := costs["by_model"].([]interface{}); ok && len(byModel) > 0 {
		fmt.Println("\nCost by Model:")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "MODEL\tCOST\tEVENTS")
		fmt.Fprintln(w, "-----\t----\t------")
		for _, m := range byModel {
			model := m.(map[string]interface{})
			fmt.Fprintf(w, "%v\t$%.4f\t%v\n",
				model["model"],
				getFloat(model, "cost"),
				model["events"],
			)
		}
		w.Flush()
	}

	return nil
}

func getAnalyticsLatency(config *Config, projectID string) error {
	path := fmt.Sprintf("/api/projects/%s/analytics/latency", projectID)
	resp, err := httpRequest(config, "GET", path, nil)
	if err != nil {
		return err
	}

	if config.Format == "json" {
		fmt.Println(string(resp))
		return nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	latency := result["latency"].(map[string]interface{})
	
	headerColor.Println("\nLatency Analytics:")
	fmt.Printf("Average:  %.2fms\n", getFloat(latency, "average"))
	fmt.Printf("Median:   %.2fms\n", getFloat(latency, "median"))
	fmt.Printf("P95:      %.2fms\n", getFloat(latency, "p95"))
	fmt.Printf("P99:      %.2fms\n", getFloat(latency, "p99"))
	fmt.Printf("Min:      %.2fms\n", getFloat(latency, "min"))
	fmt.Printf("Max:      %.2fms\n", getFloat(latency, "max"))

	return nil
}

func getAnalyticsErrors(config *Config, projectID string) error {
	path := fmt.Sprintf("/api/projects/%s/analytics/errors", projectID)
	resp, err := httpRequest(config, "GET", path, nil)
	if err != nil {
		return err
	}

	if config.Format == "json" {
		fmt.Println(string(resp))
		return nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	errors := result["errors"].(map[string]interface{})
	
	headerColor.Println("\nError Analytics:")
	fmt.Printf("Total Errors:    %v\n", errors["total"])
	fmt.Printf("Error Rate:      %.2f%%\n", getFloat(errors, "rate")*100)
	
	if recent, ok := errors["recent"].([]interface{}); ok && len(recent) > 0 {
		fmt.Println("\nRecent Errors:")
		for i, e := range recent {
			if i >= 5 {
				break
			}
			error := e.(map[string]interface{})
			fmt.Printf("  - [%s] %v\n", 
				formatTime(error["timestamp"]),
				error["message"],
			)
		}
	}

	return nil
}

func cmdEvaluate(config *Config, args []string) error {
	fs := flag.NewFlagSet("evaluate", flag.ExitOnError)
	evaluationID := fs.String("id", "", "Evaluation ID")
	async := fs.Bool("async", false, "Run asynchronously")
	fs.Parse(args)

	if *evaluationID == "" {
		return fmt.Errorf("evaluation ID is required")
	}

	body := map[string]bool{
		"async": *async,
	}

	path := fmt.Sprintf("/api/evaluations/%s/run", *evaluationID)
	resp, err := httpRequest(config, "POST", path, body)
	if err != nil {
		return err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	if *async {
		successColor.Printf("✓ Evaluation started (ID: %s)\n", *evaluationID)
	} else {
		successColor.Printf("✓ Evaluation completed\n")
		if results, ok := result["results"].(map[string]interface{}); ok {
			fmt.Printf("Score: %.2f%%\n", getFloat(results, "score")*100)
		}
	}

	return nil
}

func cmdExport(config *Config, args []string) error {
	fs := flag.NewFlagSet("export", flag.ExitOnError)
	projectID := fs.String("project", "", "Project ID")
	format := fs.String("format", "csv", "Export format (csv, json)")
	dataType := fs.String("type", "analytics", "Data type (analytics, evaluations, traces)")
	output := fs.String("output", "", "Output file (default: stdout)")
	fs.Parse(args)

	if *projectID == "" {
		return fmt.Errorf("project ID is required")
	}

	body := map[string]interface{}{
		"format":     *format,
		"data_type":  *dataType,
		"project_id": *projectID,
	}

	resp, err := httpRequest(config, "POST", "/api/export", body)
	if err != nil {
		return err
	}

	if *output != "" {
		if err := os.WriteFile(*output, resp, 0644); err != nil {
			return fmt.Errorf("failed to write file: %v", err)
		}
		successColor.Printf("✓ Exported to %s\n", *output)
	} else {
		fmt.Print(string(resp))
	}

	return nil
}

func cmdIngest(config *Config, args []string) error {
	fs := flag.NewFlagSet("ingest", flag.ExitOnError)
	projectID := fs.String("project", "", "Project ID")
	file := fs.String("file", "", "JSON file containing events")
	fs.Parse(args)

	if *projectID == "" || *file == "" {
		return fmt.Errorf("project ID and file are required")
	}

	// Read file
	data, err := os.ReadFile(*file)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	var events interface{}
	if err := json.Unmarshal(data, &events); err != nil {
		return fmt.Errorf("invalid JSON: %v", err)
	}

	// Wrap in batch format if needed
	var body map[string]interface{}
	if _, ok := events.([]interface{}); ok {
		body = map[string]interface{}{
			"events": events,
		}
	} else {
		body = events.(map[string]interface{})
	}

	path := fmt.Sprintf("/sdk/v1/projects/%s/events/batch", *projectID)
	resp, err := httpRequest(config, "POST", path, body)
	if err != nil {
		return err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	successColor.Printf("✓ Ingested %v events\n", result["events_count"])
	return nil
}

func cmdCost(config *Config, args []string) error {
	if len(args) == 0 {
		args = []string{"analyze"}
	}

	fs := flag.NewFlagSet("cost", flag.ExitOnError)
	projectID := fs.String("project", "", "Project ID")
	days := fs.Int("days", 30, "Number of days to analyze")
	fs.Parse(args[1:])

	if *projectID == "" {
		return fmt.Errorf("project ID is required")
	}

	switch args[0] {
	case "analyze":
		return analyzeCosts(config, *projectID, *days)
	case "recommendations":
		return getCostRecommendations(config, *projectID, *days)
	default:
		return fmt.Errorf("unknown subcommand: %s", args[0])
	}
}

func analyzeCosts(config *Config, projectID string, days int) error {
	path := fmt.Sprintf("/api/projects/%s/cost-optimization?days=%d", projectID, days)
	resp, err := httpRequest(config, "GET", path, nil)
	if err != nil {
		return err
	}

	if config.Format == "json" {
		fmt.Println(string(resp))
		return nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	report := result["optimization_report"].(map[string]interface{})
	
	headerColor.Printf("\nCost Analysis Report (%s):\n", report["analysis_period"])
	fmt.Printf("Total Cost:           $%.4f\n", getFloat(report, "total_cost"))
	fmt.Printf("Potential Savings:    $%.4f (%.1f%%)\n", 
		getFloat(report, "potential_savings"),
		getFloat(report, "savings_percentage"))
	fmt.Printf("Optimization Score:   %.0f/100\n", getFloat(report, "optimization_score"))

	// Model breakdown
	if models, ok := report["model_breakdown"].([]interface{}); ok && len(models) > 0 {
		fmt.Println("\nTop Cost Models:")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "MODEL\tPROVIDER\tCOST\tTOKENS/DAY\tCOST/TOKEN")
		fmt.Fprintln(w, "-----\t--------\t----\t----------\t----------")
		
		for i, m := range models {
			if i >= 5 {
				break
			}
			model := m.(map[string]interface{})
			fmt.Fprintf(w, "%v\t%v\t$%.4f\t%v\t$%.6f\n",
				model["model"],
				model["provider"],
				getFloat(model, "current_cost"),
				model["tokens_per_day"],
				getFloat(model, "cost_per_token"),
			)
		}
		w.Flush()
	}

	// Recommendations
	if recs, ok := report["recommendations"].([]interface{}); ok && len(recs) > 0 {
		fmt.Println("\nRecommendations:")
		for i, r := range recs {
			if i >= 3 {
				fmt.Printf("  ... and %d more\n", len(recs)-3)
				break
			}
			rec := r.(map[string]interface{})
			priority := rec["priority"].(string)
			var priorityIcon string
			switch priority {
			case "high":
				priorityIcon = "🔴"
			case "medium":
				priorityIcon = "🟡"
			default:
				priorityIcon = "🟢"
			}
			fmt.Printf("\n%s [%s] %s\n", priorityIcon, strings.ToUpper(priority), rec["title"])
			fmt.Printf("   %s\n", rec["description"])
			if savings, ok := rec["estimated_savings"].(string); ok && savings != "" {
				infoColor.Printf("   Estimated Savings: %s\n", savings)
			}
		}
	}

	return nil
}

func getCostRecommendations(config *Config, projectID string, days int) error {
	path := fmt.Sprintf("/api/projects/%s/cost-recommendations?days=%d", projectID, days)
	resp, err := httpRequest(config, "GET", path, nil)
	if err != nil {
		return err
	}

	if config.Format == "json" {
		fmt.Println(string(resp))
		return nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	headerColor.Println("\nCost Optimization Recommendations:")
	fmt.Printf("Optimization Score: %.0f/100\n", getFloat(result, "optimization_score"))
	fmt.Printf("Total Cost:         $%.4f\n", getFloat(result, "total_cost"))
	fmt.Printf("Potential Savings:  $%.4f\n", getFloat(result, "potential_savings"))

	if recs, ok := result["recommendations"].([]interface{}); ok {
		fmt.Printf("\n%d Recommendations:\n", len(recs))
		for _, r := range recs {
			rec := r.(map[string]interface{})
			fmt.Printf("\n• %s\n", rec["title"])
			fmt.Printf("  %s\n", rec["description"])
		}
	}

	if summary, ok := result["summary"].(map[string]interface{}); ok {
		if quickWins, ok := summary["quick_wins"].([]interface{}); ok && len(quickWins) > 0 {
			warningColor.Println("\nQuick Wins:")
			for _, win := range quickWins {
				fmt.Printf("  ⚡ %v\n", win)
			}
		}
	}

	return nil
}

func cmdABTest(config *Config, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("subcommand required: list, create, start, stop, results")
	}

	switch args[0] {
	case "list":
		return listABTests(config, args[1:])
	case "create":
		return createABTest(config, args[1:])
	case "start":
		return startABTest(config, args[1:])
	case "stop":
		return stopABTest(config, args[1:])
	case "results":
		return getABTestResults(config, args[1:])
	default:
		return fmt.Errorf("unknown subcommand: %s", args[0])
	}
}

func listABTests(config *Config, args []string) error {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	projectID := fs.String("project", "", "Project ID")
	fs.Parse(args)

	if *projectID == "" {
		return fmt.Errorf("project ID is required")
	}

	path := fmt.Sprintf("/api/projects/%s/abtests", *projectID)
	resp, err := httpRequest(config, "GET", path, nil)
	if err != nil {
		return err
	}

	if config.Format == "json" {
		fmt.Println(string(resp))
		return nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	tests, ok := result["tests"].([]interface{})
	if !ok || len(tests) == 0 {
		infoColor.Println("No A/B tests found")
		return nil
	}

	headerColor.Println("\nA/B Tests:")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tNAME\tSTATUS\tCONTROL\tVARIANT\tWINNER")
	fmt.Fprintln(w, "--\t----\t------\t-------\t-------\t------")

	for _, t := range tests {
		test := t.(map[string]interface{})
		fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\t%v\n",
			test["id"],
			test["name"],
			test["status"],
			test["control_samples"],
			test["variant_samples"],
			getString(test, "winner"),
		)
	}
	w.Flush()

	return nil
}

func createABTest(config *Config, args []string) error {
	fs := flag.NewFlagSet("create", flag.ExitOnError)
	projectID := fs.String("project", "", "Project ID")
	name := fs.String("name", "", "Test name")
	controlPrompt := fs.String("control", "", "Control prompt")
	variantPrompt := fs.String("variant", "", "Variant prompt")
	ratio := fs.Float64("ratio", 0.5, "Traffic ratio for variant (0-1)")
	fs.Parse(args)

	if *projectID == "" || *name == "" || *controlPrompt == "" || *variantPrompt == "" {
		return fmt.Errorf("project, name, control, and variant are required")
	}

	body := map[string]interface{}{
		"name":             *name,
		"control_prompt":   *controlPrompt,
		"variant_prompt":   *variantPrompt,
		"traffic_ratio":    *ratio,
		"min_sample_size":  100,
		"confidence_level": 0.95,
	}

	path := fmt.Sprintf("/api/projects/%s/abtests", *projectID)
	resp, err := httpRequest(config, "POST", path, body)
	if err != nil {
		return err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	test := result["test"].(map[string]interface{})
	successColor.Printf("✓ Created A/B test '%s' (ID: %v)\n", test["name"], test["id"])

	return nil
}

func startABTest(config *Config, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("test ID is required")
	}

	testID := args[0]
	path := fmt.Sprintf("/api/abtests/%s/start", testID)
	
	_, err := httpRequest(config, "POST", path, nil)
	if err != nil {
		return err
	}

	successColor.Printf("✓ Started A/B test %s\n", testID)
	return nil
}

func stopABTest(config *Config, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("test ID is required")
	}

	testID := args[0]
	path := fmt.Sprintf("/api/abtests/%s/stop", testID)
	
	_, err := httpRequest(config, "POST", path, nil)
	if err != nil {
		return err
	}

	successColor.Printf("✓ Stopped A/B test %s\n", testID)
	return nil
}

func getABTestResults(config *Config, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("test ID is required")
	}

	testID := args[0]
	path := fmt.Sprintf("/api/abtests/%s/results", testID)
	
	resp, err := httpRequest(config, "GET", path, nil)
	if err != nil {
		return err
	}

	if config.Format == "json" {
		fmt.Println(string(resp))
		return nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	analysis := result["analysis"].(map[string]interface{})
	
	headerColor.Printf("\nA/B Test Results: %s\n", analysis["test_name"])
	fmt.Printf("Control Samples:  %v\n", analysis["control_samples"])
	fmt.Printf("Variant Samples:  %v\n", analysis["variant_samples"])
	fmt.Printf("Min Sample Size:  %v\n", analysis["min_sample_size"])
	
	if statSig, ok := analysis["stat_significant"].(bool); ok && statSig {
		successColor.Printf("✓ Statistically Significant\n")
		fmt.Printf("Winner:           %v\n", analysis["winner"])
		fmt.Printf("Improvement:      %.2f%%\n", getFloat(analysis, "improvement_rate"))
	} else {
		warningColor.Println("⚠ Not statistically significant yet")
	}
	
	if rec, ok := analysis["recommendation"].(string); ok {
		fmt.Printf("\nRecommendation: %s\n", rec)
	}

	return nil
}

func cmdCompare(config *Config, args []string) error {
	fs := flag.NewFlagSet("compare", flag.ExitOnError)
	projectID := fs.String("project", "", "Project ID")
	timeRange := fs.String("range", "24h", "Time range (1h, 24h, 7d, 30d)")
	fs.Parse(args)

	if *projectID == "" {
		return fmt.Errorf("project ID is required")
	}

	path := fmt.Sprintf("/api/projects/%s/model-comparison?time_range=%s", *projectID, *timeRange)
	resp, err := httpRequest(config, "GET", path, nil)
	if err != nil {
		return err
	}

	if config.Format == "json" {
		fmt.Println(string(resp))
		return nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return err
	}

	comparison := result["comparison"].(map[string]interface{})
	
	headerColor.Printf("\nModel Comparison (%s):\n", comparison["time_range"])
	
	models, ok := comparison["models"].([]interface{})
	if !ok || len(models) == 0 {
		infoColor.Println("No model data available for comparison")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "MODEL\tPROVIDER\tEVENTS\tSUCCESS\tLATENCY\tCOST/TOKEN")
	fmt.Fprintln(w, "-----\t--------\t------\t-------\t-------\t----------")

	for _, m := range models {
		model := m.(map[string]interface{})
		fmt.Fprintf(w, "%v\t%v\t%v\t%.1f%%\t%.0fms\t$%.6f\n",
			model["model"],
			model["provider"],
			model["total_events"],
			getFloat(model, "success_rate")*100,
			getFloat(model, "avg_latency_ms"),
			getFloat(model, "cost_per_token"),
		)
	}
	w.Flush()

	// Summary
	if summary, ok := comparison["summary"].(map[string]interface{}); ok {
		fmt.Println("\nSummary:")
		if fastest, ok := summary["fastest_model"].(string); ok && fastest != "" {
			fmt.Printf("  Fastest:       %s\n", fastest)
		}
		if cheapest, ok := summary["cheapest_model"].(string); ok && cheapest != "" {
			fmt.Printf("  Cheapest:      %s\n", cheapest)
		}
		if reliable, ok := summary["most_reliable_model"].(string); ok && reliable != "" {
			fmt.Printf("  Most Reliable: %s\n", reliable)
		}
		if bestValue, ok := summary["best_value_model"].(string); ok && bestValue != "" {
			fmt.Printf("  Best Value:    %s\n", bestValue)
		}
	}

	// Recommendations
	if recs, ok := comparison["recommendations"].([]interface{}); ok && len(recs) > 0 {
		fmt.Println("\nRecommendations:")
		for _, rec := range recs {
			fmt.Printf("  • %v\n", rec)
		}
	}

	return nil
}

func cmdStatus(config *Config, args []string) error {
	// Check API health
	resp, err := httpRequest(config, "GET", "/health", nil)
	if err != nil {
		errorColor.Println("✗ API is not reachable")
		return err
	}

	var health map[string]interface{}
	json.Unmarshal(resp, &health)
	
	headerColor.Println("\nSystem Status:")
	successColor.Printf("✓ API: %s\n", health["status"])
	
	// Check authentication
	if config.Token != "" {
		resp, err = httpRequest(config, "GET", "/api/projects", nil)
		if err == nil {
			successColor.Println("✓ Authentication: Valid")
		} else {
			warningColor.Println("⚠ Authentication: Invalid or expired")
		}
	} else {
		infoColor.Println("ℹ Authentication: Not logged in")
	}

	// Display version
	fmt.Printf("\nCLI Version: %s\n", VERSION)
	fmt.Printf("API URL:     %s\n", config.APIUrl)

	return nil
}

// Helper functions
func formatTime(v interface{}) string {
	if v == nil {
		return ""
	}
	timeStr, ok := v.(string)
	if !ok {
		return ""
	}
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return timeStr
	}
	return t.Format("2006-01-02 15:04")
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getFloat(m map[string]interface{}, key string) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	return 0
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}