package service

import (
	"ctd-calibration/internal/model"
	"ctd-calibration/internal/repository"
	"time"
)

type RawDataService interface {
	ReceiveBatch(diveID string, dataPoints []model.RawCTDDataPoint) (int, error)
	ListRawData(diveID string, page, pageSize int) ([]model.RawCTD, int64, error)
	ListDiveIDs() ([]string, error)
	DeleteDiveData(diveID string) error
}

type rawDataService struct {
	rawRepo repository.RawCTDRepository
}

func NewRawDataService(rawRepo repository.RawCTDRepository) RawDataService {
	return &rawDataService{rawRepo: rawRepo}
}

func (s *rawDataService) ReceiveBatch(diveID string, dataPoints []model.RawCTDDataPoint) (int, error) {
	rawDataList := make([]model.RawCTD, 0, len(dataPoints))
	now := time.Now()

	for _, dp := range dataPoints {
		rawData := model.RawCTD{
			DiveID:       diveID,
			Timestamp:    time.Unix(dp.Timestamp, 0),
			Conductivity: dp.Conductivity,
			Temperature:  dp.Temperature,
			Pressure:     dp.Pressure,
			CreatedAt:    now,
		}
		rawDataList = append(rawDataList, rawData)
	}

	err := s.rawRepo.CreateBatch(rawDataList)
	if err != nil {
		return 0, err
	}

	return len(rawDataList), nil
}

func (s *rawDataService) ListRawData(diveID string, page, pageSize int) ([]model.RawCTD, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	return s.rawRepo.GetByDiveID(diveID, page, pageSize)
}

func (s *rawDataService) ListDiveIDs() ([]string, error) {
	return s.rawRepo.ListDiveIDs()
}

func (s *rawDataService) DeleteDiveData(diveID string) error {
	return s.rawRepo.DeleteByDiveID(diveID)
}
