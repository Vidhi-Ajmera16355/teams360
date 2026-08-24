package orgsnapshot

import "context"

// SnapshotProvider provides organization data as a Snapshot.
type SnapshotProvider interface {
	// GetOrganizationSnapshot returns a snapshot for the requested contract version.
	// The caller is responsible for validating the returned snapshot.
	GetOrganizationSnapshot(ctx context.Context, version string) (Snapshot, error)
}
 