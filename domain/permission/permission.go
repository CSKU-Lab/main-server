// Package permission provides a flexible, type-safe permission builder and service for access control.
//
// This package provides two patterns:
// 1. Service pattern (recommended): Use NewPermissionService with middleware for route-level guards
// 2. Builder pattern (legacy): Use User().Conditions().Check() for programmatic checks
//
// The permission service uses a fluent builder pattern with composable conditions.
// It is safe for concurrent use because each request gets a fresh Builder instance.
//
// Example usage (Service pattern):
//
//	permService := permission.NewPermissionService(...)
//
//	// In route setup:
//	router.Get("/courses/:id",
//	    middleware.RequirePermission(permService.IsCourseCreator("id")),
//	    handler)
//
// Example usage (Builder pattern):
//
//	err := perm.User(userID).
//	    Conditions(
//	        perm.Or(
//	            perm.IsInSection("Section-A"),
//	            perm.IsAdmin,
//	        ),
//	    ).
//	    Check()
//	if err != nil {
//	    return err // ErrForbidden if conditions fail
//	}
package permission

import (
	"context"
	"errors"

	"github.com/CSKU-Lab/main-server/domain/models"
)

// ErrForbidden is returned when a permission check fails.
var ErrForbidden = errors.New("forbidden")

// Condition defines the interface for permission checks.
// Implementations must provide an Evaluate method that checks
// if a user meets the specified permission criteria.
type Condition interface {
	// Evaluate checks whether the given user satisfies this condition.
	// Returns true if the condition is met, false otherwise.
	// The params map contains route parameters (e.g., "id" -> "course-123").
	Evaluate(ctx context.Context, user *models.User, params map[string]string) bool
}

// IsInSection checks if a user belongs to a specific section.
// It implements both Condition and LegacyCondition interfaces for backward compatibility.
// Deprecated: Use Service.IsSectionStudent() instead.
type IsInSection string

// IsSatisfied checks whether the user is in the specified section.
// Currently returns a mock value (true) for testing purposes.
// In production, this would query the database to verify section membership.
func (sec IsInSection) IsSatisfied(userID string) bool {
	// TODO: Replace with actual database check
	// Example: check user_sections table for (userID, sectionID) pair
	return true
}

// Evaluate implements the Condition interface for IsInSection.
func (sec IsInSection) Evaluate(ctx context.Context, user *models.User, params map[string]string) bool {
	if user == nil {
		return false
	}
	// For backward compatibility, delegate to IsSatisfied
	return sec.IsSatisfied(user.ID)
}

// isAdminCondition checks if a user has admin privileges.
// It is unexported, and accessed via the IsAdmin constant.
type isAdminCondition struct{}

// IsSatisfied checks whether the user is an admin.
// Currently returns a mock value (false) for testing purposes.
// In production, this would query the database to verify admin status.
func (isAdminCondition) IsSatisfied(userID string) bool {
	// TODO: Replace with actual database check
	// Example: check users table for admin flag
	return false
}

// Evaluate implements the Condition interface for isAdminCondition.
func (isAdminCondition) Evaluate(ctx context.Context, user *models.User, params map[string]string) bool {
	if user == nil {
		return false
	}
	for _, role := range user.Roles {
		if role == models.ADMIN {
			return true
		}
	}
	return false
}

// IsAdmin is a constant-like variable that can be used in permission checks.
// It checks if a user has admin privileges.
// It implements both Condition and LegacyCondition interfaces.
var IsAdmin = isAdminCondition{}

// orCondition implements OR logic as a Condition.
// It passes if ANY of its sub-conditions are satisfied.
type orCondition struct {
	conditions []Condition
}

// IsSatisfied returns true if ANY of the sub-conditions are satisfied (legacy method).
func (or orCondition) IsSatisfied(userID string) bool {
	// Create a minimal user for evaluation
	user := &models.User{ID: userID}
	return or.Evaluate(context.Background(), user, nil)
}

// Evaluate returns true if ANY of the sub-conditions are satisfied.
func (or orCondition) Evaluate(ctx context.Context, user *models.User, params map[string]string) bool {
	for _, c := range or.conditions {
		if c.Evaluate(ctx, user, params) {
			return true
		}
	}
	return false
}

// Or combines multiple conditions with OR logic.
// The resulting condition passes if ANY of the input conditions are satisfied.
// This function accepts both Condition and LegacyCondition types.
func Or(conditions ...Condition) Condition {
	return orCondition{conditions: conditions}
}

// Builder provides a fluent API for building and checking permissions.
// Each Builder instance is request-scoped, ensuring safe concurrent use.
// Do not reuse Builder instances across requests.
type Builder struct {
	userID     string
	conditions []Condition
}

// User creates a new request-scoped permission builder for the given user.
// Each call returns a fresh Builder instance to prevent race conditions.
//
// Example:
//
//	perm.User("user-123")
func User(userID string) *Builder {
	return &Builder{userID: userID}
}

// Conditions adds one or more conditions to the builder.
// Multiple calls accumulate conditions, which are checked with AND logic.
// Returns the builder for method chaining.
//
// Example:
//
//	perm.User(userID).
//	    Conditions(perm.IsAdmin).
//	    Conditions(perm.IsInSection("Section-A"))
func (b *Builder) Conditions(conditions ...Condition) *Builder {
	b.conditions = append(b.conditions, conditions...)
	return b
}

// Check validates that all accumulated conditions are satisfied.
// Returns nil if all conditions pass, or ErrForbidden if any condition fails.
// Conditions are evaluated with AND logic - all must be true.
//
// Example:
//
//	err := perm.User(userID).
//	    Conditions(perm.IsInSection("Section-A")).
//	    Check()
//	if err != nil {
//	    return err // Handle permission denied
//	}
func (b *Builder) Check() error {
	// Create a minimal context for evaluation
	ctx := context.Background()
	user := &models.User{ID: b.userID}
	params := map[string]string{}

	// All conditions must be satisfied (AND logic)
	for _, cond := range b.conditions {
		if !cond.Evaluate(ctx, user, params) {
			return ErrForbidden
		}
	}
	return nil
}
