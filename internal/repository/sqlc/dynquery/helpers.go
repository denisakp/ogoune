package dynquery

import "strings"

// likeEscape escapes the LIKE wildcards so user-supplied substrings match
// literally. Paired with `LIKE ? ESCAPE '\'` in the SQL.
func likeEscape(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

// validResourceTypes is the set of accepted ?type= values for monitor filters.
// Kept here (not derived from domain) so the SQL layer has no dependency on
// the domain package and the enum stays auditable next to the filter code.
var validResourceTypes = map[string]struct{}{
	"http":      {},
	"tcp":       {},
	"dns":       {},
	"icmp":      {},
	"keyword":   {},
	"protocol":  {},
	"heartbeat": {},
}

func isValidResourceType(s string) bool {
	_, ok := validResourceTypes[s]
	return ok
}

// validIncidentStatuses is the set of accepted ?status= values.
var validIncidentStatuses = map[string]struct{}{
	"open":     {},
	"resolved": {},
}

func isValidIncidentStatus(s string) bool {
	_, ok := validIncidentStatuses[s]
	return ok
}

const (
	maxQLength = 200
	maxTagLen  = 100
)
