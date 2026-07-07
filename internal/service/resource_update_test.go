package service

import (
	"context"
	"errors"
	"testing"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceService_UpdateResource_AppliesBasicAndTargetFields(t *testing.T) {
	service, _, _ := newResourceServiceForTest()

	created, err := service.CreateResource(context.Background(), &dto.CreateResourcePayload{
		Name:     "original",
		Type:     domain.ResourceHTTP,
		Target:   "https://example.com",
		Interval: 60,
		Timeout:  5,
	})
	require.NoError(t, err)

	newName := "renamed"
	newTarget := "https://example.org"
	newTimeout := 10

	updated, err := service.UpdateResource(context.Background(), created.ID, &dto.UpdateResourcePayload{
		Name:    &newName,
		Target:  &newTarget,
		Timeout: &newTimeout,
	})
	require.NoError(t, err)
	assert.Equal(t, "renamed", updated.Name)
	assert.Equal(t, "https://example.org", updated.Target)
	assert.Equal(t, 10, updated.Timeout)
}

func TestResourceService_UpdateResource_RejectsInvalidHeartbeatSettings(t *testing.T) {
	service, _, _ := newResourceServiceForTest()

	heartbeatInterval := 300
	heartbeatGrace := 60
	created, err := service.CreateResource(context.Background(), &dto.CreateResourcePayload{
		Name:              "hb",
		Type:              domain.ResourceHeartbeat,
		Interval:          60,
		Timeout:           5,
		HeartbeatInterval: &heartbeatInterval,
		HeartbeatGrace:    &heartbeatGrace,
	})
	require.NoError(t, err)

	badGrace := -1
	_, err = service.UpdateResource(context.Background(), created.ID, &dto.UpdateResourcePayload{
		HeartbeatGrace: &badGrace,
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrValidationFailed), "expected ErrValidationFailed, got %v", err)
}
