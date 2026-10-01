package store

// diffJunctionSets computes the symmetric difference between two ID sets.
// Returns (toAdd, toRemove) where:
//   - toAdd = target \ current  (IDs present in target but not in current)
//   - toRemove = current \ target  (IDs present in current but not in target)
//
// Used by M2M repositories (component, maintenance) to compute the diff
// between currently-persisted junction rows and the desired target set when
// applying an Update. Both DELETE (toRemove) and INSERT (toAdd) MUST happen
// inside the same transaction via pg.WithTx / sqlitesqlc.WithTx.
//
// Order in returned slices is not guaranteed; callers MUST NOT rely on it.
func diffJunctionSets(current, target []string) (toAdd, toRemove []string) {
	currentSet := make(map[string]struct{}, len(current))
	for _, id := range current {
		currentSet[id] = struct{}{}
	}
	targetSet := make(map[string]struct{}, len(target))
	for _, id := range target {
		targetSet[id] = struct{}{}
	}
	for id := range targetSet {
		if _, ok := currentSet[id]; !ok {
			toAdd = append(toAdd, id)
		}
	}
	for id := range currentSet {
		if _, ok := targetSet[id]; !ok {
			toRemove = append(toRemove, id)
		}
	}
	return toAdd, toRemove
}

// linkEach calls link for every id and stops at the first error. The loop
// that every Create used to spell out per dialect.
func linkEach(ids []string, link func(id string) error) error {
	for _, id := range ids {
		if err := link(id); err != nil {
			return err
		}
	}
	return nil
}

// syncJunction brings a junction table from current to target: unlink what
// is gone, then link what is new -- in that order, inside the caller's
// transaction. It is the whole of what an Update does to a junction, and it
// was written out twice per repository (once per dialect) before this.
func syncJunction(current, target []string, unlink, link func(id string) error) error {
	toAdd, toRemove := diffJunctionSets(current, target)
	if err := linkEach(toRemove, unlink); err != nil {
		return err
	}
	return linkEach(toAdd, link)
}
