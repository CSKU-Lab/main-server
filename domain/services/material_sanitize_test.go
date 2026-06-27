package services

import (
	"reflect"
	"testing"

	"github.com/CSKU-Lab/main-server/domain/registrables"
)

// fseg builds a file segment.
func fseg(content, segType string) registrables.FileSegment {
	return registrables.FileSegment{Content: content, Type: segType}
}

func TestSanitizeStudentFiles(t *testing.T) {
	t.Run("blanks hidden segment content but keeps its position and type", func(t *testing.T) {
		// Mirrors the CS-232 leak: a hidden grader line sits between editable
		// and readonly segments. The hidden content must not reach the student,
		// but the segment must stay in place so editable indices the client
		// submits still line up with the grader's assembly.
		in := []registrables.File{{
			Name:    "main.py",
			Content: "C = 25.0\nC = float(input())\n#---\n\n",
			Segments: []registrables.FileSegment{
				fseg("C = ", "readonly"),
				fseg("25.0\n", "editable"),
				fseg("C = float(input())\n", "hidden"),
				fseg("#---", "readonly"),
				fseg("\n\n", "editable"),
			},
		}}

		got := sanitizeStudentFiles(in)

		wantSegments := []registrables.FileSegment{
			fseg("C = ", "readonly"),
			fseg("25.0\n", "editable"),
			fseg("\n", "hidden"), // content stripped; trailing newline kept, type + index preserved
			fseg("#---", "readonly"),
			fseg("\n\n", "editable"),
		}
		if !reflect.DeepEqual(got[0].Segments, wantSegments) {
			t.Errorf("segments = %#v, want %#v", got[0].Segments, wantSegments)
		}

		// Flat content must be rebuilt from non-hidden segments only.
		wantContent := "C = 25.0\n#---\n\n"
		if got[0].Content != wantContent {
			t.Errorf("content = %q, want %q", got[0].Content, wantContent)
		}

		// Editable segment positions are unchanged (indices 1 and 4).
		if got[0].Segments[1].Type != "editable" || got[0].Segments[4].Type != "editable" {
			t.Errorf("editable segment indices shifted: %#v", got[0].Segments)
		}
	})

	t.Run("leaves readonly and editable untouched", func(t *testing.T) {
		in := []registrables.File{{
			Name: "a.c",
			Segments: []registrables.FileSegment{
				fseg("ro\n", "readonly"),
				fseg("ed\n", "editable"),
			},
		}}

		got := sanitizeStudentFiles(in)

		if !reflect.DeepEqual(got[0].Segments, in[0].Segments) {
			t.Errorf("readonly/editable segments altered: %#v", got[0].Segments)
		}
		if got[0].Content != "ro\ned\n" {
			t.Errorf("content = %q, want %q", got[0].Content, "ro\ned\n")
		}
	})

	t.Run("recasts exclude to readonly but keeps its content", func(t *testing.T) {
		// exclude is a visible hint; the student keeps seeing the text but must
		// not learn it is excluded from grading, so the type becomes readonly.
		in := []registrables.File{{
			Name: "a.c",
			Segments: []registrables.FileSegment{
				fseg("ed\n", "editable"),
				fseg("hint\n", "exclude"),
			},
		}}

		got := sanitizeStudentFiles(in)

		want := []registrables.FileSegment{
			fseg("ed\n", "editable"),
			fseg("hint\n", "readonly"), // recast, content preserved
		}
		if !reflect.DeepEqual(got[0].Segments, want) {
			t.Errorf("segments = %#v, want %#v", got[0].Segments, want)
		}
		if got[0].Content != "ed\nhint\n" {
			t.Errorf("content = %q, want %q", got[0].Content, "ed\nhint\n")
		}
	})

	t.Run("preserves trailing newline so the client fold does not shift indices", func(t *testing.T) {
		// A hidden segment ending in "\n" followed by an editable that is exactly
		// "\n": today normalizeHiddenSegments does not fold (hidden already ends
		// in "\n"). Blanking the hidden content must keep that trailing newline,
		// otherwise the client would fold and drop the "\n" editable, shifting
		// every later editable index and corrupting grader assembly.
		in := []registrables.File{{
			Name: "a.py",
			Segments: []registrables.FileSegment{
				fseg("ans = ", "readonly"),
				fseg("0", "editable"),
				fseg("\nsecret()\n", "hidden"),
				fseg("\n", "editable"),
				fseg("end", "readonly"),
			},
		}}

		got := sanitizeStudentFiles(in)

		// Hidden keeps its trailing newline; the "\n" editable at index 3 stays.
		if got[0].Segments[2] != (fseg("\n", "hidden")) {
			t.Errorf("hidden segment = %#v, want trailing-newline-preserving blank", got[0].Segments[2])
		}
		if got[0].Segments[3] != (fseg("\n", "editable")) {
			t.Errorf("editable index 3 changed: %#v", got[0].Segments[3])
		}
		if got[0].Content != "ans = 0\nend" {
			t.Errorf("content = %q, want %q", got[0].Content, "ans = 0\nend")
		}
	})

	t.Run("passes through files with no segments (backward compat)", func(t *testing.T) {
		in := []registrables.File{{Name: "x", Content: "whole file"}}
		got := sanitizeStudentFiles(in)
		if !reflect.DeepEqual(got, in) {
			t.Errorf("segment-less file altered: %#v", got)
		}
	})
}
