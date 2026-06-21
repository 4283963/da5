package handler

import (
	"ctd-calibration/internal/model"
	"ctd-calibration/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RawDataHandler struct {
	rawDataService service.RawDataService
}

func NewRawDataHandler(rawDataService service.RawDataService) *RawDataHandler {
	return &RawDataHandler{rawDataService: rawDataService}
}

func (h *RawDataHandler) ReceiveBatch(c *gin.Context) {
	var req model.RawCTDBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	count, err := h.rawDataService.ReceiveBatch(req.DiveID, req.Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to save raw data",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"dive_id": req.DiveID,
		"count":   count,
		"message": "raw data received successfully",
	})
}

func (h *RawDataHandler) ListRawData(c *gin.Context) {
	diveID := c.Param("dive_id")
	if diveID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dive_id is required"})
		return
	}

	var pagination model.Pagination
	if err := c.ShouldBindQuery(&pagination); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	data, total, err := h.rawDataService.ListRawData(diveID, pagination.Page, pagination.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"dive_id":   diveID,
		"data":      data,
		"total":     total,
		"page":      pagination.Page,
		"page_size": pagination.PageSize,
	})
}

func (h *RawDataHandler) ListDives(c *gin.Context) {
	diveIDs, err := h.rawDataService.ListDiveIDs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"dives": diveIDs,
		"count": len(diveIDs),
	})
}

func (h *RawDataHandler) DeleteDive(c *gin.Context) {
	diveID := c.Param("dive_id")
	if diveID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dive_id is required"})
		return
	}

	err := h.rawDataService.DeleteDiveData(diveID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "dive data deleted successfully",
		"dive_id": diveID,
	})
}
