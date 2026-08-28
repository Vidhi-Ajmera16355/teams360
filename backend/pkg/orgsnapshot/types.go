// Package orgsnapshot defines the provider-neutral organization snapshot contract.
package orgsnapshot

import "time"

// ContractVersion is the current snapshot contract version.
const ContractVersion = "1.0"

// User represents a person from an external organization data source.
type User struct {
	ID          string  `json:"id"`
	Username    string  `json:"username,omitempty"`
	DisplayName string  `json:"displayName,omitempty"`
	Email       string  `json:"email,omitempty"`
	ReportsTo   *string `json:"reportsTo,omitempty"` // ID of this user's direct manager, if any.
	Role        string  `json:"role,omitempty"`      // provider's own org role/level name, not a Team360 hierarchy level ID.
}

// Team represents a team from an external organization data source.
type Team struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	HealthCheckEnabled *bool  `json:"healthCheckEnabled,omitempty"` // tri-state: true=enabled, false=disabled, nil=not supplied.
}

// Membership associates a user with a team.
type Membership struct {
	UserID string `json:"userId"`
	TeamID string `json:"teamId"`
}

// Snapshot is a complete provider organization snapshot.
type Snapshot struct {
	ContractVersion string       `json:"contractVersion"`
	GeneratedAt     time.Time    `json:"generatedAt"`
	Teams           []Team       `json:"teams"`
	Users           []User       `json:"users"`
	Memberships     []Membership `json:"memberships"`
	OwnedFields     []OwnedField `json:"ownedFields,omitempty"` // fields the provider claims authority over.
}
