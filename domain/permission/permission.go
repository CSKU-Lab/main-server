// Package permission provides a flexible, type-safe permission builder for access control.
//
// The permission service uses a fluent builder pattern with composable conditions.
// It is safe for concurrent use because each request gets a fresh Builder instance.
//
// Example usage:
//
//	err := perm.User(userID).
//		Conditions(
//			perm.Or(
//				perm.IsInSection("Section-A"),
//				perm.IsAdmin,
//			),
//		).
//		Check()
//	if err != nil {
//		return err // ErrForbidden if conditions fail
//	}
package permission

import (
	"context"
	"errors"
)

// ErrForbidden is returned when a permission check fails.
var ErrForbidden = errors.New("forbidden")

// RepositoryProvider defines the interface for accessing repositories needed by permission checks.
// This is typically implemented at the application level and passed to permission conditions.
type RepositoryProvider interface {
	IsSectionInstructor(ctx context.Context, userID string, sectionID string) (bool, error)
	IsSubmissionOwner(ctx context.Context, userID string, submissionID string) (bool, error)
}

// Global repository provider - set at application startup
var repoProvider RepositoryProvider

// SetRepositoryProvider sets the global repository provider for permission checks.
// This must be called during application initialization.
func SetRepositoryProvider(provider RepositoryProvider) {
	repoProvider = provider
}

// Condition defines the interface for permission checks.
// Implementations must provide an IsSatisfied method that checks
// if a user meets the specified permission criteria.
type Condition interface {
	// IsSatisfied checks whether the given userID satisfies this condition.
	// Returns true if the condition is met, false otherwise.
	IsSatisfied(ctx context.Context, userID string) bool
}

// IsInSection checks if a user belongs to a specific section.
// It implements the Condition interface.
type IsInSection string

// IsSatisfied checks whether the user is in the specified section.
// Currently returns a mock value (true) for testing purposes.
// In production, this would query the database to verify section membership.
func (sec IsInSection) IsSatisfied(ctx context.Context, userID string) bool {
	// TODO: Replace with actual database check
	// Example: check user_sections table for (userID, sectionID) pair
	return true
}

// isAdminCondition checks if a user has admin privileges.
// It is unexported, and accessed via the IsAdmin constant.
type isAdminCondition struct{}

// IsSatisfied checks whether the user is an admin.
// Currently returns a mock value (false) for testing purposes.
// In production, this would query the database to verify admin status.
func (isAdminCondition) IsSatisfied(ctx context.Context, userID string) bool {
	// TODO: Replace with actual database check
	// Example: check users table for admin flag
	return false
}

// IsAdmin is a constant-like variable that can be used in permission checks.
// It checks if a user has admin privileges.
var IsAdmin = isAdminCondition{}

// isSectionInstructorCondition checks if a user is an instructor of a specific section.
type isSectionInstructorCondition struct {
	sectionID string
}

// IsSectionInstructor creates a condition that checks if the user is an instructor of the specified section.
// The sectionID parameter is the route parameter name that will be extracted from the URL.
func IsSectionInstructor(sectionID string) Condition {
	return isSectionInstructorCondition{sectionID: sectionID}
}

// IsSatisfied checks whether the user is an instructor of the specified section.
func (c isSectionInstructorCondition) IsSatisfied(ctx context.Context, userID string) bool {
	if repoProvider == nil {
		return false
	}
	isInstructor, err := repoProvider.IsSectionInstructor(ctx, userID, c.sectionID)
	if err != nil {
		return false
	}
	return isInstructor
}

// isSubmissionOwnerCondition checks if a user owns a specific submission.
type isSubmissionOwnerCondition struct {
	submissionID string
}

// IsSubmissionOwner creates a condition that checks if the user owns the specified submission.
// The submissionID parameter is the route parameter name that will be extracted from the URL.
func IsSubmissionOwner(submissionID string) Condition {
	return isSubmissionOwnerCondition{submissionID: submissionID}
}

// IsSatisfied checks whether the user owns the specified submission.
func (c isSubmissionOwnerCondition) IsSatisfied(ctx context.Context, userID string) bool {
	if repoProvider == nil {
		return false
	}
	isOwner, err := repoProvider.IsSubmissionOwner(ctx, userID, c.submissionID)
	if err != nil {
		return false
	}
	return isOwner
}

// orCondition implements OR logic as a Condition.
// It passes if ANY of its sub-conditions are satisfied.
type orCondition struct {
	conditions []Condition
}

// IsSatisfied returns true if ANY of the sub-conditions are satisfied.
func (or orCondition) IsSatisfied(ctx context.Context, userID string) bool {
	for _, c := range or.conditions {
		if c.IsSatisfied(ctx, userID) {
			return true
		}
	}
	return false
}

// Or combines multiple conditions with OR logic.
// The resulting condition passes if ANY of the input conditions are satisfied.
//
// Example:
//
//	perm.Or(
//		perm.IsInSection("Section-A"),
//		perm.IsAdmin,
//	)
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
//		Conditions(perm.IsAdmin).
//		Conditions(perm.IsInSection("Section-A"))
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
//		Conditions(perm.IsInSection("Section-A")).
//		Check(ctx)
//	if err != nil {
//		return err // Handle permission denied
//	}
func (b *Builder) Check(ctx context.Context) error {
	// All conditions must be satisfied (AND logic)
	for _, cond := range b.conditions {
		if !cond.IsSatisfied(ctx, b.userID) {
			return ErrForbidden
		}
	}
	return nil
}
