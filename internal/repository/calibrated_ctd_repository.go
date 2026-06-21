package repository

import (
	"ctd-calibration/internal/model"

	"gorm.io/gorm"
)

type CalibratedCTDRepository interface {
	CreateBatch(data []model.CalibratedCTD) error
	GetByDiveID(diveID string, page, pageSize int) ([]model.CalibratedCTD, int64, error)
	GetByDiveIDAll(diveID string) ([]model.CalibratedCTD, error)
	DeleteByDiveID(diveID string) error
	ExistsByDiveID(diveID string) (bool, error)
}

type calibratedCTDRepository struct {
	db *gorm.DB
}

func NewCalibratedCTDRepository(db *gorm.DB) CalibratedCTDRepository {
	return &calibratedCTDRepository{db: db}
}

func (r *calibratedCTDRepository) CreateBatch(data []model.CalibratedCTD) error {
	if len(data) == 0 {
		return nil
	}
	return r.db.Create(&data).Error
}

func (r *calibratedCTDRepository) GetByDiveID(diveID string, page, pageSize int) ([]model.CalibratedCTD, int64, error) {
	var data []model.CalibratedCTD
	var total int64

	offset := (page - 1) * pageSize

	err := r.db.Model(&model.CalibratedCTD{}).Where("dive_id = ?", diveID).Count(&total).Error
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

func (r *calibratedCTDRepository) GetByDiveIDAll(diveID string) ([]model.CalibratedCTD, error) {
	var data []model.CalibratedCTD
	err := r.db.Where("dive_id = ?", diveID).
		Order("timestamp ASC").
		Find(&data).Error
	return data, err
}

func (r *calibratedCTDRepository) DeleteByDiveID(diveID string) error {
	return r.db.Where("dive_id = ?", diveID).Delete(&model.CalibratedCTD{}).Error
}

func (r *calibratedCTDRepository) ExistsByDiveID(diveID string) (bool, error) {
	var count int64
	err := r.db.Model(&model.CalibratedCTD{}).Where("dive_id = ?", diveID).Count(&count).Error
	return count > 0, err
}
