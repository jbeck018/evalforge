package main

import (
	"github.com/gin-gonic/gin"
)

// Custom metrics handler methods for Server

func (s *Server) getCustomMetrics(c *gin.Context) {
	s.CustomMetricsHandler.GetCustomMetrics(c)
}

func (s *Server) createCustomMetric(c *gin.Context) {
	s.CustomMetricsHandler.CreateCustomMetric(c)
}

func (s *Server) updateCustomMetric(c *gin.Context) {
	s.CustomMetricsHandler.UpdateCustomMetric(c)
}

func (s *Server) deleteCustomMetric(c *gin.Context) {
	s.CustomMetricsHandler.DeleteCustomMetric(c)
}

func (s *Server) getMetricTemplates(c *gin.Context) {
	s.CustomMetricsHandler.GetMetricTemplates(c)
}

func (s *Server) createMetricFromTemplate(c *gin.Context) {
	s.CustomMetricsHandler.CreateMetricFromTemplate(c)
}

func (s *Server) getCustomMetricResults(c *gin.Context) {
	s.CustomMetricsHandler.GetMetricResults(c)
}