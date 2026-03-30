package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/livingopensource/bind-dns-api/internal/bind"
	"github.com/livingopensource/bind-dns-api/internal/models"
)

const Version = "1.0.0"

// Handler holds the API handlers
type Handler struct {
	manager *bind.Manager
}

// NewHandler creates a new API handler
func NewHandler(manager *bind.Manager) *Handler {
	return &Handler{
		manager: manager,
	}
}

// RegisterRoutes registers all API routes
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")
	{
		// Health check
		api.GET("/health", h.healthCheck)

		// Domain endpoints
		domains := api.Group("/domains")
		{
			domains.GET("", h.listDomains)
			domains.GET("/:name", h.getDomain)
			domains.POST("", h.createDomain)
			domains.PUT("/:name", h.updateDomain)
			domains.DELETE("/:name", h.deleteDomain)

			// Record endpoints
			records := domains.Group("/:name/records")
			{
				records.GET("", h.listRecords)
				records.POST("", h.addRecord)
				records.PUT("/:recordName/:recordType", h.updateRecord)
				records.DELETE("/:recordName/:recordType", h.deleteRecord)
			}

			// Zone reload endpoint
			domains.POST("/:name/reload", h.reloadZone)
		}

		// Global reload endpoint
		api.POST("/reload", h.reloadAll)
	}
}

// healthCheck returns the health status of the API
func (h *Handler) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, models.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   Version,
	})
}

// listDomains returns all managed domains
func (h *Handler) listDomains(c *gin.Context) {
	domains, err := h.manager.ListDomains()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    domains,
	})
}

// getDomain returns details for a specific domain
func (h *Handler) getDomain(c *gin.Context) {
	name := c.Param("name")

	domain, err := h.manager.GetDomain(name)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    domain,
	})
}

// createDomain creates a new domain
func (h *Handler) createDomain(c *gin.Context) {
	var req models.CreateDomainRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
		})
		return
	}

	if err := h.manager.CreateDomain(req.Name, req); err != nil {
		c.JSON(http.StatusConflict, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "Domain created successfully",
	})
}

// updateDomain updates an existing domain
func (h *Handler) updateDomain(c *gin.Context) {
	name := c.Param("name")

	var req models.CreateDomainRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
		})
		return
	}

	if err := h.manager.UpdateDomain(name, req); err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Domain updated successfully",
	})
}

// deleteDomain deletes a domain
func (h *Handler) deleteDomain(c *gin.Context) {
	name := c.Param("name")

	if err := h.manager.DeleteDomain(name); err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Domain deleted successfully",
	})
}

// listRecords returns all DNS records for a domain
func (h *Handler) listRecords(c *gin.Context) {
	name := c.Param("name")

	records, err := h.manager.ListRecords(name)
	if err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Data:    records,
	})
}

// addRecord adds a new DNS record to a domain
func (h *Handler) addRecord(c *gin.Context) {
	name := c.Param("name")

	var req models.CreateRecordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
		})
		return
	}

	if err := h.manager.AddRecord(name, req); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "Record added successfully",
	})
}

// updateRecord updates an existing DNS record
func (h *Handler) updateRecord(c *gin.Context) {
	name := c.Param("name")
	recordName := c.Param("recordName")
	recordType := models.DNSRecordType(c.Param("recordType"))

	var req models.UpdateRecordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
		})
		return
	}

	if err := h.manager.UpdateRecord(name, recordName, recordType, req); err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Record updated successfully",
	})
}

// deleteRecord deletes a DNS record
func (h *Handler) deleteRecord(c *gin.Context) {
	name := c.Param("name")
	recordName := c.Param("recordName")
	recordType := models.DNSRecordType(c.Param("recordType"))

	if err := h.manager.DeleteRecord(name, recordName, recordType); err != nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Record deleted successfully",
	})
}

// reloadZone reloads a specific zone
func (h *Handler) reloadZone(c *gin.Context) {
	name := c.Param("name")

	if err := h.manager.ReloadZone(name); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Zone reloaded successfully",
	})
}

// reloadAll reloads all zones
func (h *Handler) reloadAll(c *gin.Context) {
	if err := h.manager.ReloadAll(); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "All zones reloaded successfully",
	})
}
