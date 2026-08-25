// Package orgsnapshot defines the provider-neutral organization snapshot
// contract. It contains DTOs, versioning, owned-field declarations, and
// validation without depending on providers, persistence, or domain logic.
package orgsnapshot

import "time"

// ContractVersion is the current snapshot contract version.

// Backward-compatible additions stay within the same version.
// Breaking changes require a version bump.
const ContractVersion = "1.0"

// User represents a person from an external organization data source.
type User struct {
	// ID is the identifier shared with the domain user (see domain/user.User.ID).
	ID string `json:"id"`

	// Username is the stable login handle used to map to the domain user.
	Username string `json:"username,omitempty"`

	// DisplayName is the user's human-readable name.
	DisplayName string `json:"displayName,omitempty"`

	// Email is the user's contact address.
	Email string `json:"email,omitempty"`

	// ReportsTo is the ID of this user's direct manager, or nil if the
	// user has no manager in the provider's org data. Field name mirrors
	// domain/user.User.ReportsTo so the contract maps directly onto the
	// domain struct without translation. A user has at most one manager,
	// so this is a field on the user rather than a separate relationship
	// record.
	ReportsTo *string `json:"reportsTo,omitempty"`

	// IsActive is tri-state: true = active, false = deactivated/offboarded,
	// nil = not supplied. Providers should set this explicitly rather than
	// omitting a user from the snapshot, so Team360 can distinguish a
	// deactivation from a transient omission and avoid orphaning history.
	IsActive *bool `json:"isActive,omitempty"`
}

// Team represents a team from an external organization data source.
type Team struct {
	// ID is the identifier shared with the domain team (see domain/team.Team.ID).
	ID string `json:"id"`

	// Name is the team's human-readable name.
	Name string `json:"name"`

	// HealthCheckEnabled is tri-state:
	// true = enabled, false = disabled, nil = not supplied.
	HealthCheckEnabled *bool `json:"healthCheckEnabled,omitempty"`

	// IsActive is tri-state: true = active, false = disbanded/merged,
	// nil = not supplied. Providers should set this explicitly rather than
	// omitting a team from the snapshot, so Team360 can distinguish a
	// disbandment from a transient omission and avoid orphaning history.
	IsActive *bool `json:"isActive,omitempty"`
}

// Membership associates a user with a team.
type Membership struct {
	UserID string `json:"userId"`
	TeamID string `json:"teamId"`

	// Role is the user's role within this specific team membership
	// (e.g. "lead", "member"). It is scoped to the membership, not the
	// user or the team, because the same person can be a lead on one
	// team and a plain member on another. The vocabulary is open;
	// Team360 currently recognizes "lead" as significant and treats
	// unrecognized values as a plain member.
	Role string `json:"role,omitempty"`
}

// Snapshot is a complete provider organization snapshot.
type Snapshot struct {
	// ContractVersion is the version used to produce this snapshot.
	ContractVersion string `json:"contractVersion"`

	// GeneratedAt is when the snapshot was produced.
	GeneratedAt time.Time `json:"generatedAt"`

	Teams       []Team       `json:"teams"`
	Users       []User       `json:"users"`
	Memberships []Membership `json:"memberships"`

	// OwnedFields lists fields the provider claims authority over.
	OwnedFields []OwnedField `json:"ownedFields,omitempty"`
}
