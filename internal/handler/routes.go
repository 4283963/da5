package handler

import (
	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	router *gin.Engine,
	rawDataHandler *RawDataHandler,
	calibrationHandler *CalibrationHandler,
) {
	api := router.Group("/api/v1")
	{
		raw := api.Group("/raw")
		{
			raw.POST("/batch", rawDataHandler.ReceiveBatch)
			raw.GET("/dives", rawDataHandler.ListDives)
			raw.GET("/:dive_id", rawDataHandler.ListRawData)
			raw.DELETE("/:dive_id", rawDataHandler.DeleteDive)
		}

		cal := api.Group("/calibrated")
		{
			cal.POST("/calibrate", calibrationHandler.Calibrate)
			cal.POST("/recalibrate", calibrationHandler.Recalibrate)
			cal.GET("/:dive_id", calibrationHandler.ListCalibrated)
			cal.DELETE("/:dive_id", calibrationHandler.DeleteCalibrated)
		}
	}
}
