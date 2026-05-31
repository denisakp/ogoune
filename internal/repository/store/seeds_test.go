package store_test

import (
	"testing"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
)

// seedResource inserts a minimal Resource row so child tables with a
// resource_id FK can reference it. Returns the resource for further use.
func seedResource(t *testing.T, fx *internaltest.DialectFixture, id, name string) *domain.Resource {
	t.Helper()
	res := &domain.Resource{
		Base:     domain.Base{ID: id},
		Name:     name,
		Type:     domain.ResourceHTTP,
		Target:   "https://" + id + ".invalid",
		Interval: 60,
		Timeout:  10,
		IsActive: true,
	}
	if err := fx.Runtime.GormDB().Create(res).Error; err != nil {
		t.Fatalf("seed resource %q: %v", id, err)
	}
	return res
}

// seedIncident inserts a minimal Incident row referencing the given resource_id.
func seedIncident(t *testing.T, fx *internaltest.DialectFixture, id, resourceID string) *domain.Incident {
	t.Helper()
	inc := &domain.Incident{
		Base:       domain.Base{ID: id},
		ResourceID: resourceID,
		Cause:      "seed",
	}
	if err := fx.Runtime.GormDB().Create(inc).Error; err != nil {
		t.Fatalf("seed incident %q: %v", id, err)
	}
	return inc
}
