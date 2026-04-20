package authorization

import "context"

// Expected store interface internally (matches pkg/bouncer.Store)
type Store interface {
	HasPermission(ctx context.Context, subjectID string, resourceID string, permission string) (bool, error)
	GrantPermission(ctx context.Context, subjectID string, resourceID string, permission string) error
	RevokePermission(ctx context.Context, subjectID string, resourceID string, permission string) error
	CreateResource(ctx context.Context, resourceID string, name string, parent *string) error
	DeleteResource(ctx context.Context, resourceID string) error
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

// CreateResource creates a resource in the backing store.
func (s *Service) CreateResource(ctx context.Context, resourceID string, name string, parent *string) error {
	return s.store.CreateResource(ctx, resourceID, name, parent)
}

// DeleteResource deletes a resource from the backing store.
func (s *Service) DeleteResource(ctx context.Context, resourceID string) error {
	return s.store.DeleteResource(ctx, resourceID)
}
