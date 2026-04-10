package service

import (
	"context"
	"strings"
	"testing"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func strPtr(s string) *string { return &s }

func TestCreateKeywordResource_MissingKeyword(t *testing.T) {
	svc, _, _ := newResourceServiceForTest()
	payload := &dto.CreateResourcePayload{
		Name:     "Keyword Monitor",
		Type:     domain.ResourceKeyword,
		Target:   "https://example.com",
		Interval: 60,
		Timeout:  10,
	}
	_, err := svc.CreateResource(context.Background(), payload)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrValidationFailed)
	assert.Contains(t, err.Error(), "keyword is required")
}

func TestCreateKeywordResource_KeywordTooLong(t *testing.T) {
	svc, _, _ := newResourceServiceForTest()
	longKeyword := strings.Repeat("a", 501)
	payload := &dto.CreateResourcePayload{
		Name:     "Keyword Monitor",
		Type:     domain.ResourceKeyword,
		Target:   "https://example.com",
		Interval: 60,
		Timeout:  10,
		Keyword:  &longKeyword,
	}
	_, err := svc.CreateResource(context.Background(), payload)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrValidationFailed)
	assert.Contains(t, err.Error(), "must not exceed 500")
}

func TestCreateKeywordResource_InvalidKeywordMode(t *testing.T) {
	svc, _, _ := newResourceServiceForTest()
	payload := &dto.CreateResourcePayload{
		Name:        "Keyword Monitor",
		Type:        domain.ResourceKeyword,
		Target:      "https://example.com",
		Interval:    60,
		Timeout:     10,
		Keyword:     strPtr("operational"),
		KeywordMode: strPtr("invalid_mode"),
	}
	_, err := svc.CreateResource(context.Background(), payload)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrValidationFailed)
	assert.Contains(t, err.Error(), "keyword_mode must be")
}

func TestCreateKeywordResource_ValidContainsMode(t *testing.T) {
	svc, _, _ := newResourceServiceForTest()
	payload := &dto.CreateResourcePayload{
		Name:        "Keyword Monitor",
		Type:        domain.ResourceKeyword,
		Target:      "https://example.com",
		Interval:    60,
		Timeout:     10,
		Keyword:     strPtr("operational"),
		KeywordMode: strPtr("contains"),
	}
	resource, err := svc.CreateResource(context.Background(), payload)
	require.NoError(t, err)
	require.NotNil(t, resource)
	assert.Equal(t, "operational", *resource.Keyword)
	assert.Equal(t, "contains", *resource.KeywordMode)
}

func TestCreateKeywordResource_ValidNotContainsMode(t *testing.T) {
	svc, _, _ := newResourceServiceForTest()
	payload := &dto.CreateResourcePayload{
		Name:        "Keyword Monitor",
		Type:        domain.ResourceKeyword,
		Target:      "https://example.com",
		Interval:    60,
		Timeout:     10,
		Keyword:     strPtr("error"),
		KeywordMode: strPtr("not_contains"),
	}
	resource, err := svc.CreateResource(context.Background(), payload)
	require.NoError(t, err)
	require.NotNil(t, resource)
	assert.Equal(t, "error", *resource.Keyword)
	assert.Equal(t, "not_contains", *resource.KeywordMode)
}

func TestCreateKeywordResource_DefaultsToContainsMode(t *testing.T) {
	svc, _, _ := newResourceServiceForTest()
	payload := &dto.CreateResourcePayload{
		Name:     "Keyword Monitor",
		Type:     domain.ResourceKeyword,
		Target:   "https://example.com",
		Interval: 60,
		Timeout:  10,
		Keyword:  strPtr("operational"),
	}
	resource, err := svc.CreateResource(context.Background(), payload)
	require.NoError(t, err)
	require.NotNil(t, resource)
	assert.Equal(t, "contains", *resource.KeywordMode)
}

func TestUpdateKeywordResource_MissingKeyword(t *testing.T) {
	svc, repo, _ := newResourceServiceForTest()

	// First create a valid keyword resource
	existing := &domain.Resource{
		Name:        "Keyword Monitor",
		Type:        domain.ResourceKeyword,
		Target:      "https://example.com",
		Interval:    60,
		Timeout:     10,
		Keyword:     strPtr("operational"),
		KeywordMode: strPtr("contains"),
	}
	created, err := repo.Create(context.Background(), existing)
	require.NoError(t, err)

	// Now update clearing the keyword
	emptyKeyword := ""
	_, err = svc.UpdateResource(context.Background(), created.ID, &dto.UpdateResourcePayload{
		Keyword: &emptyKeyword,
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrValidationFailed)
}

func TestUpdateKeywordResource_InvalidKeywordMode(t *testing.T) {
	svc, repo, _ := newResourceServiceForTest()

	existing := &domain.Resource{
		Name:        "Keyword Monitor",
		Type:        domain.ResourceKeyword,
		Target:      "https://example.com",
		Interval:    60,
		Timeout:     10,
		Keyword:     strPtr("operational"),
		KeywordMode: strPtr("contains"),
	}
	created, err := repo.Create(context.Background(), existing)
	require.NoError(t, err)

	_, err = svc.UpdateResource(context.Background(), created.ID, &dto.UpdateResourcePayload{
		KeywordMode: strPtr("bad_mode"),
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrValidationFailed)
}
