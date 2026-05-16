package v1_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	v1 "github.com/denisakp/ogoune/internal/api/handler/v1"
	"github.com/denisakp/ogoune/internal/domain"
	dtoV1 "github.com/denisakp/ogoune/internal/dto/v1"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- helper: build a Chi router with v1 status-page routes ---

func newStatusPageRouter(repo v1.ComponentV1RepositoryInterface) *chi.Mux {
	r := chi.NewRouter()
	h := v1.NewStatusPageV1Handler(repo)
	r.Get("/api/v1/status-pages", h.List)
	return r
}

func makeComponents(n int, status domain.ComponentStatus) []*domain.Component {
	comps := make([]*domain.Component, n)
	for i := range comps {
		desc := "description " + string(rune('A'+i))
		comps[i] = &domain.Component{
			Base:                   domain.Base{ID: "comp-" + string(rune('0'+i)), CreatedAt: time.Now(), UpdatedAt: time.Now()},
			Name:                   "Component " + string(rune('A'+i)),
			Description:            &desc,
			LastNotificationStatus: status,
		}
	}
	return comps
}

// ============================================================
// Status page handler tests
// ============================================================

func TestStatusPageHandler_List_Success_ReturnsPaginatedEnvelope(t *testing.T) {
	comps := makeComponents(3, domain.ComponentStatusUp)
	repo := &mockComponentRepo{components: comps}
	router := newStatusPageRouter(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/status-pages?page=1&per_page=10", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var out map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &out))

	// Verify envelope shape
	_, hasData := out["data"]
	assert.True(t, hasData, "response should have 'data' key")
	_, hasMeta := out["meta"]
	assert.True(t, hasMeta, "response should have 'meta' key")

	data, ok := out["data"].([]interface{})
	require.True(t, ok, "'data' should be an array")
	assert.Len(t, data, 3)

	meta, ok := out["meta"].(map[string]interface{})
	require.True(t, ok, "'meta' should be an object")
	assert.Equal(t, float64(1), meta["page"])
	assert.Equal(t, float64(10), meta["per_page"])
	assert.Equal(t, float64(3), meta["total"])
}

func TestStatusPageHandler_List_Empty_ReturnsEmptyArray(t *testing.T) {
	repo := &mockComponentRepo{components: []*domain.Component{}}
	router := newStatusPageRouter(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/status-pages", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var out map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &out))

	data, ok := out["data"].([]interface{})
	require.True(t, ok, "'data' should be an array, not null")
	assert.Len(t, data, 0, "empty list should return empty array")
}

func TestStatusPageHandler_List_InvalidPagination_Returns422(t *testing.T) {
	repo := &mockComponentRepo{}
	router := newStatusPageRouter(repo)

	cases := []struct {
		name  string
		query string
	}{
		{"per_page=0", "?per_page=0"},
		{"page=-1", "?page=-1"},
		{"page=abc", "?page=abc"},
		{"per_page=xyz", "?per_page=xyz"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/status-pages"+tc.query, nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
			var out dtoV1.ErrorResponse
			require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &out))
			assert.Equal(t, "VALIDATION_FAILED", out.Error.Code)
			assert.NotEmpty(t, out.Error.Fields, "error.fields should be non-empty")
		})
	}
}

func TestStatusPageHandler_List_RepositoryError_Returns500(t *testing.T) {
	repo := &mockComponentRepo{listErr: errors.New("db connection lost")}
	router := newStatusPageRouter(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/status-pages", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	var out dtoV1.ErrorResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &out))
	assert.Equal(t, "INTERNAL_ERROR", out.Error.Code)
}

func TestStatusPageHandler_List_OverallStatusMapsFromLastNotificationStatus(t *testing.T) {
	cases := []struct {
		name           string
		status         domain.ComponentStatus
		expectedStatus string
	}{
		{"up", domain.ComponentStatusUp, "up"},
		{"degraded", domain.ComponentStatusDegraded, "degraded"},
		{"down", domain.ComponentStatusDown, "down"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			comps := makeComponents(1, tc.status)
			repo := &mockComponentRepo{components: comps}
			router := newStatusPageRouter(repo)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/status-pages", nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			require.Equal(t, http.StatusOK, rr.Code)
			var out map[string]interface{}
			require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &out))

			data := out["data"].([]interface{})
			require.Len(t, data, 1)
			item := data[0].(map[string]interface{})
			assert.Equal(t, tc.expectedStatus, item["overall_status"], "overall_status should map from LastNotificationStatus")
		})
	}
}
