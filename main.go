package main

import (
	"ctd-calibration/internal/config"
	"ctd-calibration/internal/handler"
	"ctd-calibration/internal/repository"
	"ctd-calibration/internal/service"
	"ctd-calibration/pkg/teos10"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	db, err := repository.InitDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	log.Println("Database initialized successfully")

	rawRepo := repository.NewRawCTDRepository(db)
	calibratedRepo := repository.NewCalibratedCTDRepository(db)

	calculator := teos10.NewTEOS10Calculator(cfg.Latitude)
	breaker := service.NewCircuitBreaker(cfg.LowConductivityThreshold, cfg.LowConductivityStreakLimit)

	rawDataService := service.NewRawDataService(rawRepo)
	calibrationService := service.NewCalibrationService(rawRepo, calibratedRepo, calculator, breaker)

	rawDataHandler := handler.NewRawDataHandler(rawDataService)
	calibrationHandler := handler.NewCalibrationHandler(calibrationService)

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "CTD Calibration Backend",
		})
	})

	handler.SetupRoutes(router, rawDataHandler, calibrationHandler)

	log.Printf("Server starting on port %s...", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
