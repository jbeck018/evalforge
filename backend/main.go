package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v4"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/crypto/bcrypt"
)

// Configuration
type Config struct {
	Port         string
	PostgresURL  string
	ClickHouseURL string
	RedisURL     string
	JWTSecret    string
}

// Models
type User struct {
	ID        int       `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password_hash"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Project struct {
	ID          int       `json:"id" db:"id"`
	UserID      int       `json:"user_id" db:"user_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type TraceEvent struct {
	ID           string                 `json:"id"`
	ProjectID    int                    `json:"project_id"`
	TraceID      string                 `json:"trace_id"`
	SpanID       string                 `json:"span_id"`
	ParentSpanID string                 `json:"parent_span_id,omitempty"`
	OperationType string                `json:"operation_type"`
	StartTime    time.Time              `json:"start_time"`
	EndTime      time.Time              `json:"end_time"`
	Duration     int64                  `json:"duration_ms"`
	Status       string                 `json:"status"`
	Input        map[string]interface{} `json:"input"`
	Output       map[string]interface{} `json:"output"`
	Metadata     map[string]interface{} `json:"metadata"`
	Tokens       TokenUsage             `json:"tokens"`
	Cost         float64                `json:"cost"`
	Provider     string                 `json:"provider"`
	Model        string                 `json:"model"`
}

type TokenUsage struct {
	Prompt     int `json:"prompt"`
	Completion int `json:"completion"`
	Total      int `json:"total"`
}

type EventIngestionRequest struct {
	Events []TraceEvent `json:"events"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Metrics
var (
	eventCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "evalforge_events_total",
			Help: "Total number of events processed",
		},
		[]string{"project_id", "status"},
	)
	
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "evalforge_request_duration_seconds",
			Help: "Duration of HTTP requests",
		},
		[]string{"method", "endpoint"},
	)
)

func init() {
	prometheus.MustRegister(eventCounter)
	prometheus.MustRegister(requestDuration)
}

// Database connections
type DB struct {
	Postgres   *sql.DB
	ClickHouse clickhouse.Conn
	Redis      *redis.Client
}

// Server
type Server struct {
	DB     *DB
	Config *Config
	Router *gin.Engine
}

func main() {
	// Load configuration
	config := &Config{
		Port:          getEnv("PORT", "8000"),
		PostgresURL:   getEnv("POSTGRES_URL", "postgres://evalforge:password@localhost:5432/evalforge?sslmode=disable"),
		ClickHouseURL: getEnv("CLICKHOUSE_URL", "clickhouse://localhost:9000/evalforge"),
		RedisURL:      getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:     getEnv("JWT_SECRET", "your-secret-key"),
	}

	// Initialize database connections
	db, err := initDB(config)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Initialize server
	server := &Server{
		DB:     db,
		Config: config,
	}

	server.setupRoutes()

	// Start server
	srv := &http.Server{
		Addr:    ":" + config.Port,
		Handler: server.Router,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
	log.Printf("Server started on port %s", config.Port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("Server exited")
}

func initDB(config *Config) (*DB, error) {
	// PostgreSQL connection
	pg, err := sql.Open("postgres", config.PostgresURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	if err := pg.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	// ClickHouse connection
	ch, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{"localhost:9000"},
		Auth: clickhouse.Auth{
			Database: "evalforge",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}

	if err := ch.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping ClickHouse: %w", err)
	}

	// Redis connection
	opt, err := redis.ParseURL(config.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}
	rdb := redis.NewClient(opt)

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	return &DB{
		Postgres:   pg,
		ClickHouse: ch,
		Redis:      rdb,
	}, nil
}

func (db *DB) Close() {
	if db.Postgres != nil {
		db.Postgres.Close()
	}
	if db.ClickHouse != nil {
		db.ClickHouse.Close()
	}
	if db.Redis != nil {
		db.Redis.Close()
	}
}

func (s *Server) setupRoutes() {
	s.Router = gin.Default()

	// CORS middleware
	s.Router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
	}))

	// Prometheus metrics
	s.Router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Health check
	s.Router.GET("/health", s.healthCheck)

	// Authentication routes
	auth := s.Router.Group("/api/auth")
	{
		auth.POST("/register", s.register)
		auth.POST("/login", s.login)
	}

	// Protected routes
	api := s.Router.Group("/api")
	api.Use(s.authMiddleware())
	{
		// Projects
		api.GET("/projects", s.getProjects)
		api.POST("/projects", s.createProject)
		api.GET("/projects/:id", s.getProject)

		// Event ingestion
		api.POST("/events", s.ingestEvents)

		// Analytics
		api.GET("/projects/:id/analytics", s.getAnalytics)
		api.GET("/projects/:id/traces", s.getTraces)
	}
}

func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
	})
}

func (s *Server) register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Insert user
	var userID int
	err = s.DB.Postgres.QueryRow(
		"INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id",
		req.Email, string(hashedPassword),
	).Scan(&userID)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Generate JWT token
	token, err := s.generateJWT(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token":   token,
		"user_id": userID,
	})
}

func (s *Server) login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Get user
	var user User
	err := s.DB.Postgres.QueryRow(
		"SELECT id, email, password_hash FROM users WHERE email = $1",
		req.Email,
	).Scan(&user.ID, &user.Email, &user.Password)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token
	token, err := s.generateJWT(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":   token,
		"user_id": user.ID,
	})
}

func (s *Server) getProjects(c *gin.Context) {
	userID := c.GetInt("user_id")

	rows, err := s.DB.Postgres.Query(
		"SELECT id, name, description, created_at FROM projects WHERE user_id = $1 ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch projects"})
		return
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		p.UserID = userID
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan project"})
			return
		}
		projects = append(projects, p)
	}

	c.JSON(http.StatusOK, gin.H{"projects": projects})
}

func (s *Server) createProject(c *gin.Context) {
	userID := c.GetInt("user_id")
	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var project Project
	err := s.DB.Postgres.QueryRow(
		"INSERT INTO projects (user_id, name, description) VALUES ($1, $2, $3) RETURNING id, created_at",
		userID, req.Name, req.Description,
	).Scan(&project.ID, &project.CreatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
		return
	}

	project.UserID = userID
	project.Name = req.Name
	project.Description = req.Description

	c.JSON(http.StatusCreated, gin.H{"project": project})
}

func (s *Server) getProject(c *gin.Context) {
	userID := c.GetInt("user_id")
	projectID := c.Param("id")

	var project Project
	err := s.DB.Postgres.QueryRow(
		"SELECT id, name, description, created_at FROM projects WHERE id = $1 AND user_id = $2",
		projectID, userID,
	).Scan(&project.ID, &project.Name, &project.Description, &project.CreatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch project"})
		return
	}

	project.UserID = userID
	c.JSON(http.StatusOK, gin.H{"project": project})
}

func (s *Server) ingestEvents(c *gin.Context) {
	userID := c.GetInt("user_id")
	var req EventIngestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Verify project ownership
	for _, event := range req.Events {
		var exists bool
		err := s.DB.Postgres.QueryRow(
			"SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1 AND user_id = $2)",
			event.ProjectID, userID,
		).Scan(&exists)

		if err != nil || !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid project access"})
			return
		}
	}

	// Insert events into ClickHouse
	batch, err := s.DB.ClickHouse.PrepareBatch(context.Background(), `
		INSERT INTO trace_events (
			id, project_id, trace_id, span_id, parent_span_id, operation_type,
			start_time, end_time, duration_ms, status, input, output, metadata,
			prompt_tokens, completion_tokens, total_tokens, cost, provider, model
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare batch"})
		return
	}

	for _, event := range req.Events {
		inputJSON, _ := json.Marshal(event.Input)
		outputJSON, _ := json.Marshal(event.Output)
		metadataJSON, _ := json.Marshal(event.Metadata)

		err = batch.Append(
			event.ID,
			event.ProjectID,
			event.TraceID,
			event.SpanID,
			event.ParentSpanID,
			event.OperationType,
			event.StartTime,
			event.EndTime,
			event.Duration,
			event.Status,
			string(inputJSON),
			string(outputJSON),
			string(metadataJSON),
			event.Tokens.Prompt,
			event.Tokens.Completion,
			event.Tokens.Total,
			event.Cost,
			event.Provider,
			event.Model,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to append event to batch"})
			return
		}

		// Update metrics
		eventCounter.WithLabelValues(strconv.Itoa(event.ProjectID), event.Status).Inc()
	}

	if err := batch.Send(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert events"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":        "Events ingested successfully",
		"events_count":   len(req.Events),
		"ingested_at":    time.Now(),
	})
}

func (s *Server) getAnalytics(c *gin.Context) {
	userID := c.GetInt("user_id")
	projectID := c.Param("id")

	// Verify project ownership
	var exists bool
	err := s.DB.Postgres.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1 AND user_id = $2)",
		projectID, userID,
	).Scan(&exists)

	if err != nil || !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid project access"})
		return
	}

	// Get analytics from ClickHouse
	ctx := context.Background()
	
	// Total events
	var totalEvents uint64
	err = s.DB.ClickHouse.QueryRow(ctx, 
		"SELECT count() FROM trace_events WHERE project_id = ?", 
		projectID,
	).Scan(&totalEvents)
	if err != nil {
		log.Printf("Error fetching total events: %v", err)
	}

	// Total cost
	var totalCost float64
	err = s.DB.ClickHouse.QueryRow(ctx, 
		"SELECT sum(cost) FROM trace_events WHERE project_id = ?", 
		projectID,
	).Scan(&totalCost)
	if err != nil {
		log.Printf("Error fetching total cost: %v", err)
	}

	// Average latency
	var avgLatency float64
	err = s.DB.ClickHouse.QueryRow(ctx, 
		"SELECT avg(duration_ms) FROM trace_events WHERE project_id = ?", 
		projectID,
	).Scan(&avgLatency)
	if err != nil {
		log.Printf("Error fetching average latency: %v", err)
	}

	// Error rate
	var errorCount uint64
	err = s.DB.ClickHouse.QueryRow(ctx, 
		"SELECT count() FROM trace_events WHERE project_id = ? AND status = 'error'", 
		projectID,
	).Scan(&errorCount)
	if err != nil {
		log.Printf("Error fetching error count: %v", err)
	}

	errorRate := float64(0)
	if totalEvents > 0 {
		errorRate = float64(errorCount) / float64(totalEvents) * 100
	}

	c.JSON(http.StatusOK, gin.H{
		"analytics": gin.H{
			"total_events":    totalEvents,
			"total_cost":      totalCost,
			"average_latency": avgLatency,
			"error_rate":      errorRate,
		},
	})
}

func (s *Server) getTraces(c *gin.Context) {
	userID := c.GetInt("user_id")
	projectID := c.Param("id")

	// Verify project ownership
	var exists bool
	err := s.DB.Postgres.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1 AND user_id = $2)",
		projectID, userID,
	).Scan(&exists)

	if err != nil || !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid project access"})
		return
	}

	limit := 100
	if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 && l <= 1000 {
		limit = l
	}

	ctx := context.Background()
	rows, err := s.DB.ClickHouse.Query(ctx, `
		SELECT id, trace_id, span_id, operation_type, start_time, end_time, 
		       duration_ms, status, cost, provider, model
		FROM trace_events 
		WHERE project_id = ? 
		ORDER BY start_time DESC 
		LIMIT ?
	`, projectID, limit)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch traces"})
		return
	}
	defer rows.Close()

	var traces []map[string]interface{}
	for rows.Next() {
		var (
			id, traceID, spanID, operationType, status, provider, model string
			startTime, endTime                                          time.Time
			duration                                                    int64
			cost                                                        float64
		)

		if err := rows.Scan(&id, &traceID, &spanID, &operationType, &startTime, &endTime, &duration, &status, &cost, &provider, &model); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan trace"})
			return
		}

		traces = append(traces, map[string]interface{}{
			"id":             id,
			"trace_id":       traceID,
			"span_id":        spanID,
			"operation_type": operationType,
			"start_time":     startTime,
			"end_time":       endTime,
			"duration_ms":    duration,
			"status":         status,
			"cost":           cost,
			"provider":       provider,
			"model":          model,
		})
	}

	c.JSON(http.StatusOK, gin.H{"traces": traces})
}

func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := authHeader
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(s.Config.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			userID := int(claims["user_id"].(float64))
			c.Set("user_id", userID)
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (s *Server) generateJWT(userID int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 days
	})

	return token.SignedString([]byte(s.Config.JWTSecret))
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}