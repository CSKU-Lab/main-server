package permission

import (
	"context"
	"testing"

	"github.com/CSKU-Lab/main-server/domain/models"
	"github.com/stretchr/testify/assert"
)

// TestIsInSection tests the IsInSection condition.
func TestIsInSection(t *testing.T) {
	tests := []struct {
		name     string
		section  IsInSection
		user     *models.User
		expected bool
	}{
		{
			name:     "user in section",
			section:  IsInSection("Section-A"),
			user:     &models.User{ID: "user-123"},
			expected: true,
		},
		{
			name:     "another section",
			section:  IsInSection("Section-B"),
			user:     &models.User{ID: "user-456"},
			expected: true,
		},
		{
			name:     "nil user",
			section:  IsInSection("Section-A"),
			user:     nil,
			expected: false,
		},
	}

	ctx := context.Background()
	params := map[string]string{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.section.Evaluate(ctx, tt.user, params)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsAdminCondition tests the isAdminCondition.
func TestIsAdminCondition(t *testing.T) {
	tests := []struct {
		name     string
		user     *models.User
		expected bool
	}{
		{
			name:     "user with admin role",
			user:     &models.User{ID: "user-123", Roles: []models.Role{models.ADMIN}},
			expected: true,
		},
		{
			name:     "user without admin role",
			user:     &models.User{ID: "user-456", Roles: []models.Role{models.STUDENT}},
			expected: false,
		},
		{
			name:     "nil user",
			user:     nil,
			expected: false,
		},
		{
			name:     "user with multiple roles including admin",
			user:     &models.User{ID: "user-789", Roles: []models.Role{models.STUDENT, models.ADMIN}},
			expected: true,
		},
	}

	ctx := context.Background()
	params := map[string]string{}
	adminCond := isAdminCondition{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adminCond.Evaluate(ctx, tt.user, params)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestOrCondition tests the OR condition logic.
func TestOrCondition(t *testing.T) {
	tests := []struct {
		name       string
		conditions []Condition
		user       *models.User
		expected   bool
	}{
		{
			name: "one condition true",
			conditions: []Condition{
				IsInSection("Section-A"),
				&isAdminCondition{},
			},
			user:     &models.User{ID: "user-123"},
			expected: true,
		},
		{
			name: "both conditions false",
			conditions: []Condition{
				&isAdminCondition{},
				&isAdminCondition{},
			},
			user:     &models.User{ID: "user-123", Roles: []models.Role{models.STUDENT}},
			expected: false,
		},
		{
			name:       "single condition true",
			conditions: []Condition{IsInSection("Section-A")},
			user:       &models.User{ID: "user-123"},
			expected:   true,
		},
		{
			name: "multiple true conditions",
			conditions: []Condition{
				IsInSection("Section-A"),
				IsInSection("Section-B"),
			},
			user:     &models.User{ID: "user-123"},
			expected: true,
		},
	}

	ctx := context.Background()
	params := map[string]string{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orCond := orCondition{conditions: tt.conditions}
			result := orCond.Evaluate(ctx, tt.user, params)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestOrEmpty tests OR with no conditions.
func TestOrEmpty(t *testing.T) {
	orCond := orCondition{conditions: []Condition{}}
	ctx := context.Background()
	params := map[string]string{}
	user := &models.User{ID: "user-123"}

	result := orCond.Evaluate(ctx, user, params)
	assert.False(t, result, "expected false for empty OR")
}

// TestAndCondition tests the AND condition logic.
func TestAndCondition(t *testing.T) {
	tests := []struct {
		name       string
		conditions []Condition
		user       *models.User
		expected   bool
	}{
		{
			name: "all conditions true",
			conditions: []Condition{
				IsInSection("Section-A"),
				IsInSection("Section-B"),
			},
			user:     &models.User{ID: "user-123"},
			expected: true,
		},
		{
			name: "one condition false",
			conditions: []Condition{
				IsInSection("Section-A"),
				&isAdminCondition{},
			},
			user:     &models.User{ID: "user-123", Roles: []models.Role{models.STUDENT}},
			expected: false,
		},
		{
			name:       "empty conditions",
			conditions: []Condition{},
			user:       &models.User{ID: "user-123"},
			expected:   true, // Empty AND returns true
		},
	}

	ctx := context.Background()
	params := map[string]string{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			andCond := andCondition{conditions: tt.conditions}
			result := andCond.Evaluate(ctx, tt.user, params)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestNotCondition tests the NOT condition logic.
func TestNotCondition(t *testing.T) {
	ctx := context.Background()
	params := map[string]string{}

	t.Run("not true is false", func(t *testing.T) {
		notCond := notCondition{condition: IsInSection("Section-A")}
		user := &models.User{ID: "user-123"}
		result := notCond.Evaluate(ctx, user, params)
		assert.False(t, result)
	})

	t.Run("not false is true", func(t *testing.T) {
		notCond := notCondition{condition: &isAdminCondition{}}
		user := &models.User{ID: "user-123", Roles: []models.Role{models.STUDENT}}
		result := notCond.Evaluate(ctx, user, params)
		assert.True(t, result)
	})
}

// TestUserBuilder tests the User builder function.
func TestUserBuilder(t *testing.T) {
	builder := User("user-123")
	assert.NotNil(t, builder)
	assert.Equal(t, "user-123", builder.userID)
	assert.Empty(t, builder.conditions)
}

// TestBuilderConditions tests adding conditions to a builder.
func TestBuilderConditions(t *testing.T) {
	builder := User("user-123").
		Conditions(IsInSection("Section-A"))

	assert.Len(t, builder.conditions, 1)

	// Test chaining
	builder.Conditions(&isAdminCondition{})
	assert.Len(t, builder.conditions, 2)
}

// TestBuilderCheckPass tests the Check method with all conditions passing.
func TestBuilderCheckPass(t *testing.T) {
	err := User("user-123").
		Conditions(IsInSection("Section-A")).
		Check()

	assert.NoError(t, err)
}

// TestBuilderCheckFail tests the Check method when conditions fail.
func TestBuilderCheckFail(t *testing.T) {
	err := User("user-123").
		Conditions(&isAdminCondition{}).
		Check()

	assert.Equal(t, ErrForbidden, err)
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
			shouldFail: false,
		},
		{
			name: "one condition fails",
			conditions: []Condition{
				IsInSection("Section-A"),
				&isAdminCondition{},
			},
			shouldFail: true,
		},
		{
			name: "both conditions fail",
			conditions: []Condition{
				&isAdminCondition{},
				&isAdminCondition{},
			},
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := User("user-123").Conditions(tt.conditions...).Check()

			if tt.shouldFail {
				assert.Equal(t, ErrForbidden, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestBuilderCheckNoConditions tests Check with no conditions.
func TestBuilderCheckNoConditions(t *testing.T) {
	err := User("user-123").Check()
	assert.NoError(t, err)
}

// TestBuilderConcurrentSafety tests that separate builders don't interfere.
func TestBuilderConcurrentSafety(t *testing.T) {
	builder1 := User("user-1")
	builder2 := User("user-2")

	builder1.Conditions(IsInSection("Section-A"))
	builder2.Conditions(&isAdminCondition{})

	assert.Equal(t, "user-1", builder1.userID)
	assert.Equal(t, "user-2", builder2.userID)
	assert.Len(t, builder1.conditions, 1)
	assert.Len(t, builder2.conditions, 1)
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
					&isAdminCondition{},
					IsInSection("Section-A"),
				),
			},
			shouldFail: false,
		},
		{
			name: "admin AND (admin OR section) (fails)",
			conditions: []Condition{
				&isAdminCondition{},
				Or(
					&isAdminCondition{},
					IsInSection("Section-A"),
				),
			},
			shouldFail: true,
		},
		{
			name: "section-a AND (admin OR section-a) (passes)",
			conditions: []Condition{
				IsInSection("Section-A"),
				Or(
					&isAdminCondition{},
					IsInSection("Section-A"),
				),
			},
			shouldFail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := User("user-123").Conditions(tt.conditions...).Check()

			if tt.shouldFail {
				assert.Equal(t, ErrForbidden, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestErrForbiddenValue tests that ErrForbidden is properly defined.
func TestErrForbiddenValue(t *testing.T) {
	assert.NotNil(t, ErrForbidden)
	assert.Equal(t, "forbidden", ErrForbidden.Error())
}

// TestBuilderMethodChaining tests that builder methods return the builder for chaining.
func TestBuilderMethodChaining(t *testing.T) {
	builder := User("user-123")
	returned := builder.Conditions(IsInSection("Section-A"))

	assert.Equal(t, builder, returned)

	// Test full chain
	err := User("user-123").
		Conditions(IsInSection("Section-A")).
		Conditions(IsInSection("Section-B")).
		Check()

	assert.NoError(t, err)
}

// TestIsAdminVariable tests the IsAdmin variable.
func TestIsAdminVariable(t *testing.T) {
	ctx := context.Background()
	params := map[string]string{}

	t.Run("IsAdmin with admin user", func(t *testing.T) {
		user := &models.User{ID: "user-123", Roles: []models.Role{models.ADMIN}}
		result := IsAdmin.Evaluate(ctx, user, params)
		assert.True(t, result)
	})

	t.Run("IsAdmin with non-admin user", func(t *testing.T) {
		user := &models.User{ID: "user-123", Roles: []models.Role{models.STUDENT}}
		result := IsAdmin.Evaluate(ctx, user, params)
		assert.False(t, result)
	})
}
