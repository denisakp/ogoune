package internaltest_test

import (
	"sync"
	"testing"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/internaltest"
)

func TestSetupSQLite_AppliesMigrationsAndIsolates(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(2)

	t.Run("parallel_a", func(t *testing.T) {
		defer wg.Done()
		t.Parallel()
		fx := internaltest.SetupSQLite(t)
		if fx.Dialect != "sqlite" {
			t.Fatalf("expected dialect sqlite, got %q", fx.Dialect)
		}
		tag := &domain.Tags{Name: "from-a"}
		tag.EnsureID()
		if err := fx.Runtime.GormDB().Create(tag).Error; err != nil {
			t.Fatalf("create in fixture A: %v", err)
		}
		var count int64
		if err := fx.Runtime.GormDB().Model(&domain.Tags{}).Count(&count).Error; err != nil {
			t.Fatalf("count A: %v", err)
		}
		if count != 1 {
			t.Errorf("expected fixture A to see exactly its own row, got count=%d", count)
		}
	})

	t.Run("parallel_b", func(t *testing.T) {
		defer wg.Done()
		t.Parallel()
		fx := internaltest.SetupSQLite(t)
		tag := &domain.Tags{Name: "from-b"}
		tag.EnsureID()
		if err := fx.Runtime.GormDB().Create(tag).Error; err != nil {
			t.Fatalf("create in fixture B: %v", err)
		}
		var count int64
		if err := fx.Runtime.GormDB().Model(&domain.Tags{}).Count(&count).Error; err != nil {
			t.Fatalf("count B: %v", err)
		}
		if count != 1 {
			t.Errorf("expected fixture B to see exactly its own row, got count=%d", count)
		}
	})
}
