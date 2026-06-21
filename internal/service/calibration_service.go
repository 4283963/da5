package service

import (
	"ctd-calibration/internal/model"
	"ctd-calibration/internal/repository"
	"ctd-calibration/pkg/teos10"
	"errors"
	"time"
)

type CalibrationService interface {
	CalibrateDive(diveID string) (*model.CalibrateResponse, error)
	ListCalibratedData(diveID string, page, pageSize int) ([]model.CalibratedCTD, int64, error)
	RecalibrateDive(diveID string) (*model.CalibrateResponse, error)
	DeleteCalibratedData(diveID string) error
	GetCircuitBreakerStatus() CircuitBreakerStatus
	ResetCircuitBreaker()
}

var ErrCircuitBreakerTripped = errors.New("校准功能已熔断停用：传感器异常，请擦洗传感器后重置熔断器")

type calibrationService struct {
	rawRepo        repository.RawCTDRepository
	calibratedRepo repository.CalibratedCTDRepository
	calculator     *teos10.TEOS10Calculator
	breaker        *CircuitBreaker
}

func NewCalibrationService(
	rawRepo repository.RawCTDRepository,
	calibratedRepo repository.CalibratedCTDRepository,
	calculator *teos10.TEOS10Calculator,
	breaker *CircuitBreaker,
) CalibrationService {
	return &calibrationService{
		rawRepo:        rawRepo,
		calibratedRepo: calibratedRepo,
		calculator:     calculator,
		breaker:        breaker,
	}
}

func (s *calibrationService) CalibrateDive(diveID string) (*model.CalibrateResponse, error) {
	if s.breaker.IsTripped() {
		return nil, ErrCircuitBreakerTripped
	}

	exists, err := s.calibratedRepo.ExistsByDiveID(diveID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("calibrated data already exists for this dive, use recalibrate instead")
	}

	return s.performCalibration(diveID)
}

func (s *calibrationService) RecalibrateDive(diveID string) (*model.CalibrateResponse, error) {
	if s.breaker.IsTripped() {
		return nil, ErrCircuitBreakerTripped
	}

	err := s.calibratedRepo.DeleteByDiveID(diveID)
	if err != nil {
		return nil, err
	}

	return s.performCalibration(diveID)
}

func (s *calibrationService) performCalibration(diveID string) (*model.CalibrateResponse, error) {
	rawDataList, err := s.rawRepo.GetByDiveIDAll(diveID)
	if err != nil {
		return nil, err
	}

	if len(rawDataList) == 0 {
		return nil, errors.New("no raw data found for this dive")
	}

	calibratedList := make([]model.CalibratedCTD, 0, len(rawDataList))
	now := time.Now()

	for _, raw := range rawDataList {
		if s.breaker.Observe(raw.Conductivity) {
			break
		}

		func() {
			defer func() {
				if r := recover(); r != nil {
				}
			}()

			result := s.calculator.Calibrate(raw.Conductivity, raw.Temperature, raw.Pressure)

			calibrated := model.CalibratedCTD{
				DiveID:      raw.DiveID,
				Timestamp:   raw.Timestamp,
				Temperature: raw.Temperature,
				Pressure:    raw.Pressure,
				Salinity:    result.Salinity,
				Depth:       result.Depth,
				Density:     result.Density,
				CreatedAt:   now,
			}
			calibratedList = append(calibratedList, calibrated)
		}()
	}

	if s.breaker.IsTripped() {
		return nil, ErrCircuitBreakerTripped
	}

	err = s.calibratedRepo.CreateBatch(calibratedList)
	if err != nil {
		return nil, err
	}

	return &model.CalibrateResponse{
		DiveID:       diveID,
		Count:        len(calibratedList),
		Data:         calibratedList,
		CalibratedAt: now,
	}, nil
}

func (s *calibrationService) ListCalibratedData(diveID string, page, pageSize int) ([]model.CalibratedCTD, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	return s.calibratedRepo.GetByDiveID(diveID, page, pageSize)
}

func (s *calibrationService) DeleteCalibratedData(diveID string) error {
	return s.calibratedRepo.DeleteByDiveID(diveID)
}

func (s *calibrationService) GetCircuitBreakerStatus() CircuitBreakerStatus {
	return s.breaker.Status()
}

func (s *calibrationService) ResetCircuitBreaker() {
	s.breaker.Reset()
}
