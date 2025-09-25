package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"database-classifier/internal/domain"
)

type DatabaseHandler struct {
	databaseService domain.DatabaseService
}

func NewDatabaseHandler(databaseService domain.DatabaseService) *DatabaseHandler {
	return &DatabaseHandler{
		databaseService: databaseService,
	}
}

// CreateDatabase handles POST /api/v1/database
func (h *DatabaseHandler) CreateDatabase(c *gin.Context) {
	var req domain.CreateDatabaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	id, err := h.databaseService.CreateConnection(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create database connection",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id": id.String(),
	})
}

// GetDatabase handles GET /api/v1/database/:id
func (h *DatabaseHandler) GetDatabase(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid database ID",
		})
		return
	}

	conn, err := h.databaseService.GetConnection(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Database connection not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, conn)
}

// GetAllDatabases handles GET /api/v1/database
func (h *DatabaseHandler) GetAllDatabases(c *gin.Context) {
	connections, err := h.databaseService.GetAllConnections(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get database connections",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"databases": connections,
		"total":     len(connections),
	})
}

// UpdateDatabase handles PUT /api/v1/database/:id
func (h *DatabaseHandler) UpdateDatabase(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid database ID",
		})
		return
	}

	var req domain.CreateDatabaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	err = h.databaseService.UpdateConnection(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update database connection",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Database connection updated successfully",
	})
}

// DeleteDatabase handles DELETE /api/v1/database/:id
func (h *DatabaseHandler) DeleteDatabase(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid database ID",
		})
		return
	}

	err = h.databaseService.DeleteConnection(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete database connection",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Database connection deleted successfully",
	})
}

// TestDatabase handles POST /api/v1/database/:id/test
func (h *DatabaseHandler) TestDatabase(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid database ID",
		})
		return
	}

	err = h.databaseService.TestConnection(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Connection test failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Connection test successful",
	})
}
