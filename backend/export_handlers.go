package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Export handler methods for Server

func (s *Server) handleExport(c *gin.Context) {
	s.ExportHandler.HandleExport(c)
}

func (s *Server) handleExportStatus(c *gin.Context) {
	s.ExportHandler.HandleExportStatus(c)
}

func (s *Server) handleScheduledExport(c *gin.Context) {
	s.ExportHandler.HandleScheduledExport(c)
}

func (s *Server) listExportTemplates(c *gin.Context) {
	userID := c.GetInt("user_id")
	
	rows, err := s.DB.Postgres.Query(`
		SELECT id, name, description, export_config, is_public, created_at
		FROM export_templates
		WHERE user_id = $1 OR is_public = true
		ORDER BY created_at DESC
	`, userID)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch templates"})
		return
	}
	defer rows.Close()
	
	var templates []gin.H
	for rows.Next() {
		var id int
		var name, description, exportConfig string
		var isPublic bool
		var createdAt string
		
		err := rows.Scan(&id, &name, &description, &exportConfig, &isPublic, &createdAt)
		if err != nil {
			continue
		}
		
		templates = append(templates, gin.H{
			"id":            id,
			"name":          name,
			"description":   description,
			"export_config": exportConfig,
			"is_public":     isPublic,
			"created_at":    createdAt,
		})
	}
	
	c.JSON(http.StatusOK, gin.H{"templates": templates})
}

func (s *Server) createExportTemplate(c *gin.Context) {
	userID := c.GetInt("user_id")
	
	var req struct {
		Name         string      `json:"name" binding:"required"`
		Description  string      `json:"description"`
		ExportConfig interface{} `json:"export_config" binding:"required"`
		IsPublic     bool        `json:"is_public"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	var templateID int
	err := s.DB.Postgres.QueryRow(`
		INSERT INTO export_templates (user_id, name, description, export_config, is_public)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, userID, req.Name, req.Description, req.ExportConfig, req.IsPublic).Scan(&templateID)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create template"})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"id":      templateID,
		"message": "Export template created successfully",
	})
}