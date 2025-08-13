package main

import (
	"github.com/gin-gonic/gin"
)

// Notification handler methods for Server

func (s *Server) getNotificationConfigs(c *gin.Context) {
	s.NotificationHandler.GetNotificationConfigs(c)
}

func (s *Server) createNotificationConfig(c *gin.Context) {
	s.NotificationHandler.CreateNotificationConfig(c)
}

func (s *Server) updateNotificationConfig(c *gin.Context) {
	s.NotificationHandler.UpdateNotificationConfig(c)
}

func (s *Server) deleteNotificationConfig(c *gin.Context) {
	s.NotificationHandler.DeleteNotificationConfig(c)
}

func (s *Server) testNotification(c *gin.Context) {
	s.NotificationHandler.TestNotification(c)
}

func (s *Server) getAlertThresholds(c *gin.Context) {
	s.NotificationHandler.GetAlertThresholds(c)
}

func (s *Server) setAlertThreshold(c *gin.Context) {
	s.NotificationHandler.SetAlertThreshold(c)
}