package notifications

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// NotificationHandler handles notification-related API endpoints
type NotificationHandler struct {
	manager *NotificationManager
	db      *sql.DB
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(db *sql.DB) *NotificationHandler {
	return &NotificationHandler{
		manager: NewNotificationManager(db),
		db:      db,
	}
}

// GetNotificationConfigs returns notification configs for a project
func (h *NotificationHandler) GetNotificationConfigs(c *gin.Context) {
	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Verify user has access to project
	userID := c.GetInt("user_id")
	var hasAccess bool
	err = h.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM projects 
			WHERE id = $1 AND user_id = $2
		)
	`, projectID, userID).Scan(&hasAccess)

	if err != nil || !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Get notification configs
	rows, err := h.db.Query(`
		SELECT id, channel, type, config, enabled, created_at, updated_at
		FROM notification_configs
		WHERE project_id = $1
		ORDER BY created_at DESC
	`, projectID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch configs"})
		return
	}
	defer rows.Close()

	configs := []gin.H{}
	for rows.Next() {
		var id int
		var channel, notifType, configJSON string
		var enabled bool
		var createdAt, updatedAt string

		err := rows.Scan(&id, &channel, &notifType, &configJSON, &enabled, &createdAt, &updatedAt)
		if err != nil {
			continue
		}

		configs = append(configs, gin.H{
			"id":         id,
			"channel":    channel,
			"type":       notifType,
			"config":     configJSON,
			"enabled":    enabled,
			"created_at": createdAt,
			"updated_at": updatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"configs": configs})
}

// CreateNotificationConfig creates a new notification config
func (h *NotificationHandler) CreateNotificationConfig(c *gin.Context) {
	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Verify user has access to project
	userID := c.GetInt("user_id")
	var hasAccess bool
	err = h.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM projects 
			WHERE id = $1 AND user_id = $2
		)
	`, projectID, userID).Scan(&hasAccess)

	if err != nil || !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var req struct {
		Channel NotificationChannel    `json:"channel" binding:"required,oneof=slack email webhook"`
		Type    NotificationType       `json:"type" binding:"required"`
		Config  map[string]interface{} `json:"config" binding:"required"`
		Enabled bool                   `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config := &NotificationConfig{
		ProjectID: projectID,
		Channel:   req.Channel,
		Type:      req.Type,
		Config:    req.Config,
		Enabled:   req.Enabled,
	}

	if err := h.manager.SaveConfig(config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save config"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      config.ID,
		"message": "Notification config created successfully",
	})
}

// UpdateNotificationConfig updates a notification config
func (h *NotificationHandler) UpdateNotificationConfig(c *gin.Context) {
	configID, err := strconv.Atoi(c.Param("configId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid config ID"})
		return
	}

	// Verify user has access
	userID := c.GetInt("user_id")
	var hasAccess bool
	err = h.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM notification_configs nc
			JOIN projects p ON nc.project_id = p.id
			WHERE nc.id = $1 AND p.user_id = $2
		)
	`, configID, userID).Scan(&hasAccess)

	if err != nil || !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var req struct {
		Channel NotificationChannel    `json:"channel"`
		Type    NotificationType       `json:"type"`
		Config  map[string]interface{} `json:"config"`
		Enabled *bool                  `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update config
	updates := []string{}
	args := []interface{}{}
	argCount := 1

	if req.Channel != "" {
		updates = append(updates, "channel = $"+strconv.Itoa(argCount))
		args = append(args, req.Channel)
		argCount++
	}

	if req.Type != "" {
		updates = append(updates, "type = $"+strconv.Itoa(argCount))
		args = append(args, req.Type)
		argCount++
	}

	if req.Config != nil {
		configJSON, _ := json.Marshal(req.Config)
		updates = append(updates, "config = $"+strconv.Itoa(argCount))
		args = append(args, string(configJSON))
		argCount++
	}

	if req.Enabled != nil {
		updates = append(updates, "enabled = $"+strconv.Itoa(argCount))
		args = append(args, *req.Enabled)
		argCount++
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No updates provided"})
		return
	}

	args = append(args, configID)
	query := "UPDATE notification_configs SET " + strings.Join(updates, ", ") + 
		", updated_at = NOW() WHERE id = $" + strconv.Itoa(argCount)

	_, err = h.db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update config"})
		return
	}

	// Reload configs
	h.manager.LoadConfigs()

	c.JSON(http.StatusOK, gin.H{"message": "Config updated successfully"})
}

// DeleteNotificationConfig deletes a notification config
func (h *NotificationHandler) DeleteNotificationConfig(c *gin.Context) {
	configID, err := strconv.Atoi(c.Param("configId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid config ID"})
		return
	}

	// Verify user has access
	userID := c.GetInt("user_id")
	var hasAccess bool
	err = h.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM notification_configs nc
			JOIN projects p ON nc.project_id = p.id
			WHERE nc.id = $1 AND p.user_id = $2
		)
	`, configID, userID).Scan(&hasAccess)

	if err != nil || !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	_, err = h.db.Exec("DELETE FROM notification_configs WHERE id = $1", configID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete config"})
		return
	}

	// Reload configs
	h.manager.LoadConfigs()

	c.JSON(http.StatusOK, gin.H{"message": "Config deleted successfully"})
}

// TestNotification sends a test notification
func (h *NotificationHandler) TestNotification(c *gin.Context) {
	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Verify user has access
	userID := c.GetInt("user_id")
	var projectName string
	err = h.db.QueryRow(`
		SELECT name FROM projects 
		WHERE id = $1 AND user_id = $2
	`, projectID, userID).Scan(&projectName)

	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var req struct {
		Type NotificationType `json:"type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Send test notification
	testData := map[string]interface{}{
		"project_name":  projectName,
		"eval_name":     "Test Evaluation",
		"success_rate":  0.95,
		"total_samples": 100,
		"error_message": "This is a test error message",
		"error_rate":    0.05,
		"current_cost":  50.00,
		"threshold":     100.00,
		"metric":        "latency_p95",
		"current_value": 500.0,
		"test_name":     "Test A/B Experiment",
		"variant_a":     "Control",
		"variant_b":     "Treatment",
		"winner":        "Treatment",
		"improvement":   15.5,
		"total_events":  1000,
		"avg_latency":   250.5,
		"total_cost":    25.50,
	}

	err = h.manager.SendNotification(projectID, req.Type, testData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send notification: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Test notification sent successfully"})
}

// GetAlertThresholds returns alert thresholds for a project
func (h *NotificationHandler) GetAlertThresholds(c *gin.Context) {
	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Verify user has access
	userID := c.GetInt("user_id")
	var hasAccess bool
	err = h.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM projects 
			WHERE id = $1 AND user_id = $2
		)
	`, projectID, userID).Scan(&hasAccess)

	if err != nil || !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	rows, err := h.db.Query(`
		SELECT id, metric_type, threshold_value, comparison_operator, 
		       time_window_minutes, cooldown_minutes, enabled, last_triggered
		FROM alert_thresholds
		WHERE project_id = $1
		ORDER BY metric_type
	`, projectID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch thresholds"})
		return
	}
	defer rows.Close()

	thresholds := []gin.H{}
	for rows.Next() {
		var id int
		var metricType, operator string
		var thresholdValue float64
		var timeWindow, cooldown int
		var enabled bool
		var lastTriggered sql.NullTime

		err := rows.Scan(&id, &metricType, &thresholdValue, &operator,
			&timeWindow, &cooldown, &enabled, &lastTriggered)
		if err != nil {
			continue
		}

		threshold := gin.H{
			"id":                  id,
			"metric_type":         metricType,
			"threshold_value":     thresholdValue,
			"comparison_operator": operator,
			"time_window_minutes": timeWindow,
			"cooldown_minutes":    cooldown,
			"enabled":             enabled,
		}

		if lastTriggered.Valid {
			threshold["last_triggered"] = lastTriggered.Time
		}

		thresholds = append(thresholds, threshold)
	}

	c.JSON(http.StatusOK, gin.H{"thresholds": thresholds})
}

// SetAlertThreshold sets an alert threshold
func (h *NotificationHandler) SetAlertThreshold(c *gin.Context) {
	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Verify user has access
	userID := c.GetInt("user_id")
	var hasAccess bool
	err = h.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM projects 
			WHERE id = $1 AND user_id = $2
		)
	`, projectID, userID).Scan(&hasAccess)

	if err != nil || !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var req struct {
		MetricType    string  `json:"metric_type" binding:"required"`
		Threshold     float64 `json:"threshold_value" binding:"required"`
		Operator      string  `json:"comparison_operator" binding:"required,oneof=> < >= <= ="`
		TimeWindow    int     `json:"time_window_minutes"`
		Cooldown      int     `json:"cooldown_minutes"`
		Enabled       bool    `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set defaults
	if req.TimeWindow == 0 {
		req.TimeWindow = 5
	}
	if req.Cooldown == 0 {
		req.Cooldown = 30
	}

	// Upsert threshold
	var thresholdID int
	err = h.db.QueryRow(`
		INSERT INTO alert_thresholds 
		(project_id, metric_type, threshold_value, comparison_operator, 
		 time_window_minutes, cooldown_minutes, enabled)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (project_id, metric_type) 
		DO UPDATE SET 
			threshold_value = $3,
			comparison_operator = $4,
			time_window_minutes = $5,
			cooldown_minutes = $6,
			enabled = $7,
			updated_at = NOW()
		RETURNING id
	`, projectID, req.MetricType, req.Threshold, req.Operator,
		req.TimeWindow, req.Cooldown, req.Enabled).Scan(&thresholdID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set threshold"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      thresholdID,
		"message": "Alert threshold set successfully",
	})
}