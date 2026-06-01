package dynquery

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	dtoV1 "github.com/denisakp/ogoune/internal/dto/v1"
)

// MonitorFilter holds the parsed ?tag=, ?type=, ?is_active=, ?q= filters for
// the v1 /monitors list endpoint. Pointer fields: nil means "no filter".
type MonitorFilter struct {
	Tag      *string
	Type     *string
	IsActive *bool
	Q        *string
}

// Validate returns field-level errors for invalid filter values. Empty filter
// is valid.
func (f *MonitorFilter) Validate() []dtoV1.FieldError {
	var errs []dtoV1.FieldError
	if f.Tag != nil && len(*f.Tag) > maxTagLen {
		errs = append(errs, dtoV1.FieldError{Field: "tag", Message: fmt.Sprintf("must be %d characters or fewer", maxTagLen)})
	}
	if f.Type != nil && !isValidResourceType(*f.Type) {
		errs = append(errs, dtoV1.FieldError{Field: "type", Message: "must be one of http, tcp, dns, icmp, keyword, protocol, heartbeat"})
	}
	if f.Q != nil && len(*f.Q) > maxQLength {
		errs = append(errs, dtoV1.FieldError{Field: "q", Message: fmt.Sprintf("must be %d characters or fewer", maxQLength)})
	}
	return errs
}

// BuildMonitorsQuery builds the rows + count SQL for the v1 monitors list
// endpoint. Column names, JOINs and operators are hardcoded; user-derived
// values are parameterised via squirrel.
func BuildMonitorsQuery(f MonitorFilter, page, perPage int, ph sq.PlaceholderFormat) (rowsSQL string, rowsArgs []any, countSQL string, countArgs []any, err error) {
	preds := monitorPredicates(f)

	rowsBuilder := sq.Select("r.*").
		From("resources r").
		Where(sq.And(preds)).
		OrderBy("r.created_at DESC").
		Limit(uint64(perPage)).
		Offset(uint64((page - 1) * perPage)).
		PlaceholderFormat(ph)
	if f.Tag != nil {
		rowsBuilder = rowsBuilder.
			Join("resource_tags rt ON rt.resource_id = r.id").
			Join("tags t ON t.id = rt.tag_id")
	}
	rowsSQL, rowsArgs, err = rowsBuilder.ToSql()
	if err != nil {
		return "", nil, "", nil, err
	}

	countBuilder := sq.Select("COUNT(*)").
		From("resources r").
		Where(sq.And(preds)).
		PlaceholderFormat(ph)
	if f.Tag != nil {
		countBuilder = countBuilder.
			Join("resource_tags rt ON rt.resource_id = r.id").
			Join("tags t ON t.id = rt.tag_id")
	}
	countSQL, countArgs, err = countBuilder.ToSql()
	if err != nil {
		return "", nil, "", nil, err
	}
	return rowsSQL, rowsArgs, countSQL, countArgs, nil
}

// BuildMonitorIDsQuery builds an ID-only SELECT for the same filter shape as
// BuildMonitorsQuery. Repo impls use this to get matching IDs, then re-fetch
// full rows + preloads via existing well-tested paths (preserves Tags /
// Component associations that a Raw SELECT * would not populate).
func BuildMonitorIDsQuery(f MonitorFilter, page, perPage int, ph sq.PlaceholderFormat) (string, []any, error) {
	preds := monitorPredicates(f)
	b := sq.Select("r.id").
		From("resources r").
		Where(sq.And(preds)).
		OrderBy("r.created_at DESC").
		Limit(uint64(perPage)).
		Offset(uint64((page - 1) * perPage)).
		PlaceholderFormat(ph)
	if f.Tag != nil {
		b = b.
			Join("resource_tags rt ON rt.resource_id = r.id").
			Join("tags t ON t.id = rt.tag_id")
	}
	return b.ToSql()
}

// BuildMonitorCountQuery builds the COUNT(*) for the same filter shape.
func BuildMonitorCountQuery(f MonitorFilter, ph sq.PlaceholderFormat) (string, []any, error) {
	preds := monitorPredicates(f)
	b := sq.Select("COUNT(DISTINCT r.id)").
		From("resources r").
		Where(sq.And(preds)).
		PlaceholderFormat(ph)
	if f.Tag != nil {
		b = b.
			Join("resource_tags rt ON rt.resource_id = r.id").
			Join("tags t ON t.id = rt.tag_id")
	}
	return b.ToSql()
}

// monitorPredicates returns the WHERE predicates shared by rows + count queries.
// Default is `is_active = TRUE` unless caller explicitly filters on it.
func monitorPredicates(f MonitorFilter) sq.And {
	preds := sq.And{}

	if f.IsActive != nil {
		preds = append(preds, sq.Eq{"r.is_active": *f.IsActive})
	} else {
		preds = append(preds, sq.Eq{"r.is_active": true})
	}
	if f.Type != nil {
		preds = append(preds, sq.Eq{"r.type": *f.Type})
	}
	if f.Tag != nil {
		preds = append(preds, sq.Eq{"t.name": *f.Tag})
	}
	if f.Q != nil {
		like := "%" + likeEscape(*f.Q) + "%"
		preds = append(preds, sq.Or{
			sq.Expr(`LOWER(r.name) LIKE LOWER(?) ESCAPE '\'`, like),
			sq.Expr(`LOWER(r.target) LIKE LOWER(?) ESCAPE '\'`, like),
		})
	}
	return preds
}
