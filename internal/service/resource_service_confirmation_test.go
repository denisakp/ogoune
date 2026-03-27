package service

import (
	"context"
	"testing"

	"github.com/denisakp/pulseguard/internal/domain"
	"github.com/denisakp/pulseguard/internal/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceServiceConfirmation_DefaultsAppliedOnCreate(t *testing.T) {
	t.Setenv("CONFIRMATION_CHECKS", "4")
	t.Setenv("CONFIRMATION_INTERVAL", "20")

	svc, repo, _ := newResourceServiceForTest()
	created, err := svc.CreateResource(context.Background(), &dto.CreateResourcePayload{
		Name:     "c-defaults",
		Type:     domain.ResourceHTTP,
		Target:   "https://example.com",
		Interval: 60,
		Timeout:  5,
	})
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, 4, created.ConfirmationChecks)
	assert.Equal(t, 20, created.ConfirmationInterval)

	stored, err := repo.FindByID(context.Background(), created.ID)
	require.NoError(t, err)
	assert.Equal(t, 4, stored.ConfirmationChecks)
	assert.Equal(t, 20, stored.ConfirmationInterval)
}

func TestResourceServiceConfirmation_PerResourceOverrideValues(t *testing.T) {
	svc, _, _ := newResourceServiceForTest()
	checks := 4
	interval := 15
	created, err := svc.CreateResource(context.Background(), &dto.CreateResourcePayload{
		Name:                 "c-override",
		Type:                 domain.ResourceHTTP,
		Target:               "https://example.com",
		Interval:             60,
		Timeout:              5,
		ConfirmationChecks:   &checks,
		ConfirmationInterval: &interval,
	})
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, 4, created.ConfirmationChecks)
	assert.Equal(t, 15, created.ConfirmationInterval)
}

func TestResourceServiceConfirmation_ImmediateModeChecksEqualsOne(t *testing.T) {
	svc, _, _ := newResourceServiceForTest()
	checks := 1
	interval := 10
	created, err := svc.CreateResource(context.Background(), &dto.CreateResourcePayload{
		Name:                 "c-immediate",
		Type:                 domain.ResourceHTTP,
		Target:               "https://example.com",
		Interval:             60,
		Timeout:              5,
		ConfirmationChecks:   &checks,
		ConfirmationInterval: &interval,
	})
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, 1, created.ConfirmationChecks)
}

func TestResourceServiceConfirmation_PreserveExistingOnUpdateWhenOmitted(t *testing.T) {
	svc, repo, _ := newResourceServiceForTest()
	checks := 3
	confirmInterval := 20
	created, err := svc.CreateResource(context.Background(), &dto.CreateResourcePayload{
		Name:                 "before-update",
		Type:                 domain.ResourceHTTP,
		Target:               "https://example.com",
		Interval:             60,
		Timeout:              5,
		ConfirmationChecks:   &checks,
		ConfirmationInterval: &confirmInterval,
	})
	require.NoError(t, err)

	newName := "after-update"
	updated, err := svc.UpdateResource(context.Background(), created.ID, &dto.UpdateResourcePayload{
		Name: &newName,
	})
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, 3, updated.ConfirmationChecks)
	assert.Equal(t, 20, updated.ConfirmationInterval)

	stored, err := repo.FindByID(context.Background(), created.ID)
	require.NoError(t, err)
	assert.Equal(t, 3, stored.ConfirmationChecks)
	assert.Equal(t, 20, stored.ConfirmationInterval)
}
