// Command openapi-yaml renders api/openapi/v1.yaml from api/openapi/v1.json.
//
// swag emits both, but only its JSON is deterministic: Go's JSON encoder sorts
// map keys, and swag's YAML writer iterates the map. Two lines of the LiveStats
// schema therefore changed position on roughly a third of runs, which made the
// OpenAPI drift guard in `make ci-local` permanently unreliable -- it
// regenerates the contract and diffs it, so it could never come back clean, and
// a gate that is red whatever you do stops carrying information.
//
// The conversion goes through a yaml.Node rather than a map. YAML is a superset
// of JSON, so the parser reads the contract into a node tree that preserves
// DOCUMENT ORDER, and re-emitting that tree reproduces the JSON's key order
// exactly.
//
// Not through a map[string]any: that is what the first attempt at this did, on
// the assumption that the marshaller would sort keys. It does not -- three runs
// over one fixed input produced two different files -- and the bug was simply
// moved from swag's encoder into this one.
package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const yamlIndent = 2

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: openapi-yaml <input.json> <output.yaml>")
		os.Exit(2)
	}
	if err := run(os.Args[1], os.Args[2]); err != nil {
		fmt.Fprintln(os.Stderr, "openapi-yaml:", err)
		os.Exit(1)
	}
}

func run(in, out string) error {
	src, err := os.ReadFile(in)
	if err != nil {
		return err
	}

	converted, err := convert(src)
	if err != nil {
		return fmt.Errorf("convert %s: %w", in, err)
	}
	return os.WriteFile(out, converted, 0o644)
}

// convert turns a JSON document into YAML, preserving key order.
func convert(src []byte) ([]byte, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(src, &doc); err != nil {
		return nil, err
	}
	blockStyle(&doc)

	var buf writeCollector
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(yamlIndent)
	if err := enc.Encode(&doc); err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	return buf.b, nil
}

// blockStyle strips the flow style the parser carries over from JSON.
//
// yaml.v3 preserves the input's presentation, and JSON braces are flow style, so
// re-emitting the parsed tree unchanged produced the entire 5,000-line contract
// on a single line. Valid YAML, and unreadable -- an artifact people open to read
// the API has to look like one.
func blockStyle(n *yaml.Node) {
	if n == nil {
		return
	}
	n.Style = 0
	for _, c := range n.Content {
		blockStyle(c)
	}
}

type writeCollector struct{ b []byte }

func (w *writeCollector) Write(p []byte) (int, error) {
	w.b = append(w.b, p...)
	return len(p), nil
}
