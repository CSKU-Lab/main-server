package permission

import (
	"context"
	"testing"
)

// TestIsInSection tests the IsInSection condition.
func TestIsInSection(t *testing.T) {
	tests := []struct {
		name     string
		section  IsInSection
		userID   string
		expected bool
	}{
		{
			name:     "user in section",
			section:  IsInSection("Section-A"),
			userID:   "user-123",
			expected: true,
		},
		{
			name:     "another section",
			section:  IsInSection("Section-B"),
			userID:   "user-456",
			expected: true,
		},
		{
			name:     "empty user id",
			section:  IsInSection("Section-A"),
			userID:   "",
			expected: true,
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.section.IsSatisfied(ctx, tt.userID)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestIsAdmin tests the IsAdmin condition.
func TestIsAdmin(t *testing.T) {
	tests := []struct {
		name     string
		userID   string
		expected bool
	}{
		{
			name:     "admin check",
			userID:   "user-123",
			expected: false, // Mock returns false
		},
		{
			name:     "another user",
			userID:   "user-456",
			expected: false,
		},
		{
			name:     "empty user id",
			userID:   "",
			expected: false,
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAdmin.IsSatisfied(ctx, tt.userID)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestIsSectionInstructor tests the IsSectionInstructor condition.
func TestIsSectionInstructor(t *testing.T) {
	tests := []struct {
		name     string
		param    string
		userID   string
		expected bool
	}{
		{
			name:     "instructor in section",
			param:    "id",
			userID:   "user-123",
			expected: true, // Mock returns true
		},
		{
			name:     "another instructor",
			param:    "sectionID",
			userID:   "user-456",
			expected: true, // Mock returns true
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond := IsSectionInstructor(tt.param)
			result := cond.IsSatisfied(ctx, tt.userID)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestIsSectionStudent tests the IsSectionStudent condition.
func TestIsSectionStudent(t *testing.T) {
	tests := []struct {
		name     string
		param    string
		userID   string
		expected bool
	}{
		{
			name:     "student in section",
			param:    "id",
			userID:   "user-123",
			expected: true, // Mock returns true
		},
		{
			name:     "another student",
			param:    "sectionID",
			userID:   "user-456",
			expected: true, // Mock returns true
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond := IsSectionStudent(tt.param)
			result := cond.IsSatisfied(ctx, tt.userID)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestOr tests the OR condition logic.
func TestOr(t *testing.T) {
	tests := []struct {
		name       string
		conditions []Condition
		userID     string
		expected   bool
	}{
		{
			name: "one condition true",
			conditions: []Condition{
				IsInSection("Section-A"),
				IsAdmin,
			},
			userID:   "user-123",
			expected: true, // IsInSection returns true (mock)
		},
		{
			name: "both conditions false",
			conditions: []Condition{
				IsAdmin,
				IsAdmin,
			},
			userID:   "user-123",
			expected: false, // Both return false
		},
		{
			name:       "single condition true",
			conditions: []Condition{IsInSection("Section-A")},
			userID:     "user-123",
			expected:   true,
		},
		{
			name: "multiple true conditions",
			conditions: []Condition{
				IsInSection("Section-A"),
				IsInSection("Section-B"),
			},
			userID:   "user-123",
			expected: true, // First condition is true
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orCond := Or(tt.conditions...)
			result := orCond.IsSatisfied(ctx, tt.userID)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestOrEmpty tests OR with no conditions.
func TestOrEmpty(t *testing.T) {
	ctx := context.Background()
	orCond := Or()
	result := orCond.IsSatisfied(ctx, "user-123")
	if result {
		t.Errorf("expected false for empty OR, got %v", result)
	}
}

// TestUserBuilder tests the User builder function.
func TestUserBuilder(t *testing.T) {
	builder := User("user-123")
	if builder == nil {
		t.Error("expected non-nil builder")
	}
	if builder.userID != "user-123" {
		t.Errorf("expected userID 'user-123', got %q", builder.userID)
	}
	if len(builder.conditions) != 0 {
		t.Errorf("expected 0 conditions, got %d", len(builder.conditions))
	}
}

// TestBuilderConditions tests adding conditions to a builder.
func TestBuilderConditions(t *testing.T) {
	builder := User("user-123").
		Conditions(IsInSection("Section-A"))

	if len(builder.conditions) != 1 {
		t.Errorf("expected 1 condition, got %d", len(builder.conditions))
	}

	// Test chaining
	builder.Conditions(IsAdmin, IsInSection("Section-B"))
	if len(builder.conditions) != 3 {
		t.Errorf("expected 3 conditions, got %d", len(builder.conditions))
	}
}

// TestBuilderCheck tests the Check method with all conditions passing.
func TestBuilderCheckPass(t *testing.T) {
	ctx := context.Background()
	err := User("user-123").
		Conditions(IsInSection("Section-A")).
		Check(ctx)

	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

// TestBuilderCheckFail tests the Check method when conditions fail.
func TestBuilderCheckFail(t *testing.T) {
	ctx := context.Background()
	err := User("user-123").
		Conditions(IsAdmin). // Mock returns false
		Check(ctx)

	if err != ErrForbidden {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

// TestBuilderCheckMultipleConditions tests AND logic with multiple conditions.
func TestBuilderCheckMultipleConditions(t *testing.T) {
	tests := []struct {
		name       string
		conditions []Condition
		shouldFail bool
	}{
		{
			name: "all conditions pass",
			conditions: []Condition{
				IsInSection("Section-A"),
				IsInSection("Section-B"),
			},
			shouldFail: false, // All return true (mock)
		},
		{
			name: "one condition fails",
			conditions: []Condition{
				IsInSection("Section-A"),
				IsAdmin, // Returns false (mock)
			},
			shouldFail: true, // AND logic: one false fails entire check
		},
		{
			name: "both conditions fail",
			conditions: []Condition{
				IsAdmin,
				IsAdmin,
			},
			shouldFail: true,
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := User("user-123").Conditions(tt.conditions...).Check(ctx)

			if tt.shouldFail && err != ErrForbidden {
				t.Errorf("expected ErrForbidden, got %v", err)
			}
			if !tt.shouldFail && err != nil {
				t.Errorf("expected nil, got %v", err)
			}
		})
	}
}

// TestBuilderCheckNoConditions tests Check with no conditions.
func TestBuilderCheckNoConditions(t *testing.T) {
	ctx := context.Background()
	err := User("user-123").Check(ctx)
	if err != nil {
		t.Errorf("expected nil error for no conditions, got %v", err)
	}
}

// TestBuilderConcurrentSafety tests that separate builders don't interfere.
func TestBuilderConcurrentSafety(t *testing.T) {
	builder1 := User("user-1")
	builder2 := User("user-2")

	builder1.Conditions(IsInSection("Section-A"))
	builder2.Conditions(IsAdmin)

	if builder1.userID != "user-1" {
		t.Errorf("builder1 userID changed: expected 'user-1', got %q", builder1.userID)
	}
	if builder2.userID != "user-2" {
		t.Errorf("builder2 userID changed: expected 'user-2', got %q", builder2.userID)
	}

	if len(builder1.conditions) != 1 {
		t.Errorf("builder1 conditions polluted: expected 1, got %d", len(builder1.conditions))
	}
	if len(builder2.conditions) != 1 {
		t.Errorf("builder2 conditions polluted: expected 1, got %d", len(builder2.conditions))
	}
}

// TestComplexPermissionCheck tests nested OR conditions with AND.
func TestComplexPermissionCheck(t *testing.T) {
	tests := []struct {
		name       string
		conditions []Condition
		shouldFail bool
	}{
		{
			name: "admin OR section-a (passes)",
			conditions: []Condition{
				Or(
					IsAdmin,
					IsInSection("Section-A"),
				),
			},
			shouldFail: false, // OR passes because IsInSection is true
		},
		{
			name: "admin AND (admin OR section) (fails)",
			conditions: []Condition{
				IsAdmin, // Fails
				Or(
					IsAdmin,
					IsInSection("Section-A"),
				),
			},
			shouldFail: true, // AND fails because IsAdmin is false
		},
		{
			name: "section-a AND (admin OR section-a) (passes)",
			conditions: []Condition{
				IsInSection("Section-A"),
				Or(
					IsAdmin,
					IsInSection("Section-A"),
				),
			},
			shouldFail: false, // Both conditions pass
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := User("user-123").Conditions(tt.conditions...).Check(ctx)

			if tt.shouldFail && err != ErrForbidden {
				t.Errorf("expected ErrForbidden, got %v", err)
			}
			if !tt.shouldFail && err != nil {
				t.Errorf("expected nil, got %v", err)
			}
		})
	}
}

// TestErrForbiddenValue tests that ErrForbidden is properly defined.
func TestErrForbiddenValue(t *testing.T) {
	if ErrForbidden == nil {
		t.Error("expected ErrForbidden to be non-nil")
	}
	if ErrForbidden.Error() != "forbidden" {
		t.Errorf("expected error message 'forbidden', got %q", ErrForbidden.Error())
	}
}

// TestBuilderMethodChaining tests that builder methods return the builder for chaining.
func TestBuilderMethodChaining(t *testing.T) {
	builder := User("user-123")
	returned := builder.Conditions(IsInSection("Section-A"))

	if returned != builder {
		t.Error("expected Conditions to return the same builder for chaining")
	}

	// Test full chain
	ctx := context.Background()
	err := User("user-123").
		Conditions(IsInSection("Section-A")).
		Conditions(IsInSection("Section-B")).
		Check(ctx)

	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}
