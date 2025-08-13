package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// SecurityHeaders adds security headers to responses
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent XSS attacks
		c.Header("X-XSS-Protection", "1; mode=block")
		
		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")
		
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")
		
		// Enable strict transport security (HSTS) for HTTPS
		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}
		
		// Content Security Policy
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self' ws: wss:")
		
		// Referrer Policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// Feature Policy / Permissions Policy
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=(), interest-cohort=()")
		
		c.Next()
	}
}

// InputSanitizer sanitizes user input to prevent injection attacks
func InputSanitizer() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Sanitize query parameters
		for key, values := range c.Request.URL.Query() {
			for i, value := range values {
				values[i] = sanitizeString(value)
			}
			c.Request.URL.RawQuery = strings.Replace(c.Request.URL.RawQuery, key+"=", key+"="+values[0], 1)
		}
		
		c.Next()
	}
}

// sanitizeString removes potentially dangerous characters
func sanitizeString(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")
	
	// Remove control characters
	for i := 0; i < 32; i++ {
		if i != 9 && i != 10 && i != 13 { // Keep tab, newline, carriage return
			input = strings.ReplaceAll(input, string(rune(i)), "")
		}
	}
	
	// Limit length to prevent DoS
	if len(input) > 10000 {
		input = input[:10000]
	}
	
	return input
}

// RequestSizeLimiter limits the size of incoming requests
func RequestSizeLimiter(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}

// APIKeyValidator validates API keys for SDK access
func APIKeyValidator(validateFunc func(string) (int, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			// Try Bearer token as fallback
			auth := c.GetHeader("Authorization")
			if strings.HasPrefix(auth, "Bearer proj_") {
				apiKey = strings.TrimPrefix(auth, "Bearer ")
			}
		}
		
		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "API key required"})
			return
		}
		
		projectID, err := validateFunc(apiKey)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
			return
		}
		
		c.Set("project_id", projectID)
		c.Set("api_key", apiKey)
		c.Next()
	}
}

// AuditLogger logs all API requests for audit purposes
func AuditLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Log request details
		userID, _ := c.Get("user_id")
		projectID, _ := c.Get("project_id")
		
		// Create audit log entry
		auditLog := map[string]interface{}{
			"timestamp":  c.Request.Context().Value("request_time"),
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
			"user_id":    userID,
			"project_id": projectID,
		}
		
		// Process the request
		c.Next()
		
		// Add response details
		auditLog["status_code"] = c.Writer.Status()
		auditLog["response_size"] = c.Writer.Size()
		
		// Log to audit system (implement based on your logging infrastructure)
		// For now, just set it in context for other middleware to use
		c.Set("audit_log", auditLog)
	}
}