package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/fake"
)

func TestUpdateSettings_BrandingValidation(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(s *domain.StatusPageSettings)
		wantErr error
	}{
		{
			name: "valid primary color and overrides",
			mutate: func(s *domain.StatusPageSettings) {
				s.PrimaryColor = "#4f46e5"
				s.ThemeOverrides = map[string]string{
					"--status-bg":     "#ffffff",
					"--status-up":     "#10b981",
					"--status-radius": "8px",
				}
			},
		},
		{
			name:    "invalid hex color rejected",
			mutate:  func(s *domain.StatusPageSettings) { s.PrimaryColor = "rebeccapurple" },
			wantErr: ErrInvalidHexColor,
		},
		{
			name: "unknown override key rejected",
			mutate: func(s *domain.StatusPageSettings) {
				s.ThemeOverrides = map[string]string{"--evil": "#000000"}
			},
			wantErr: ErrInvalidThemeKey,
		},
		{
			name: "invalid radius rejected",
			mutate: func(s *domain.StatusPageSettings) {
				s.ThemeOverrides = map[string]string{"--status-radius": "huge"}
			},
			wantErr: ErrInvalidThemeValue,
		},
		{
			name: "empty overrides allowed",
			mutate: func(s *domain.StatusPageSettings) {
				s.ThemeOverrides = map[string]string{}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := fake.NewStatusPageSettingsFake()
			svc := NewStatusPageSettingsService(repo)
			s := &domain.StatusPageSettings{Name: "Acme"}
			tc.mutate(s)
			err := svc.UpdateSettings(context.Background(), s)
			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr), "got %v", err)
				return
			}
			require.NoError(t, err)
		})
	}
}
