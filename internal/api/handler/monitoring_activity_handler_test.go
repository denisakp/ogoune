package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/fake"
	"github.com/denisakp/ogoune/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMonitoringActivityHandler_ListActivities_ResponseDataIsReadableText(t *testing.T) {
	repo := fake.NewMonitoringActivityFake()
	service := service.NewMonitoringActivityService(repo)
	h := NewMonitoringActivityHandler(service)

	err := repo.Create(nil, &domain.MonitoringActivity{
		Base:         domain.Base{ID: "act-1", CreatedAt: time.Now()},
		ResourceID:   "res-1",
		Message:      "failed",
		Success:      false,
		ResponseData: []byte("dial tcp timeout"),
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/monitoring-activities", nil)
	rr := httptest.NewRecorder()

	h.ListActivities(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var out map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &out))
	activities, ok := out["activities"].([]interface{})
	require.True(t, ok)
	require.Len(t, activities, 1)

	item, ok := activities[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "dial tcp timeout", item["response_data"])
}
