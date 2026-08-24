package orgsnapshot_test

import (
	"context"
	"testing"

	"github.com/agopalakrishnan/teams360/backend/pkg/orgsnapshot"
)

// fakeProvider is a minimal stand-in used only to confirm SnapshotProvider
// is implementable without pulling in any HTTP/auth/persistence concerns.
type fakeProvider struct {
	snapshot orgsnapshot.Snapshot
}

func (f fakeProvider) GetOrganizationSnapshot(ctx context.Context, version string) (orgsnapshot.Snapshot, error) {
	return f.snapshot, nil
}

func TestSnapshotProvider_Implementable(t *testing.T) {
	var p orgsnapshot.SnapshotProvider = fakeProvider{snapshot: validSnapshot()}

	snap, err := p.GetOrganizationSnapshot(context.Background(), orgsnapshot.ContractVersion)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if errs := snap.Validate(); len(errs) != 0 {
		t.Fatalf("expected fake provider's snapshot to be valid, got %+v", errs)
	}
}
