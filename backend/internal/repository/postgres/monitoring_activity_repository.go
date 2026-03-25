package postgres

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/repository"
	"gorm.io/gorm"
)

// MonitoringActivityRepositoryImpl implements the MonitoringActivityRepository interface using GORM.
type MonitoringActivityRepositoryImpl struct {
	db *gorm.DB
}

// NewMonitoringActivityRepository creates a new MonitoringActivityRepository using GORM.
func NewMonitoringActivityRepository(db *gorm.DB) repository.MonitoringActivityRepository {
	return &MonitoringActivityRepositoryImpl{db: db}
}

// Create persists a new monitoring activity record to the database.
func (r *MonitoringActivityRepositoryImpl) Create(ctx context.Context, activity *domain.MonitoringActivity) error {
	if activity == nil {
		return repository.ErrInvalidInput
	}
	return r.db.WithContext(ctx).Create(activity).Error
}

// List retrieves all monitoring activities with pagination, ordered by creation time descending.
func (r *MonitoringActivityRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*domain.MonitoringActivity, error) {
	var activities []*domain.MonitoringActivity
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Preload("Resource").
		Find(&activities).Error

	if err != nil {
		return nil, fmt.Errorf("failed to list monitoring activities: %w", err)
	}
	return activities, nil
}

// FindByResourceID retrieves all monitoring activities for a specific resource with pagination.
func (r *MonitoringActivityRepositoryImpl) FindByResourceID(ctx context.Context, resourceID string, limit, offset int) ([]*domain.MonitoringActivity, error) {
	if resourceID == "" {
		return nil, repository.ErrInvalidInput
	}

	var activities []*domain.MonitoringActivity
	err := r.db.WithContext(ctx).
		Where("resource_id = ?", resourceID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Preload("Resource").
		Find(&activities).Error

	if err != nil {
		return nil, fmt.Errorf("failed to find monitoring activities for resource %s: %w", resourceID, err)
	}
	return activities, nil
}

// CountTransitionsInWindow counts success/failure flips within the provided time window.
func (r *MonitoringActivityRepositoryImpl) CountTransitionsInWindow(ctx context.Context, resourceID string, windowStart time.Time) (int, error) {
	if resourceID == "" {
		return 0, repository.ErrInvalidInput
	}

	type activityPoint struct {
		Success bool
	}

	var points []activityPoint
	err := r.db.WithContext(ctx).
		Model(&domain.MonitoringActivity{}).
		Select("success").
		Where("resource_id = ? AND created_at >= ?", resourceID, windowStart).
		Order("created_at ASC").
		Find(&points).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count transitions for resource %s: %w", resourceID, err)
	}

	transitions := 0
	for index := 1; index < len(points); index++ {
		if points[index].Success != points[index-1].Success {
			transitions++
		}
	}

	return transitions, nil
}

// GetUptimeStats retrieves the hourly uptime percentage for a resource over the last 24 hours.
func (r *MonitoringActivityRepositoryImpl) GetUptimeStats(ctx context.Context, resourceID string) ([]domain.UptimeStat, error) {
	if resourceID == "" {
		return nil, repository.ErrInvalidInput
	}

	type activityPoint struct {
		CreatedAt time.Time
		Success   bool
	}

	var points []activityPoint
	since := time.Now().Add(-24 * time.Hour)
	err := r.db.WithContext(ctx).
		Model(&domain.MonitoringActivity{}).
		Select("created_at, success").
		Where("resource_id = ? AND created_at >= ?", resourceID, since).
		Order("created_at ASC").
		Find(&points).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get uptime stats for resource %s: %w", resourceID, err)
	}

	type aggregate struct {
		success int
		total   int
	}

	byHour := make(map[time.Time]aggregate)
	for _, p := range points {
		hour := p.CreatedAt.Truncate(time.Hour)
		current := byHour[hour]
		if p.Success {
			current.success++
		}
		current.total++
		byHour[hour] = current
	}

	hours := make([]time.Time, 0, len(byHour))
	for h := range byHour {
		hours = append(hours, h)
	}
	sort.Slice(hours, func(i, j int) bool { return hours[i].Before(hours[j]) })

	stats := make([]domain.UptimeStat, 0, len(hours))
	for _, h := range hours {
		v := byHour[h]
		uptime := 0.0
		if v.total > 0 {
			uptime = math.Round((float64(v.success)/float64(v.total)*100)*100) / 100
		}
		stats = append(stats, domain.UptimeStat{
			Hour:            h,
			UptimePercent:   uptime,
			SuccessfulCount: v.success,
			TotalCount:      v.total,
		})
	}

	return stats, nil
}

// GetRecentResponseTimes retrieves the most recent response times for a resource.
func (r *MonitoringActivityRepositoryImpl) GetRecentResponseTimes(ctx context.Context, resourceID string, limit int) ([]domain.ResponseTimePoint, error) {
	if resourceID == "" {
		return nil, repository.ErrInvalidInput
	}

	if limit <= 0 {
		limit = 50 // default limit
	}

	var responsePoints []domain.ResponseTimePoint

	query := `
		SELECT
			created_at as timestamp,
			response_time
		FROM monitoring_activities
		WHERE resource_id = ?
			AND success = true
		ORDER BY created_at DESC
		LIMIT ?
	`

	err := r.db.WithContext(ctx).Raw(query, resourceID, limit).Scan(&responsePoints).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get recent response times for resource %s: %w", resourceID, err)
	}

	// Reverse the slice to have chronological order (oldest to newest)
	for i, j := 0, len(responsePoints)-1; i < j; i, j = i+1, j-1 {
		responsePoints[i], responsePoints[j] = responsePoints[j], responsePoints[i]
	}

	return responsePoints, nil
}

// GetGlobalUptimeStats calculates the overall uptime percentage across all resources
// for a given time range in hours.
func (r *MonitoringActivityRepositoryImpl) GetGlobalUptimeStats(ctx context.Context, hours int) (float64, error) {
	since := time.Now().Add(-time.Duration(hours) * time.Hour)

	var total int64
	err := r.db.WithContext(ctx).
		Model(&domain.MonitoringActivity{}).
		Where("created_at >= ?", since).
		Count(&total).Error
	if err != nil {
		return 0.0, fmt.Errorf("failed to get global uptime stats: %w", err)
	}

	if total == 0 {
		return 0.0, nil
	}

	var successful int64
	err = r.db.WithContext(ctx).
		Model(&domain.MonitoringActivity{}).
		Where("created_at >= ? AND success = ?", since, true).
		Count(&successful).Error
	if err != nil {
		return 0.0, fmt.Errorf("failed to get global uptime stats: %w", err)
	}

	uptime := math.Round((float64(successful)/float64(total)*100)*100) / 100
	return uptime, nil
}
