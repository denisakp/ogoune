//go:build tools

// Package tools pins build-time tooling (not compiled into the binary) so the
// OpenAPI generator is reproducible from go.mod/go.sum without a global install.
package tools

import (
	_ "github.com/swaggo/swag/v2/cmd/swag"
)
