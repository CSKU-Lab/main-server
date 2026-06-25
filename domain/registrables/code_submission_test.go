package registrables

import (
	"reflect"
	"testing"

	"github.com/CSKU-Lab/main-server/domain/models"
	taskPB "github.com/CSKU-Lab/main-server/genproto/task/v1"
)

// seg is a tiny helper to build a task segment.
func seg(content, segType string) *taskPB.Segment {
	return &taskPB.Segment{Content: content, Type: segType}
}

// task builds a TaskResponse with a single allowed runner and its files.
func task(runnerID string, files ...*taskPB.File) *taskPB.TaskResponse {
	return &taskPB.TaskResponse{
		AllowedRunners: []*taskPB.AllowedRunner{
			{RunnerId: runnerID, Files: files},
		},
	}
}

func TestAssembleGraderFiles(t *testing.T) {
	tests := []struct {
		name      string
		submitted []submittedFile
		runnerID  string
		task      *taskPB.TaskResponse
		want      models.SubmissionFiles
	}{
		{
			// Full assembly: editable replaced by student content, readonly + hidden
			// kept, exclude dropped, segment order preserved.
			name:     "segmented file assembles editable+readonly+hidden, drops exclude",
			runnerID: "cpp",
			submitted: []submittedFile{
				{
					Name: "main.cpp",
					EditableSegments: []editableSegment{
						{Index: 1, Content: "int x = student_value();"},
					},
				},
			},
			task: task("cpp", &taskPB.File{
				Name: "main.cpp",
				Segments: []*taskPB.Segment{
					seg("#include <iostream>\n", "readonly"),     // index 0
					seg("/*will be replaced*/", "editable"),       // index 1
					seg("// hidden grader main\n", "hidden"),      // index 2
					seg("// hint only, not compiled\n", "exclude"), // index 3
				},
			}),
			want: models.SubmissionFiles{
				{
					Name:    "main.cpp",
					Content: "#include <iostream>\nint x = student_value();// hidden grader main\n",
				},
			},
		},
		{
			// Multiple editable segments map to the right slots by index.
			name:     "multiple editable segments mapped by index",
			runnerID: "py",
			submitted: []submittedFile{
				{
					Name: "main.py",
					EditableSegments: []editableSegment{
						{Index: 0, Content: "a=1\n"},
						{Index: 2, Content: "c=3\n"},
					},
				},
			},
			task: task("py", &taskPB.File{
				Name: "main.py",
				Segments: []*taskPB.Segment{
					seg("", "editable"),       // index 0 -> a=1
					seg("b=2\n", "readonly"),  // index 1
					seg("", "editable"),       // index 2 -> c=3
				},
			}),
			want: models.SubmissionFiles{
				{Name: "main.py", Content: "a=1\nb=2\nc=3\n"},
			},
		},
		{
			// An editable segment with no matching submitted index assembles as empty.
			name:     "missing editable index becomes empty string",
			runnerID: "py",
			submitted: []submittedFile{
				{
					Name:             "main.py",
					EditableSegments: []editableSegment{}, // student submitted nothing
				},
			},
			task: task("py", &taskPB.File{
				Name: "main.py",
				Segments: []*taskPB.Segment{
					seg("header\n", "readonly"),
					seg("PLACEHOLDER", "editable"), // index 1, no submission -> ""
				},
			}),
			want: models.SubmissionFiles{
				{Name: "main.py", Content: "header\n"},
			},
		},
		{
			// Backward compat: task file has no segments -> first editable segment
			// content is used as the full file content.
			name:     "no segments uses first editable content (backward compat)",
			runnerID: "py",
			submitted: []submittedFile{
				{
					Name: "main.py",
					EditableSegments: []editableSegment{
						{Index: 0, Content: "print('hi')"},
					},
				},
			},
			task: task("py", &taskPB.File{Name: "main.py"}), // no segments
			want: models.SubmissionFiles{
				{Name: "main.py", Content: "print('hi')"},
			},
		},
		{
			// Backward compat: file name not present in the task at all.
			name:     "unknown task file uses first editable content (backward compat)",
			runnerID: "py",
			submitted: []submittedFile{
				{
					Name: "extra.py",
					EditableSegments: []editableSegment{
						{Index: 0, Content: "x=1"},
					},
				},
			},
			task: task("py"), // no files
			want: models.SubmissionFiles{
				{Name: "extra.py", Content: "x=1"},
			},
		},
		{
			// Backward compat with no editable segments at all -> empty content.
			name:     "no segments and no editable content yields empty",
			runnerID: "py",
			submitted: []submittedFile{
				{Name: "main.py", EditableSegments: nil},
			},
			task: task("py", &taskPB.File{Name: "main.py"}),
			want: models.SubmissionFiles{
				{Name: "main.py", Content: ""},
			},
		},
		{
			// Runner mismatch: no allowed runner matches -> taskFiles empty ->
			// every submitted file falls through to backward-compat path.
			name:     "runner mismatch falls back to flat content",
			runnerID: "go", // task only allows "cpp"
			submitted: []submittedFile{
				{
					Name: "main.cpp",
					EditableSegments: []editableSegment{
						{Index: 0, Content: "int main(){}"},
					},
				},
			},
			task: task("cpp", &taskPB.File{
				Name:     "main.cpp",
				Segments: []*taskPB.Segment{seg("x", "readonly")},
			}),
			want: models.SubmissionFiles{
				{Name: "main.cpp", Content: "int main(){}"},
			},
		},
		{
			// Multiple files are each assembled independently and order preserved.
			name:     "multiple files assembled independently",
			runnerID: "cpp",
			submitted: []submittedFile{
				{
					Name:             "a.cpp",
					EditableSegments: []editableSegment{{Index: 0, Content: "A"}},
				},
				{
					Name:             "b.cpp",
					EditableSegments: []editableSegment{{Index: 1, Content: "B"}},
				},
			},
			task: task("cpp",
				&taskPB.File{
					Name:     "a.cpp",
					Segments: []*taskPB.Segment{seg("", "editable"), seg("//a", "readonly")},
				},
				&taskPB.File{
					Name:     "b.cpp",
					Segments: []*taskPB.Segment{seg("//b", "readonly"), seg("", "editable")},
				},
			),
			want: models.SubmissionFiles{
				{Name: "a.cpp", Content: "A//a"},
				{Name: "b.cpp", Content: "//bB"},
			},
		},
		{
			// All segments excluded -> empty assembled content.
			name:     "all exclude segments yield empty content",
			runnerID: "py",
			submitted: []submittedFile{
				{Name: "main.py", EditableSegments: []editableSegment{}},
			},
			task: task("py", &taskPB.File{
				Name:     "main.py",
				Segments: []*taskPB.Segment{seg("hint", "exclude"), seg("more", "exclude")},
			}),
			want: models.SubmissionFiles{
				{Name: "main.py", Content: ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := assembleGraderFiles(tt.submitted, tt.runnerID, tt.task)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("assembled mismatch\n got: %#v\nwant: %#v", got, tt.want)
			}
		})
	}
}
