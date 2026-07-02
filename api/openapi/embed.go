// Package openapi embeds the generated OpenAPI 3.1 contract so the runtime can
// serve it without depending on a file on disk (no 503 in containers).
// The contract is generated from Go annotations by `make openapi`; never edit
// v1.json/v1.yaml by hand.
package openapi

import _ "embed"

// SpecJSON is the canonical OpenAPI 3.1 contract as JSON bytes.
//
//go:embed v1.json
var SpecJSON []byte
