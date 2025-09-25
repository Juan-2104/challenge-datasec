package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"database-classifier/internal/domain"
)

type ScanHandler struct {
	scanService domain.ScanService
}

func NewScanHandler(scanService domain.ScanService) *ScanHandler {
	return &ScanHandler{
		scanService: scanService,
	}
}

// StartScan handles POST /api/v1/database/:id/scan
func (h *ScanHandler) StartScan(c *gin.Context) {
	idParam := c.Param("id")
	databaseID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid database ID",
		})
		return
	}

	scanID, err := h.scanService.StartScan(c.Request.Context(), databaseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to start scan",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"scan_id": scanID.String(),
		"message": "Scan started successfully",
		"status":  "pending",
	})
}

// GetScanResult handles GET /api/v1/scan/:scanId
func (h *ScanHandler) GetScanResult(c *gin.Context) {
	scanIDParam := c.Param("scanId")
	scanID, err := uuid.Parse(scanIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid scan ID",
		})
		return
	}

	result, err := h.scanService.GetScanResult(c.Request.Context(), scanID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Scan result not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetScanHistory handles GET /api/v1/database/:id/scan/history
func (h *ScanHandler) GetScanHistory(c *gin.Context) {
	idParam := c.Param("id")
	databaseID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid database ID",
		})
		return
	}

	// Get limit from query parameter (default: 10)
	limitParam := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitParam)
	if err != nil || limit <= 0 {
		limit = 10
	}

	results, err := h.scanService.GetScanHistory(c.Request.Context(), databaseID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get scan history",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"scans": results,
		"total": len(results),
		"limit": limit,
	})
}

// GetLatestClassification handles GET /api/v1/database/:id/classification
func (h *ScanHandler) GetLatestClassification(c *gin.Context) {
	idParam := c.Param("id")
	databaseID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid database ID",
		})
		return
	}

	result, err := h.scanService.GetLatestClassification(c.Request.Context(), databaseID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "No classification found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// CancelScan handles POST /api/v1/scan/:scanId/cancel
func (h *ScanHandler) CancelScan(c *gin.Context) {
	scanIDParam := c.Param("scanId")
	scanID, err := uuid.Parse(scanIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid scan ID",
		})
		return
	}

	err = h.scanService.CancelScan(c.Request.Context(), scanID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to cancel scan",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Scan cancelled successfully",
	})
}
