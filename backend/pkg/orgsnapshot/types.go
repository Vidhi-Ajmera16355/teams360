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

	// DisplayName is the user's human-readable name.
	DisplayName string `json:"displayName,omitempty"`

	// Email is the user's contact address.
	Email string `json:"email,omitempty"`
}

// Team represents a team from an external organization data source.
type Team struct {
	// ID is the identifier shared with the domain team (see domain/team.Team.ID).
	ID string `json:"id"`

	// Name is the team's human-readable name.
	Name string `json:"name"`

	// ParentTeamID references the parent team, if any.
	ParentTeamID string `json:"parentTeamId,omitempty"`

	// HealthCheckEnabled is tri-state:
	// true = enabled, false = disabled, nil = not supplied.
	HealthCheckEnabled *bool `json:"healthCheckEnabled,omitempty"`
}

// Membership associates a user with a team.
type Membership struct {
	UserID string `json:"userId"`
	TeamID string `json:"teamId"`
}

// ReportingRelationship represents a manager -> report relationship.
type ReportingRelationship struct {
	ManagerID string `json:"managerId"`
	ReportID  string `json:"reportId"`
}

// Snapshot is a complete provider organization snapshot.
type Snapshot struct {
	// ContractVersion is the version used to produce this snapshot.
	ContractVersion string `json:"contractVersion"`

	// GeneratedAt is when the snapshot was produced.
	GeneratedAt time.Time `json:"generatedAt"`

	Teams                  []Team                  `json:"teams"`
	Users                  []User                  `json:"users"`
	Memberships            []Membership            `json:"memberships"`
	ReportingRelationships []ReportingRelationship `json:"reportingRelationships"`

	// OwnedFields lists fields the provider claims authority over.
	OwnedFields []OwnedField `json:"ownedFields,omitempty"`
}
