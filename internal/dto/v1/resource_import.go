package v1

// DuplicatePolicy controls how the importer treats a manifest resource whose
// name already exists. It defaults to "skip".
type DuplicatePolicy string

const (
	// DuplicatePolicySkip reports an existing name as skipped and imports the rest.
	DuplicatePolicySkip DuplicatePolicy = "skip"
	// DuplicatePolicyError fails the (all-or-nothing) import when a name already exists.
	DuplicatePolicyError DuplicatePolicy = "error"
)

// RowAction is the outcome decided for a single manifest row.
type RowAction string

const (
	RowActionCreate RowAction = "create"
	RowActionSkip   RowAction = "skip"
	RowActionError  RowAction = "error"
)

// RowResult is the per-row outcome of validation / import.
type RowResult struct {
	Index  int       `json:"index"`
	Name   string    `json:"name"`
	Valid  bool      `json:"valid"`
	Action RowAction `json:"action"`
	Errors []string  `json:"errors,omitempty"`
}

// ImportReport is the aggregate outcome of a dry-run or real import.
type ImportReport struct {
	DryRun  bool        `json:"dry_run"`
	Total   int         `json:"total"`
	Created int         `json:"created"`
	Skipped int         `json:"skipped"`
	Failed  int         `json:"failed"`
	Rows    []RowResult `json:"rows"`
}

// ImportOptions carries the request-level knobs for an import.
type ImportOptions struct {
	DryRun          bool
	DuplicatePolicy DuplicatePolicy
}
