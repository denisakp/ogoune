package monitoring

import "github.com/denisakp/ogoune/internal/port"

// Compile-time interface satisfaction checks.
var _ port.MonitoringIncidentProcessor = (*IncidentService)(nil)
