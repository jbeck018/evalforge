package notifications

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	TypeEvaluationComplete NotificationType = "evaluation_complete"
	TypeErrorAlert         NotificationType = "error_alert"
	TypeCostAlert          NotificationType = "cost_alert"
	TypePerformanceAlert   NotificationType = "performance_alert"
	TypeABTestResult       NotificationType = "ab_test_result"
	TypeDailyReport        NotificationType = "daily_report"
)

// NotificationChannel represents a notification channel
type NotificationChannel string

const (
	ChannelSlack   NotificationChannel = "slack"
	ChannelEmail   NotificationChannel = "email"
	ChannelWebhook NotificationChannel = "webhook"
)

// NotificationConfig stores notification configuration
type NotificationConfig struct {
	ID        int                    `json:"id"`
	ProjectID int                    `json:"project_id"`
	Channel   NotificationChannel    `json:"channel"`
	Type      NotificationType       `json:"type"`
	Config    map[string]interface{} `json:"config"`
	Enabled   bool                   `json:"enabled"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// NotificationManager manages all notification channels
type NotificationManager struct {
	db           *sql.DB
	slackClients map[int]*SlackClient
	configs      map[int][]NotificationConfig
	mu           sync.RWMutex
}

// NewNotificationManager creates a new notification manager
func NewNotificationManager(db *sql.DB) *NotificationManager {
	nm := &NotificationManager{
		db:           db,
		slackClients: make(map[int]*SlackClient),
		configs:      make(map[int][]NotificationConfig),
	}
	
	// Load configs on startup
	nm.LoadConfigs()
	
	return nm
}

// LoadConfigs loads notification configs from database
func (nm *NotificationManager) LoadConfigs() error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	rows, err := nm.db.Query(`
		SELECT id, project_id, channel, type, config, enabled, created_at, updated_at
		FROM notification_configs
		WHERE enabled = true
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Clear existing configs
	nm.configs = make(map[int][]NotificationConfig)
	nm.slackClients = make(map[int]*SlackClient)

	for rows.Next() {
		var config NotificationConfig
		var configJSON string
		
		err := rows.Scan(
			&config.ID,
			&config.ProjectID,
			&config.Channel,
			&config.Type,
			&configJSON,
			&config.Enabled,
			&config.CreatedAt,
			&config.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning notification config: %v", err)
			continue
		}

		// Parse config JSON
		if err := json.Unmarshal([]byte(configJSON), &config.Config); err != nil {
			log.Printf("Error parsing config JSON: %v", err)
			continue
		}

		// Add to project configs
		nm.configs[config.ProjectID] = append(nm.configs[config.ProjectID], config)

		// Initialize Slack client if needed
		if config.Channel == ChannelSlack {
			if webhookURL, ok := config.Config["webhook_url"].(string); ok {
				nm.slackClients[config.ProjectID] = NewSlackClient(webhookURL)
			}
		}
	}

	return nil
}

// SaveConfig saves a notification configuration
func (nm *NotificationManager) SaveConfig(config *NotificationConfig) error {
	configJSON, err := json.Marshal(config.Config)
	if err != nil {
		return err
	}

	if config.ID == 0 {
		// Insert new config
		err = nm.db.QueryRow(`
			INSERT INTO notification_configs (project_id, channel, type, config, enabled)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, created_at, updated_at
		`, config.ProjectID, config.Channel, config.Type, string(configJSON), config.Enabled).
			Scan(&config.ID, &config.CreatedAt, &config.UpdatedAt)
	} else {
		// Update existing config
		_, err = nm.db.Exec(`
			UPDATE notification_configs
			SET channel = $2, type = $3, config = $4, enabled = $5, updated_at = NOW()
			WHERE id = $1
		`, config.ID, config.Channel, config.Type, string(configJSON), config.Enabled)
		config.UpdatedAt = time.Now()
	}

	if err != nil {
		return err
	}

	// Reload configs
	nm.LoadConfigs()
	
	return nil
}

// SendNotification sends a notification based on type and project settings
func (nm *NotificationManager) SendNotification(projectID int, notifType NotificationType, data map[string]interface{}) error {
	nm.mu.RLock()
	configs, exists := nm.configs[projectID]
	nm.mu.RUnlock()

	if !exists || len(configs) == 0 {
		return nil // No notifications configured
	}

	var errors []error

	for _, config := range configs {
		if config.Type != notifType && config.Type != "*" {
			continue
		}

		switch config.Channel {
		case ChannelSlack:
			if err := nm.sendSlackNotification(projectID, notifType, data); err != nil {
				errors = append(errors, err)
			}
		case ChannelEmail:
			// TODO: Implement email notifications
		case ChannelWebhook:
			if err := nm.sendWebhookNotification(config, data); err != nil {
				errors = append(errors, err)
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("notification errors: %v", errors)
	}

	return nil
}

// sendSlackNotification sends a Slack notification
func (nm *NotificationManager) sendSlackNotification(projectID int, notifType NotificationType, data map[string]interface{}) error {
	nm.mu.RLock()
	client, exists := nm.slackClients[projectID]
	nm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("slack client not configured for project %d", projectID)
	}

	switch notifType {
	case TypeEvaluationComplete:
		evalName := getStringValue(data, "eval_name", "Unknown")
		projectName := getStringValue(data, "project_name", "Unknown")
		successRate := getFloatValue(data, "success_rate", 0)
		totalSamples := getIntValue(data, "total_samples", 0)
		return client.SendEvaluationComplete(evalName, projectName, successRate, totalSamples)

	case TypeErrorAlert:
		projectName := getStringValue(data, "project_name", "Unknown")
		errorMessage := getStringValue(data, "error_message", "Unknown error")
		errorRate := getFloatValue(data, "error_rate", 0)
		return client.SendErrorAlert(projectName, errorMessage, errorRate)

	case TypeCostAlert:
		projectName := getStringValue(data, "project_name", "Unknown")
		currentCost := getFloatValue(data, "current_cost", 0)
		threshold := getFloatValue(data, "threshold", 0)
		return client.SendCostAlert(projectName, currentCost, threshold)

	case TypePerformanceAlert:
		projectName := getStringValue(data, "project_name", "Unknown")
		metric := getStringValue(data, "metric", "Unknown")
		currentValue := getFloatValue(data, "current_value", 0)
		threshold := getFloatValue(data, "threshold", 0)
		return client.SendPerformanceAlert(projectName, metric, currentValue, threshold)

	case TypeABTestResult:
		testName := getStringValue(data, "test_name", "Unknown")
		variantA := getStringValue(data, "variant_a", "A")
		variantB := getStringValue(data, "variant_b", "B")
		winner := getStringValue(data, "winner", "Unknown")
		improvement := getFloatValue(data, "improvement", 0)
		return client.SendABTestResult(testName, variantA, variantB, winner, improvement)

	case TypeDailyReport:
		projectName := getStringValue(data, "project_name", "Unknown")
		return client.SendDailyReport(projectName, data)

	default:
		return fmt.Errorf("unknown notification type: %s", notifType)
	}
}

// sendWebhookNotification sends a webhook notification
func (nm *NotificationManager) sendWebhookNotification(config NotificationConfig, data map[string]interface{}) error {
	webhookURL, ok := config.Config["url"].(string)
	if !ok {
		return fmt.Errorf("webhook URL not configured")
	}

	// Create webhook payload
	payload := map[string]interface{}{
		"type":       config.Type,
		"project_id": config.ProjectID,
		"timestamp":  time.Now().Unix(),
		"data":       data,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Send webhook
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(webhookURL, "application/json", bytes.NewBuffer(payloadJSON))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// Helper functions to safely extract values from map
func getStringValue(data map[string]interface{}, key string, defaultValue string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return defaultValue
}

func getFloatValue(data map[string]interface{}, key string, defaultValue float64) float64 {
	if val, ok := data[key].(float64); ok {
		return val
	}
	if val, ok := data[key].(int); ok {
		return float64(val)
	}
	return defaultValue
}

func getIntValue(data map[string]interface{}, key string, defaultValue int) int {
	if val, ok := data[key].(int); ok {
		return val
	}
	if val, ok := data[key].(float64); ok {
		return int(val)
	}
	return defaultValue
}