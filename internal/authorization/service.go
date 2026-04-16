package authorization

import "context"

// Expected store interface internally (matches pkg/bouncer.Store)
type Store interface {
	HasPermission(ctx context.Context, subjectID string, resourceID string, permission string) (bool, error)
	GrantPermission(ctx context.Context, subjectID string, resourceID string, permission string) error
	RevokePermission(ctx context.Context, subjectID string, resourceID string, permission string) error
}

// Service provides authorization capabilities by interacting with the storage layer.
type Service struct {
	store Store
}

// NewService creates a new Service instance backed by the provided store.
func NewService(store Store) *Service {
	return &Service{store: store}
}

// HasPermission checks if the subject has the requested permission on the resource.
func (s *Service) HasPermission(ctx context.Context, subjectID string, resourceID string, permission string) (bool, error) {
	return s.store.HasPermission(ctx, subjectID, resourceID, permission)
}

// GrantPermission grants the requested permission on the resource to the subject.
func (s *Service) GrantPermission(ctx context.Context, subjectID string, resourceID string, permission string) error {
	return s.store.GrantPermission(ctx, subjectID, resourceID, permission)
}

// RevokePermission revokes the requested permission on the resource from the subject.
func (s *Service) RevokePermission(ctx context.Context, subjectID string, resourceID string, permission string) error {
	return s.store.RevokePermission(ctx, subjectID, resourceID, permission)
}
