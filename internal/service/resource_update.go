package service

import (
	"context"
	"fmt"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/dto"
)

func (s *ResourceService) applyResourcePayload(ctx context.Context, resource *domain.Resource, payload *dto.UpdateResourcePayload) error {
	applyBasicFields(resource, payload)

	if err := applyTargetField(resource, payload); err != nil {
		return err
	}

	if err := s.applyConfirmationSettings(resource, payload); err != nil {
		return err
	}

	if payload.IsActive != nil {
		resource.IsActive = *payload.IsActive
	}

	if err := s.applyComponentID(ctx, resource, payload); err != nil {
		return err
	}

	if err := s.applyTags(ctx, resource, payload); err != nil {
		return err
	}

	if err := domain.ValidateResourceTarget(resource.Target, resource.Type); err != nil {
		return fmt.Errorf(errWrapFmt, ErrValidationFailed, err)
	}

	applyExpiryThresholds(resource, payload)
	s.applySmartAlertingFields(resource, payload)

	return validateSmartAlertingFields(resource.FlapThreshold, resource.FlapWindowSeconds, resource.FlapMaxDurationMinutes, resource.ReminderIntervalMinutes)
}

func applyBasicFields(resource *domain.Resource, payload *dto.UpdateResourcePayload) {
	if payload.Name != nil {
		resource.Name = *payload.Name
	}
	if payload.Type != nil {
		resource.Type = *payload.Type
	}
	if payload.Interval != nil {
		resource.Interval = *payload.Interval
	}
	if payload.Timeout != nil {
		resource.Timeout = *payload.Timeout
	}
}

func applyTargetField(resource *domain.Resource, payload *dto.UpdateResourcePayload) error {
	if payload.Target == nil {
		return nil
	}
	resType := resource.Type
	if payload.Type != nil {
		resType = *payload.Type
	}
	if err := domain.ValidateResourceTarget(*payload.Target, resType); err != nil {
		return fmt.Errorf(errWrapFmt, ErrValidationFailed, err)
	}
	resource.Target = *payload.Target
	return nil
}

func applyExpiryThresholds(resource *domain.Resource, payload *dto.UpdateResourcePayload) {
	if payload.ExpiryAlertThresholds == nil {
		return
	}
	if *payload.ExpiryAlertThresholds == "" {
		resource.ExpiryAlertThresholds = nil
	} else {
		resource.ExpiryAlertThresholds = payload.ExpiryAlertThresholds
	}
}

func (s *ResourceService) applyConfirmationSettings(resource *domain.Resource, payload *dto.UpdateResourcePayload) error {
	defaultChecks, defaultInterval := confirmationDefaults()
	if resource.ConfirmationChecks < 1 {
		resource.ConfirmationChecks = defaultChecks
	}
	if resource.ConfirmationInterval <= 0 {
		resource.ConfirmationInterval = defaultInterval
	}
	if payload.ConfirmationChecks != nil {
		resource.ConfirmationChecks = *payload.ConfirmationChecks
	}
	if payload.ConfirmationInterval != nil {
		resource.ConfirmationInterval = *payload.ConfirmationInterval
	}
	if err := domain.ValidateConfirmationSettings(resource.Interval, resource.ConfirmationChecks, resource.ConfirmationInterval); err != nil {
		return fmt.Errorf(errWrapFmt, ErrValidationFailed, err)
	}
	return nil
}

func (s *ResourceService) applyComponentID(ctx context.Context, resource *domain.Resource, payload *dto.UpdateResourcePayload) error {
	if payload.ComponentID == nil {
		return nil
	}
	if s.components == nil {
		return fmt.Errorf("%w: component support is not configured", ErrValidationFailed)
	}
	if *payload.ComponentID == "" {
		resource.ComponentID = nil
		return nil
	}
	if _, err := s.components.GetComponent(ctx, *payload.ComponentID); err != nil {
		return fmt.Errorf("%w: invalid component reference", ErrValidationFailed)
	}
	resource.ComponentID = payload.ComponentID
	return nil
}

func (s *ResourceService) applyTags(ctx context.Context, resource *domain.Resource, payload *dto.UpdateResourcePayload) error {
	if payload.Tags == nil {
		return nil
	}
	if len(*payload.Tags) == 0 {
		resource.Tags = []*domain.Tags{}
		return nil
	}
	tags, err := s.findOrCreateTags(ctx, *payload.Tags)
	if err != nil {
		return fmt.Errorf("failed to process tags: %w", err)
	}
	resource.Tags = tags
	return nil
}

func (s *ResourceService) applySmartAlertingFields(resource *domain.Resource, payload *dto.UpdateResourcePayload) {
	if payload.FlapDetectionEnabled != nil {
		resource.FlapDetectionEnabled = *payload.FlapDetectionEnabled
	}
	if payload.FlapThreshold != nil {
		resource.FlapThreshold = *payload.FlapThreshold
	}
	if payload.FlapWindowSeconds != nil {
		resource.FlapWindowSeconds = *payload.FlapWindowSeconds
	}
	if payload.FlapMaxDurationMinutes != nil {
		resource.FlapMaxDurationMinutes = *payload.FlapMaxDurationMinutes
	}
	if payload.ReminderIntervalMinutes != nil {
		resource.ReminderIntervalMinutes = *payload.ReminderIntervalMinutes
	}
}

func (s *ResourceService) validateTypeSpecificUpdate(resource *domain.Resource, payload *dto.UpdateResourcePayload) error {
	switch resource.Type {
	case domain.ResourceHeartbeat:
		return applyAndValidateHeartbeat(resource, payload)
	case domain.ResourceKeyword:
		return applyAndValidateKeyword(resource, payload)
	case domain.ResourceProtocol:
		return applyAndValidateProtocol(resource, payload)
	default:
		return nil
	}
}

func applyAndValidateHeartbeat(resource *domain.Resource, payload *dto.UpdateResourcePayload) error {
	if payload.HeartbeatInterval != nil {
		resource.HeartbeatInterval = payload.HeartbeatInterval
	}
	if payload.HeartbeatGrace != nil {
		resource.HeartbeatGrace = payload.HeartbeatGrace
	}
	if resource.HeartbeatInterval != nil && resource.HeartbeatGrace != nil {
		if err := domain.ValidateHeartbeatSettings(*resource.HeartbeatInterval, *resource.HeartbeatGrace); err != nil {
			return fmt.Errorf(errWrapFmt, ErrValidationFailed, err)
		}
	}
	return nil
}

func applyAndValidateKeyword(resource *domain.Resource, payload *dto.UpdateResourcePayload) error {
	if payload.Keyword != nil {
		resource.Keyword = payload.Keyword
	}
	if payload.KeywordMode != nil {
		resource.KeywordMode = payload.KeywordMode
	}
	return validateKeywordFields(resource.Keyword, resource.KeywordMode)
}

func applyAndValidateProtocol(resource *domain.Resource, payload *dto.UpdateResourcePayload) error {
	if payload.ProtocolType != nil {
		resource.ProtocolType = payload.ProtocolType
	}
	if payload.ProtocolPort != nil {
		resource.ProtocolPort = payload.ProtocolPort
	}
	return validateProtocolFields(resource.ProtocolType, resource.ProtocolPort)
}

func (s *ResourceService) reconcileComponentChange(ctx context.Context, resource *domain.Resource, previousComponentID *string) {
	if resource.ComponentID != nil && s.components != nil {
		_ = s.components.RecalculateAndNotify(ctx, *resource.ComponentID)
	}
	if previousComponentID == nil || s.components == nil {
		return
	}
	if resource.ComponentID != nil && *previousComponentID == *resource.ComponentID {
		return
	}
	count, err := s.resources.CountByComponentID(ctx, *previousComponentID)
	if err != nil {
		return
	}
	if count == 0 {
		_ = s.components.DeleteComponent(ctx, *previousComponentID)
	} else {
		_ = s.components.RecalculateAndNotify(ctx, *previousComponentID)
	}
}
