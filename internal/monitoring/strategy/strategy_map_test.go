package strategy

import (
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
)

func TestStrategyMapCompleteness(t *testing.T) {
	expectedTypes := []domain.ResourceType{
		domain.ResourceHTTP,
		domain.ResourceTCP,
		domain.ResourceDNS,
	}

	strategies := map[domain.ResourceType]domain.CheckStrategy{
		domain.ResourceHTTP: NewHTTPStrategy(30 * time.Second),
		domain.ResourceTCP:  NewTCPStrategy(30 * time.Second),
		domain.ResourceDNS:  NewDNSStrategy(30 * time.Second),
	}

	for _, rt := range expectedTypes {
		if _, ok := strategies[rt]; !ok {
			t.Errorf("strategy map missing ResourceType: %s", rt)
		}
	}

	for rt := range strategies {
		found := false
		for _, expected := range expectedTypes {
			if rt == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("strategy map has unknown ResourceType: %s", rt)
		}
	}
}

func TestNewDNSStrategyPublic(t *testing.T) {
	strategy := NewDNSStrategy(30 * time.Second)
	if strategy == nil {
		t.Fatal("NewDNSStrategy returned nil")
	}
}
