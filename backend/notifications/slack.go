package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SlackClient handles Slack notifications
type SlackClient struct {
	WebhookURL string
	Channel    string
	Username   string
	IconEmoji  string
}

// NewSlackClient creates a new Slack notification client
func NewSlackClient(webhookURL string) *SlackClient {
	return &SlackClient{
		WebhookURL: webhookURL,
		Username:   "EvalForge",
		IconEmoji:  ":chart_with_upwards_trend:",
	}
}

// SlackMessage represents a Slack message
type SlackMessage struct {
	Channel     string       `json:"channel,omitempty"`
	Username    string       `json:"username,omitempty"`
	Text        string       `json:"text"`
	IconEmoji   string       `json:"icon_emoji,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

// Attachment represents a Slack message attachment
type Attachment struct {
	Color      string  `json:"color,omitempty"`
	Title      string  `json:"title,omitempty"`
	TitleLink  string  `json:"title_link,omitempty"`
	Text       string  `json:"text,omitempty"`
	Pretext    string  `json:"pretext,omitempty"`
	Footer     string  `json:"footer,omitempty"`
	FooterIcon string  `json:"footer_icon,omitempty"`
	Timestamp  int64   `json:"ts,omitempty"`
	Fields     []Field `json:"fields,omitempty"`
}

// Field represents a field in a Slack attachment
type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// Send sends a notification to Slack
func (s *SlackClient) Send(message *SlackMessage) error {
	if s.WebhookURL == "" {
		return fmt.Errorf("slack webhook URL not configured")
	}

	// Set defaults
	if message.Username == "" {
		message.Username = s.Username
	}
	if message.IconEmoji == "" {
		message.IconEmoji = s.IconEmoji
	}
	if message.Channel != "" && s.Channel != "" {
		message.Channel = s.Channel
	}

	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal slack message: %w", err)
	}

	resp, err := http.Post(s.WebhookURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to send slack notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// SendEvaluationComplete sends a notification when an evaluation completes
func (s *SlackClient) SendEvaluationComplete(evalName string, projectName string, successRate float64, totalSamples int) error {
	color := "good"
	if successRate < 0.8 {
		color = "warning"
	}
	if successRate < 0.6 {
		color = "danger"
	}

	message := &SlackMessage{
		Text: fmt.Sprintf("Evaluation completed for project *%s*", projectName),
		Attachments: []Attachment{
			{
				Color:     color,
				Title:     evalName,
				Timestamp: time.Now().Unix(),
				Fields: []Field{
					{
						Title: "Success Rate",
						Value: fmt.Sprintf("%.1f%%", successRate*100),
						Short: true,
					},
					{
						Title: "Total Samples",
						Value: fmt.Sprintf("%d", totalSamples),
						Short: true,
					},
				},
				Footer:     "EvalForge",
				FooterIcon: "https://evalforge.com/icon.png",
			},
		},
	}

	return s.Send(message)
}

// SendErrorAlert sends an error alert notification
func (s *SlackClient) SendErrorAlert(projectName string, errorMessage string, errorRate float64) error {
	message := &SlackMessage{
		Text: fmt.Sprintf("⚠️ High error rate detected in project *%s*", projectName),
		Attachments: []Attachment{
			{
				Color:     "danger",
				Title:     "Error Alert",
				Text:      errorMessage,
				Timestamp: time.Now().Unix(),
				Fields: []Field{
					{
						Title: "Error Rate",
						Value: fmt.Sprintf("%.1f%%", errorRate*100),
						Short: true,
					},
					{
						Title: "Project",
						Value: projectName,
						Short: true,
					},
				},
				Footer:     "EvalForge",
				FooterIcon: "https://evalforge.com/icon.png",
			},
		},
	}

	return s.Send(message)
}

// SendCostAlert sends a cost threshold alert
func (s *SlackClient) SendCostAlert(projectName string, currentCost float64, threshold float64) error {
	message := &SlackMessage{
		Text: fmt.Sprintf("💰 Cost threshold exceeded for project *%s*", projectName),
		Attachments: []Attachment{
			{
				Color:     "warning",
				Title:     "Cost Alert",
				Text:      fmt.Sprintf("Project has exceeded the configured cost threshold"),
				Timestamp: time.Now().Unix(),
				Fields: []Field{
					{
						Title: "Current Cost",
						Value: fmt.Sprintf("$%.2f", currentCost),
						Short: true,
					},
					{
						Title: "Threshold",
						Value: fmt.Sprintf("$%.2f", threshold),
						Short: true,
					},
					{
						Title: "Overage",
						Value: fmt.Sprintf("$%.2f (%.1f%%)", currentCost-threshold, ((currentCost-threshold)/threshold)*100),
						Short: true,
					},
				},
				Footer:     "EvalForge",
				FooterIcon: "https://evalforge.com/icon.png",
			},
		},
	}

	return s.Send(message)
}

// SendPerformanceAlert sends a performance degradation alert
func (s *SlackClient) SendPerformanceAlert(projectName string, metric string, currentValue float64, threshold float64) error {
	message := &SlackMessage{
		Text: fmt.Sprintf("🐌 Performance degradation detected in project *%s*", projectName),
		Attachments: []Attachment{
			{
				Color:     "warning",
				Title:     "Performance Alert",
				Text:      fmt.Sprintf("Performance metric '%s' has exceeded threshold", metric),
				Timestamp: time.Now().Unix(),
				Fields: []Field{
					{
						Title: "Metric",
						Value: metric,
						Short: true,
					},
					{
						Title: "Current Value",
						Value: fmt.Sprintf("%.2f ms", currentValue),
						Short: true,
					},
					{
						Title: "Threshold",
						Value: fmt.Sprintf("%.2f ms", threshold),
						Short: true,
					},
				},
				Footer:     "EvalForge",
				FooterIcon: "https://evalforge.com/icon.png",
			},
		},
	}

	return s.Send(message)
}

// SendABTestResult sends A/B test results notification
func (s *SlackClient) SendABTestResult(testName string, variantA, variantB string, winner string, improvement float64) error {
	message := &SlackMessage{
		Text: fmt.Sprintf("🎯 A/B Test completed: *%s*", testName),
		Attachments: []Attachment{
			{
				Color:     "good",
				Title:     "Test Results",
				Timestamp: time.Now().Unix(),
				Fields: []Field{
					{
						Title: "Variant A",
						Value: variantA,
						Short: true,
					},
					{
						Title: "Variant B", 
						Value: variantB,
						Short: true,
					},
					{
						Title: "Winner",
						Value: fmt.Sprintf("🏆 %s", winner),
						Short: true,
					},
					{
						Title: "Improvement",
						Value: fmt.Sprintf("+%.1f%%", improvement),
						Short: true,
					},
				},
				Footer:     "EvalForge",
				FooterIcon: "https://evalforge.com/icon.png",
			},
		},
	}

	return s.Send(message)
}

// SendDailyReport sends a daily summary report
func (s *SlackClient) SendDailyReport(projectName string, metrics map[string]interface{}) error {
	fields := []Field{}
	
	if totalEvents, ok := metrics["total_events"].(int); ok {
		fields = append(fields, Field{
			Title: "Total Events",
			Value: fmt.Sprintf("%d", totalEvents),
			Short: true,
		})
	}
	
	if avgLatency, ok := metrics["avg_latency"].(float64); ok {
		fields = append(fields, Field{
			Title: "Avg Latency",
			Value: fmt.Sprintf("%.2f ms", avgLatency),
			Short: true,
		})
	}
	
	if totalCost, ok := metrics["total_cost"].(float64); ok {
		fields = append(fields, Field{
			Title: "Total Cost",
			Value: fmt.Sprintf("$%.2f", totalCost),
			Short: true,
		})
	}
	
	if errorRate, ok := metrics["error_rate"].(float64); ok {
		fields = append(fields, Field{
			Title: "Error Rate",
			Value: fmt.Sprintf("%.2f%%", errorRate*100),
			Short: true,
		})
	}

	message := &SlackMessage{
		Text: fmt.Sprintf("📊 Daily Report for *%s*", projectName),
		Attachments: []Attachment{
			{
				Color:     "good",
				Title:     "Daily Metrics Summary",
				Timestamp: time.Now().Unix(),
				Fields:    fields,
				Footer:     "EvalForge",
				FooterIcon: "https://evalforge.com/icon.png",
			},
		},
	}

	return s.Send(message)
}