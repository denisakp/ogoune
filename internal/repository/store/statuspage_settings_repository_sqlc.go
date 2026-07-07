package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	domain "github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/port"
	pgsqlc "github.com/denisakp/ogoune/internal/repository/sqlc/pg"
	sqlitesqlc "github.com/denisakp/ogoune/internal/repository/sqlc/sqlite"
)

func encodeDNSRecords(records []domain.DNSRecord) []byte {
	if records == nil {
		records = []domain.DNSRecord{}
	}
	b, err := json.Marshal(records)
	if err != nil {
		return []byte("[]")
	}
	return b
}

func decodeDNSRecords(raw []byte) []domain.DNSRecord {
	if len(raw) == 0 {
		return nil
	}
	var out []domain.DNSRecord
	_ = json.Unmarshal(raw, &out)
	return out
}

func encodeDNSRecordsString(records []domain.DNSRecord) string {
	return string(encodeDNSRecords(records))
}

func encodeThemeOverrides(m map[string]string) []byte {
	if m == nil {
		m = map[string]string{}
	}
	b, err := json.Marshal(m)
	if err != nil {
		return []byte("{}")
	}
	return b
}

func encodeThemeOverridesString(m map[string]string) string {
	return string(encodeThemeOverrides(m))
}

func decodeThemeOverrides(raw []byte) map[string]string {
	if len(raw) == 0 {
		return map[string]string{}
	}
	out := map[string]string{}
	_ = json.Unmarshal(raw, &out)
	if out == nil {
		return map[string]string{}
	}
	return out
}

func defaultPrimaryColor(c string) string {
	if c == "" {
		return "#4f46e5"
	}
	return c
}

func defaultDomainStatus(s string) domain.DomainStatus {
	if s == "" {
		return domain.DomainStatusPending
	}
	return domain.DomainStatus(s)
}

func defaultDomainSSL(s string) domain.DomainSSLStatus {
	if s == "" {
		return domain.DomainSSLStatusNone
	}
	return domain.DomainSSLStatus(s)
}

type StatusPageSettingsRepositorySQLC struct {
	pgQ     *pgsqlc.Queries
	sqliteQ *sqlitesqlc.Queries
}

func NewStatusPageSettingsRepositorySQLC(rt SqlcRuntime) port.StatusPageSettingsRepository {
	r := &StatusPageSettingsRepositorySQLC{}
	if pool := rt.PgxPool(); pool != nil {
		r.pgQ = pgsqlc.New(pool)
	} else if db := rt.SQLiteDB(); db != nil {
		r.sqliteQ = sqlitesqlc.New(db)
	}
	return r
}

func (r *StatusPageSettingsRepositorySQLC) unconfigured() error {
	return fmt.Errorf("statuspage_settings_sqlc: unconfigured runtime")
}

func defaultStatusPageSettings() *domain.StatusPageSettings {
	return &domain.StatusPageSettings{
		Name:                 "Status Page",
		EnableDetailsPage:    true,
		ShowUptimePercentage: true,
		HidePausedMonitors:   true,
		ShowIncidentHistory:  true,
	}
}

func (r *StatusPageSettingsRepositorySQLC) Get(ctx context.Context) (*domain.StatusPageSettings, error) {
	switch {
	case r.pgQ != nil:
		row, err := r.pgQ.GetStatusPageSettings(ctx)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return defaultStatusPageSettings(), nil
			}
			return nil, fmt.Errorf("sqlc: get status page settings: %w", err)
		}
		return statusPageSettingsFromPG(row), nil
	case r.sqliteQ != nil:
		row, err := r.sqliteQ.GetStatusPageSettings(ctx)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return defaultStatusPageSettings(), nil
			}
			return nil, fmt.Errorf("sqlc: get status page settings: %w", err)
		}
		return statusPageSettingsFromSQLite(row), nil
	default:
		return nil, r.unconfigured()
	}
}

func (r *StatusPageSettingsRepositorySQLC) Upsert(ctx context.Context, s *domain.StatusPageSettings) error {
	// Mirror GORM behavior: read existing, create if absent (preserve ID + CreatedAt
	// on update), else update.
	now := time.Now()
	switch {
	case r.pgQ != nil:
		existing, err := r.pgQ.GetStatusPageSettings(ctx)
		if errors.Is(err, pgx.ErrNoRows) {
			s.EnsureID()
			if s.CreatedAt.IsZero() {
				s.CreatedAt = now
			}
			s.UpdatedAt = now
			return r.pgQ.CreateStatusPageSettings(ctx, pgsqlc.CreateStatusPageSettingsParams{
				ID:                     s.ID,
				Name:                   s.Name,
				HomepageUrl:            s.HomepageURL,
				CustomDomain:           s.CustomDomain,
				UmamiWebsiteID:         s.UmamiWebsiteID,
				UmamiScriptUrl:         s.UmamiScriptURL,
				EnableDetailsPage:      s.EnableDetailsPage,
				ShowUptimePercentage:   s.ShowUptimePercentage,
				HidePausedMonitors:     s.HidePausedMonitors,
				ShowIncidentHistory:    s.ShowIncidentHistory,
				CustomDomainStatus:     string(defaultDomainStatus(string(s.CustomDomainStatus))),
				CustomDomainSslStatus:  string(defaultDomainSSL(string(s.CustomDomainSSL))),
				CustomDomainDnsRecords: encodeDNSRecords(s.CustomDomainDNS),
				LogoUrlLight:           s.LogoURLLight,
				LogoUrlDark:            s.LogoURLDark,
				FaviconUrl:             s.FaviconURL,
				PrimaryColor:           defaultPrimaryColor(s.PrimaryColor),
				ThemeOverrides:         encodeThemeOverrides(s.ThemeOverrides),
				CreatedAt:              pgtype.Timestamptz{Time: s.CreatedAt, Valid: true},
				UpdatedAt:              pgtype.Timestamptz{Time: s.UpdatedAt, Valid: true},
			})
		}
		if err != nil {
			return fmt.Errorf("sqlc: upsert status page settings (lookup): %w", err)
		}
		s.ID = existing.ID
		s.CreatedAt = existing.CreatedAt.Time
		s.UpdatedAt = now
		return r.pgQ.UpdateStatusPageSettings(ctx, pgsqlc.UpdateStatusPageSettingsParams{
			ID:                     s.ID,
			Name:                   s.Name,
			HomepageUrl:            s.HomepageURL,
			CustomDomain:           s.CustomDomain,
			UmamiWebsiteID:         s.UmamiWebsiteID,
				UmamiScriptUrl:         s.UmamiScriptURL,
			EnableDetailsPage:      s.EnableDetailsPage,
			ShowUptimePercentage:   s.ShowUptimePercentage,
			HidePausedMonitors:     s.HidePausedMonitors,
			ShowIncidentHistory:    s.ShowIncidentHistory,
			CustomDomainStatus:     string(defaultDomainStatus(string(s.CustomDomainStatus))),
			CustomDomainSslStatus:  string(defaultDomainSSL(string(s.CustomDomainSSL))),
			CustomDomainDnsRecords: encodeDNSRecords(s.CustomDomainDNS),
			LogoUrlLight:           s.LogoURLLight,
			LogoUrlDark:            s.LogoURLDark,
			FaviconUrl:             s.FaviconURL,
			PrimaryColor:           defaultPrimaryColor(s.PrimaryColor),
			ThemeOverrides:         encodeThemeOverrides(s.ThemeOverrides),
			UpdatedAt:              pgtype.Timestamptz{Time: s.UpdatedAt, Valid: true},
		})
	case r.sqliteQ != nil:
		existing, err := r.sqliteQ.GetStatusPageSettings(ctx)
		if errors.Is(err, sql.ErrNoRows) {
			s.EnsureID()
			if s.CreatedAt.IsZero() {
				s.CreatedAt = now
			}
			s.UpdatedAt = now
			return r.sqliteQ.CreateStatusPageSettings(ctx, sqlitesqlc.CreateStatusPageSettingsParams{
				ID:                     s.ID,
				Name:                   s.Name,
				HomepageUrl:            s.HomepageURL,
				CustomDomain:           s.CustomDomain,
				UmamiWebsiteID:         s.UmamiWebsiteID,
				UmamiScriptUrl:         s.UmamiScriptURL,
				EnableDetailsPage:      boolToInt64(s.EnableDetailsPage),
				ShowUptimePercentage:   boolToInt64(s.ShowUptimePercentage),
				HidePausedMonitors:     boolToInt64(s.HidePausedMonitors),
				ShowIncidentHistory:    boolToInt64(s.ShowIncidentHistory),
				CustomDomainStatus:     string(defaultDomainStatus(string(s.CustomDomainStatus))),
				CustomDomainSslStatus:  string(defaultDomainSSL(string(s.CustomDomainSSL))),
				CustomDomainDnsRecords: encodeDNSRecordsString(s.CustomDomainDNS),
				LogoUrlLight:           s.LogoURLLight,
				LogoUrlDark:            s.LogoURLDark,
				FaviconUrl:             s.FaviconURL,
				PrimaryColor:           defaultPrimaryColor(s.PrimaryColor),
				ThemeOverrides:         encodeThemeOverridesString(s.ThemeOverrides),
				CreatedAt:              s.CreatedAt,
				UpdatedAt:              s.UpdatedAt,
			})
		}
		if err != nil {
			return fmt.Errorf("sqlc: upsert status page settings (lookup): %w", err)
		}
		s.ID = existing.ID
		s.CreatedAt = existing.CreatedAt
		s.UpdatedAt = now
		return r.sqliteQ.UpdateStatusPageSettings(ctx, sqlitesqlc.UpdateStatusPageSettingsParams{
			ID:                     s.ID,
			Name:                   s.Name,
			HomepageUrl:            s.HomepageURL,
			CustomDomain:           s.CustomDomain,
			UmamiWebsiteID:         s.UmamiWebsiteID,
				UmamiScriptUrl:         s.UmamiScriptURL,
			EnableDetailsPage:      boolToInt64(s.EnableDetailsPage),
			ShowUptimePercentage:   boolToInt64(s.ShowUptimePercentage),
			HidePausedMonitors:     boolToInt64(s.HidePausedMonitors),
			ShowIncidentHistory:    boolToInt64(s.ShowIncidentHistory),
			CustomDomainStatus:     string(defaultDomainStatus(string(s.CustomDomainStatus))),
			CustomDomainSslStatus:  string(defaultDomainSSL(string(s.CustomDomainSSL))),
			CustomDomainDnsRecords: encodeDNSRecordsString(s.CustomDomainDNS),
			LogoUrlLight:           s.LogoURLLight,
			LogoUrlDark:            s.LogoURLDark,
			FaviconUrl:             s.FaviconURL,
			PrimaryColor:           defaultPrimaryColor(s.PrimaryColor),
			ThemeOverrides:         encodeThemeOverridesString(s.ThemeOverrides),
			UpdatedAt:              s.UpdatedAt,
		})
	default:
		return r.unconfigured()
	}
}

func statusPageSettingsFromPG(row pgsqlc.StatusPageSetting) *domain.StatusPageSettings {
	return &domain.StatusPageSettings{
		Base: domain.Base{
			ID:        row.ID,
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
		},
		Name:                 row.Name,
		HomepageURL:          row.HomepageUrl,
		CustomDomain:         row.CustomDomain,
		UmamiWebsiteID:       row.UmamiWebsiteID,
		UmamiScriptURL:       row.UmamiScriptUrl,
		EnableDetailsPage:    row.EnableDetailsPage,
		ShowUptimePercentage: row.ShowUptimePercentage,
		HidePausedMonitors:   row.HidePausedMonitors,
		ShowIncidentHistory:  row.ShowIncidentHistory,
		CustomDomainStatus:   defaultDomainStatus(row.CustomDomainStatus),
		CustomDomainSSL:      defaultDomainSSL(row.CustomDomainSslStatus),
		CustomDomainDNS:      decodeDNSRecords(row.CustomDomainDnsRecords),
		LogoURLLight:         row.LogoUrlLight,
		LogoURLDark:          row.LogoUrlDark,
		FaviconURL:           row.FaviconUrl,
		PrimaryColor:         defaultPrimaryColor(row.PrimaryColor),
		ThemeOverrides:       decodeThemeOverrides(row.ThemeOverrides),
	}
}

func statusPageSettingsFromSQLite(row sqlitesqlc.StatusPageSetting) *domain.StatusPageSettings {
	return &domain.StatusPageSettings{
		Base: domain.Base{
			ID:        row.ID,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		},
		Name:                 row.Name,
		HomepageURL:          row.HomepageUrl,
		CustomDomain:         row.CustomDomain,
		UmamiWebsiteID:       row.UmamiWebsiteID,
		UmamiScriptURL:       row.UmamiScriptUrl,
		EnableDetailsPage:    row.EnableDetailsPage != 0,
		ShowUptimePercentage: row.ShowUptimePercentage != 0,
		HidePausedMonitors:   row.HidePausedMonitors != 0,
		ShowIncidentHistory:  row.ShowIncidentHistory != 0,
		CustomDomainStatus:   defaultDomainStatus(row.CustomDomainStatus),
		CustomDomainSSL:      defaultDomainSSL(row.CustomDomainSslStatus),
		CustomDomainDNS:      decodeDNSRecords([]byte(row.CustomDomainDnsRecords)),
		LogoURLLight:         row.LogoUrlLight,
		LogoURLDark:          row.LogoUrlDark,
		FaviconURL:           row.FaviconUrl,
		PrimaryColor:         defaultPrimaryColor(row.PrimaryColor),
		ThemeOverrides:       decodeThemeOverrides([]byte(row.ThemeOverrides)),
	}
}
