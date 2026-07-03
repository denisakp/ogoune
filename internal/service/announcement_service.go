package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	"github.com/denisakp/ogoune/internal/repository"
)

// Announcement service domain errors.
var (
	ErrAnnouncementNotFound   = errors.New("announcement: not found")
	ErrAnnouncementValidation = errors.New("announcement: validation failed")
)

// AnnouncementService manages instance-wide operator banners.
type AnnouncementService struct {
	repo port.AnnouncementRepository
}

func NewAnnouncementService(repo port.AnnouncementRepository) *AnnouncementService {
	return &AnnouncementService{repo: repo}
}

// ListActive returns active announcements, newest first.
func (s *AnnouncementService) ListActive(ctx context.Context) ([]*domain.Announcement, error) {
	return s.repo.ListActive(ctx)
}

// Create validates and persists a new active announcement.
func (s *AnnouncementService) Create(ctx context.Context, in *domain.Announcement) (*domain.Announcement, error) {
	in.Title = strings.TrimSpace(in.Title)
	in.Description = strings.TrimSpace(in.Description)
	if in.Title == "" {
		return nil, fmt.Errorf("%w: title is required", ErrAnnouncementValidation)
	}
	if !in.Severity.IsValid() {
		return nil, fmt.Errorf("%w: severity must be one of info|warning|success|error", ErrAnnouncementValidation)
	}
	in.Active = true
	return s.repo.Create(ctx, in)
}

// Delete removes an announcement; missing → ErrAnnouncementNotFound.
func (s *AnnouncementService) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrAnnouncementNotFound
		}
		return err
	}
	return nil
}
