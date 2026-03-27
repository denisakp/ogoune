package service

import (
	"context"
	"testing"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ptr(s string) *string { return &s }

// TestCreateResource_ExpiryThresholdsStored verifies that expiry_alert_thresholds
// provided in the payload is persisted on the created resource.
func TestCreateResource_ExpiryThresholdsStored(t *testing.T) {
	svc, resourceRepo, _ := newResourceServiceForTest()

	thresholds := "30,14,7"
	payload := &dto.CreateResourcePayload{
		Name:                  "SSL Monitor",
		Type:                  domain.ResourceHTTP,
		Target:                "https://example.com",
		Interval:              60,
		Timeout:               5,
		Tags:                  []string{},
		ExpiryAlertThresholds: &thresholds,
	}

	resource, err := svc.CreateResource(context.Background(), payload)
	require.NoError(t, err)
	require.NotNil(t, resource)
	require.NotNil(t, resource.ExpiryAlertThresholds)
	assert.Equal(t, "30,14,7", *resource.ExpiryAlertThresholds)

	// Verify persistence via repo
	resources, err := resourceRepo.List(context.Background(), 10, 0)
	require.NoError(t, err)
	require.Len(t, resources, 1)
	require.NotNil(t, resources[0].ExpiryAlertThresholds)
	assert.Equal(t, "30,14,7", *resources[0].ExpiryAlertThresholds)
}

// TestUpdateResource_ThresholdPropagated verifies that a provided threshold is
// set on the resource after an update.
func TestUpdateResource_ThresholdPropagated(t *testing.T) {
	svc, resourceRepo, _ := newResourceServiceForTest()

	// Seed a resource first
	base := &dto.CreateResourcePayload{
		Name: "Monitor", Type: domain.ResourceHTTP,
		Target: "https://example.com", Interval: 60, Timeout: 5, Tags: []string{},
	}
	created, err := svc.CreateResource(context.Background(), base)
	require.NoError(t, err)

	thresholds := "60,30,7"
	_, err = svc.UpdateResource(context.Background(), created.ID, &dto.UpdateResourcePayload{
		ExpiryAlertThresholds: &thresholds,
	})
	require.NoError(t, err)

	resources, err := resourceRepo.List(context.Background(), 10, 0)
	require.NoError(t, err)
	require.Len(t, resources, 1)
	require.NotNil(t, resources[0].ExpiryAlertThresholds)
	assert.Equal(t, "60,30,7", *resources[0].ExpiryAlertThresholds)
}

// TestUpdateResource_ThresholdCleared verifies that setting threshold to an
// empty string clears the stored value (sets it to nil).
func TestUpdateResource_ThresholdCleared(t *testing.T) {
	svc, resourceRepo, _ := newResourceServiceForTest()

	// Seed with an existing threshold
	initial := "30,14,7"
	base := &dto.CreateResourcePayload{
		Name: "Monitor", Type: domain.ResourceHTTP,
		Target: "https://example.com", Interval: 60, Timeout: 5, Tags: []string{},
		ExpiryAlertThresholds: &initial,
	}
	created, err := svc.CreateResource(context.Background(), base)
	require.NoError(t, err)
	require.NotNil(t, created.ExpiryAlertThresholds)

	// Clear thresholds by sending empty string
	empty := ""
	_, err = svc.UpdateResource(context.Background(), created.ID, &dto.UpdateResourcePayload{
		ExpiryAlertThresholds: &empty,
	})
	require.NoError(t, err)

	resources, err := resourceRepo.List(context.Background(), 10, 0)
	require.NoError(t, err)
	require.Len(t, resources, 1)
	assert.Nil(t, resources[0].ExpiryAlertThresholds, "empty string should clear the threshold")
}
