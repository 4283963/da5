package handler

import (
	"ctd-calibration/internal/model"
	"ctd-calibration/internal/service"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CalibrationHandler struct {
	calibrationService service.CalibrationService
}

func NewCalibrationHandler(calibrationService service.CalibrationService) *CalibrationHandler {
	return &CalibrationHandler{calibrationService: calibrationService}
}

func (h *CalibrationHandler) Calibrate(c *gin.Context) {
	var req model.CalibrateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	result, err := h.calibrationService.CalibrateDive(req.DiveID)
	if err != nil {
		if errors.Is(err, service.ErrCircuitBreakerTripped) {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   err.Error(),
				"dive_id": req.DiveID,
				"hint":    "请擦洗传感器后调用重置接口恢复校准功能",
			})
			return
		}
		if err.Error() == "calibrated data already exists for this dive, use recalibrate instead" {
			c.JSON(http.StatusConflict, gin.H{
				"error":   err.Error(),
				"dive_id": req.DiveID,
			})
			return
		}
		if err.Error() == "no raw data found for this dive" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   err.Error(),
				"dive_id": req.DiveID,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "calibration failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *CalibrationHandler) Recalibrate(c *gin.Context) {
	var req model.CalibrateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	result, err := h.calibrationService.RecalibrateDive(req.DiveID)
	if err != nil {
		if errors.Is(err, service.ErrCircuitBreakerTripped) {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   err.Error(),
				"dive_id": req.DiveID,
				"hint":    "请擦洗传感器后调用重置接口恢复校准功能",
			})
			return
		}
		if err.Error() == "no raw data found for this dive" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   err.Error(),
				"dive_id": req.DiveID,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "recalibration failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *CalibrationHandler) ListCalibrated(c *gin.Context) {
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

	data, total, err := h.calibrationService.ListCalibratedData(diveID, pagination.Page, pagination.PageSize)
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

func (h *CalibrationHandler) DeleteCalibrated(c *gin.Context) {
	diveID := c.Param("dive_id")
	if diveID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dive_id is required"})
		return
	}

	err := h.calibrationService.DeleteCalibratedData(diveID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "calibrated data deleted successfully",
		"dive_id": diveID,
	})
}

func (h *CalibrationHandler) GetCircuitBreaker(c *gin.Context) {
	status := h.calibrationService.GetCircuitBreakerStatus()
	c.JSON(http.StatusOK, status)
}

func (h *CalibrationHandler) ResetCircuitBreaker(c *gin.Context) {
	h.calibrationService.ResetCircuitBreaker()
	c.JSON(http.StatusOK, gin.H{
		"message": "熔断器已重置，校准功能已恢复",
		"status":  h.calibrationService.GetCircuitBreakerStatus(),
	})
}
