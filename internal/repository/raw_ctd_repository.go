package repository

import (
	"ctd-calibration/internal/model"

	"gorm.io/gorm"
)

type RawCTDRepository interface {
	CreateBatch(rawData []model.RawCTD) error
	GetByDiveID(diveID string, page, pageSize int) ([]model.RawCTD, int64, error)
	GetByDiveIDAll(diveID string) ([]model.RawCTD, error)
	ListDiveIDs() ([]string, error)
	DeleteByDiveID(diveID string) error
}

type rawCTDRepository struct {
	db *gorm.DB
}

func NewRawCTDRepository(db *gorm.DB) RawCTDRepository {
	return &rawCTDRepository{db: db}
}

func (r *rawCTDRepository) CreateBatch(rawData []model.RawCTD) error {
	if len(rawData) == 0 {
		return nil
	}
	return r.db.Create(&rawData).Error
}

func (r *rawCTDRepository) GetByDiveID(diveID string, page, pageSize int) ([]model.RawCTD, int64, error) {
	var data []model.RawCTD
	var total int64

	offset := (page - 1) * pageSize

	err := r.db.Model(&model.RawCTD{}).Where("dive_id = ?", diveID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where("dive_id = ?", diveID).
		Order("timestamp ASC").
		Offset(offset).
		Limit(pageSize).
		Find(&data).Error

	return data, total, err
}

func (r *rawCTDRepository) GetByDiveIDAll(diveID string) ([]model.RawCTD, error) {
	var data []model.RawCTD
	err := r.db.Where("dive_id = ?", diveID).
		Order("timestamp ASC").
		Find(&data).Error
	return data, err
}

func (r *rawCTDRepository) ListDiveIDs() ([]string, error) {
	var diveIDs []string
	err := r.db.Model(&model.RawCTD{}).
		Distinct("dive_id").
		Order("dive_id ASC").
		Pluck("dive_id", &diveIDs).Error
	return diveIDs, err
}

func (r *rawCTDRepository) DeleteByDiveID(diveID string) error {
	return r.db.Where("dive_id = ?", diveID).Delete(&model.RawCTD{}).Error
}
