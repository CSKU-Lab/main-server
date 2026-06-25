//go:build ignore

// Demo program. Run: go run assemble_demo.go
// Shows how assembleGraderFiles turns task template segments + student editable
// segments into the flat file content that go-grader compiles/runs.
// Writes a human-readable report to assemble_report.txt.
package main

import (
	"fmt"
	"os"
	"strings"
)

// --- Minimal mirrors of the real types (kept local so the demo is standalone) ---

type segment struct {
	Content string
	Type    string // editable | readonly | hidden | exclude
}

type editableSegment struct {
	Index   int
	Content string
}

// assemble mirrors registrables.assembleGraderFiles for a single file.
func assemble(taskSegments []segment, submitted []editableSegment) string {
	// No template segments -> backward compat: first editable content as full file.
	if len(taskSegments) == 0 {
		if len(submitted) > 0 {
			return submitted[0].Content
		}
		return ""
	}

	editableByIndex := make(map[int]string, len(submitted))
	for _, es := range submitted {
		editableByIndex[es.Index] = es.Content
	}

	var b strings.Builder
	for i, seg := range taskSegments {
		switch seg.Type {
		case "editable":
			b.WriteString(editableByIndex[i])
		case "readonly", "hidden":
			b.WriteString(seg.Content)
		case "exclude":
			// dropped
		}
	}
	return b.String()
}

type scenario struct {
	name      string
	segments  []segment
	submitted []editableSegment
}

func main() {
	scenarios := []scenario{
		{
			name: "full assembly (editable+readonly+hidden, drops exclude)",
			segments: []segment{
				{"#include <iostream>\n", "readonly"},
				{"/*replaced*/", "editable"},
				{"// hidden grader main\n", "hidden"},
				{"// hint, not compiled\n", "exclude"},
			},
			submitted: []editableSegment{{1, "int x = student_value();"}},
		},
		{
			name: "multiple editable segments by index",
			segments: []segment{
				{"", "editable"},
				{"b=2\n", "readonly"},
				{"", "editable"},
			},
			submitted: []editableSegment{{0, "a=1\n"}, {2, "c=3\n"}},
		},
		{
			name: "missing editable index -> empty",
			segments: []segment{
				{"header\n", "readonly"},
				{"PLACEHOLDER", "editable"},
			},
			submitted: []editableSegment{},
		},
		{
			name:      "no segments -> first editable (backward compat)",
			segments:  nil,
			submitted: []editableSegment{{0, "print('hi')"}},
		},
		{
			name: "all exclude -> empty",
			segments: []segment{
				{"hint", "exclude"},
				{"more", "exclude"},
			},
			submitted: []editableSegment{},
		},
	}

	var out strings.Builder
	out.WriteString("assembleGraderFiles — assembly demo\n")
	out.WriteString("===================================\n\n")

	for n, sc := range scenarios {
		fmt.Fprintf(&out, "[%d] %s\n", n+1, sc.name)
		out.WriteString("  template segments:\n")
		if len(sc.segments) == 0 {
			out.WriteString("    (none)\n")
		}
		for i, seg := range sc.segments {
			fmt.Fprintf(&out, "    #%d %-9s %q\n", i, seg.Type, seg.Content)
		}
		out.WriteString("  student editable:\n")
		if len(sc.submitted) == 0 {
			out.WriteString("    (none)\n")
		}
		for _, es := range sc.submitted {
			fmt.Fprintf(&out, "    @%d %q\n", es.Index, es.Content)
		}
		result := assemble(sc.segments, sc.submitted)
		fmt.Fprintf(&out, "  => assembled (%d bytes):\n", len(result))
		fmt.Fprintf(&out, "    %q\n\n", result)
	}

	const path = "assemble_report.txt"
	if err := os.WriteFile(path, []byte(out.String()), 0644); err != nil {
		fmt.Fprintln(os.Stderr, "write failed:", err)
		os.Exit(1)
	}
	fmt.Print(out.String())
	fmt.Printf("\nwritten to %s\n", path)
}
