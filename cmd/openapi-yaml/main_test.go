package main

import (
	"strings"
	"testing"
)

// The whole point of this command is that the same input always produces the
// same bytes. The first implementation went through a map[string]any on the
// assumption that the marshaller sorted keys; it does not, and three runs over
// one fixed input produced two different files. This test would have caught it.
func TestConvert_IsDeterministic(t *testing.T) {
	src := []byte(`{
	  "openapi": "3.1.0",
	  "components": {
	    "schemas": {
	      "LiveStats": {
	        "type": "object",
	        "properties": {
	          "uptime_2h":  {"type": "number"},
	          "uptime_24h": {"type": "number"},
	          "uptime_7d":  {"type": "number"},
	          "uptime_30d": {"type": "number"},
	          "avg_response_time_24h": {"type": "integer"},
	          "last_response_time": {"type": "integer"}
	        }
	      }
	    }
	  }
	}`)

	first, err := convert(src)
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	for i := 0; i < 50; i++ {
		again, err := convert(src)
		if err != nil {
			t.Fatalf("convert (iteration %d): %v", i, err)
		}
		if string(again) != string(first) {
			t.Fatalf("output changed on iteration %d:\n--- first ---\n%s\n--- again ---\n%s", i, first, again)
		}
	}
}

// Key order comes from the document, not from a map. The JSON encoder sorts
// keys, so the contract's order is stable — and reproducing it is what makes the
// regenerate-and-diff guard meaningful.
func TestConvert_PreservesDocumentOrder(t *testing.T) {
	src := []byte(`{"zebra": 1, "apple": 2, "mango": 3}`)

	out, err := convert(src)
	if err != nil {
		t.Fatalf("convert: %v", err)
	}

	got := string(out)
	zebra, apple, mango := strings.Index(got, "zebra"), strings.Index(got, "apple"), strings.Index(got, "mango")
	if !(zebra < apple && apple < mango) {
		t.Errorf("keys were reordered; document order must survive:\n%s", got)
	}
}

func TestConvert_RejectsGarbage(t *testing.T) {
	if _, err := convert([]byte("{not json at all")); err == nil {
		t.Error("expected an error on malformed input rather than a half-written contract")
	}
}

// yaml.v3 preserves the input presentation, and JSON braces are flow style, so
// an unstripped tree re-emits the whole contract on one line. Valid, and
// unreadable — this is an artifact people open to read the API.
func TestConvert_EmitsBlockStyle(t *testing.T) {
	src := []byte(`{"components": {"schemas": {"Foo": {"type": "object"}}}}`)

	out, err := convert(src)
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	got := string(out)
	if strings.Contains(got, "{") || strings.Contains(got, "}") {
		t.Errorf("flow style survived; the contract would render as one line:\n%s", got)
	}
	if lines := strings.Count(strings.TrimSpace(got), "\n") + 1; lines < 4 {
		t.Errorf("expected nested block mappings across several lines, got %d:\n%s", lines, got)
	}
}
