package model

import "time"

type RawCTD struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	DiveID       string    `gorm:"index;size:64" json:"dive_id"`
	Timestamp    time.Time `gorm:"index" json:"timestamp"`
	Conductivity float64   `gorm:"not null" json:"conductivity"`
	Temperature  float64   `gorm:"not null" json:"temperature"`
	Pressure     float64   `gorm:"not null" json:"pressure"`
	CreatedAt    time.Time `json:"created_at"`
}

type CalibratedCTD struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	DiveID      string    `gorm:"index;size:64" json:"dive_id"`
	Timestamp   time.Time `gorm:"index" json:"timestamp"`
	Temperature float64   `json:"temperature"`
	Pressure    float64   `json:"pressure"`
	Salinity    float64   `json:"salinity"`
	Depth       float64   `json:"depth"`
	Density     float64   `json:"density"`
	CreatedAt   time.Time `json:"created_at"`
}

type RawCTDBatchRequest struct {
	DiveID string            `json:"dive_id" binding:"required"`
	Data   []RawCTDDataPoint `json:"data" binding:"required,min=1"`
}

type RawCTDDataPoint struct {
	Timestamp    int64   `json:"timestamp" binding:"required"`
	Conductivity float64 `json:"conductivity" binding:"required"`
	Temperature  float64 `json:"temperature" binding:"required"`
	Pressure     float64 `json:"pressure" binding:"required"`
}

type CalibrateRequest struct {
	DiveID string `json:"dive_id" binding:"required"`
}

type CalibrateResponse struct {
	DiveID       string            `json:"dive_id"`
	Count        int               `json:"count"`
	Data         []CalibratedCTD   `json:"data,omitempty"`
	CalibratedAt time.Time         `json:"calibrated_at"`
}

type Pagination struct {
	Page     int `form:"page,default=1"`
	PageSize int `form:"page_size,default=20"`
}

func (p Pagination) Offset() int {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 20
	}
	return (p.Page - 1) * p.PageSize
}
