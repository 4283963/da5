package service

import (
	"fmt"
	"sync"
	"time"
)

type CircuitBreakerStatus struct {
	Tripped               bool       `json:"tripped"`
	LowConductivityStreak int        `json:"low_conductivity_streak"`
	Threshold             float64    `json:"threshold"`
	StreakLimit           int        `json:"streak_limit"`
	TrippedAt             *time.Time `json:"tripped_at,omitempty"`
	Reason                string     `json:"reason,omitempty"`
}

type CircuitBreaker struct {
	mu        sync.Mutex
	tripped   bool
	streak    int
	threshold float64
	limit     int
	trippedAt time.Time
	reason    string
}

func NewCircuitBreaker(threshold float64, limit int) *CircuitBreaker {
	return &CircuitBreaker{
		threshold: threshold,
		limit:     limit,
	}
}

func (cb *CircuitBreaker) IsTripped() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.tripped
}

func (cb *CircuitBreaker) Observe(conductivity float64) bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.tripped {
		return true
	}

	if conductivity < cb.threshold {
		cb.streak++
	} else {
		cb.streak = 0
	}

	if cb.streak >= cb.limit {
		cb.tripped = true
		cb.trippedAt = time.Now()
		cb.reason = fmt.Sprintf(
			"连续 %d 条海水数据电导率低于阈值 %.4f，疑似传感器镜头被污物覆盖或老化",
			cb.streak, cb.threshold,
		)
		fmt.Println("设备罢工了，快去擦洗传感器")
		return true
	}

	return false
}

func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.tripped = false
	cb.streak = 0
	cb.reason = ""
	cb.trippedAt = time.Time{}
}

func (cb *CircuitBreaker) Status() CircuitBreakerStatus {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	status := CircuitBreakerStatus{
		Tripped:               cb.tripped,
		LowConductivityStreak: cb.streak,
		Threshold:             cb.threshold,
		StreakLimit:           cb.limit,
		Reason:                cb.reason,
	}

	if cb.tripped {
		t := cb.trippedAt
		status.TrippedAt = &t
	}

	return status
}
