// Package orgsnapshot defines the provider-neutral organization snapshot contract.
package orgsnapshot

import (
	"encoding/json"
	"errors"
	"time"
)

// ContractVersion is the current snapshot contract version.
const ContractVersion = "1.0"

// User represents a person from an external organization data source.
type User struct {
	ID               string  `json:"id"`
	Username         string  `json:"username"`
	DisplayName      string  `json:"displayName"`
	Email            string  `json:"email"`
	HierarchyLevelID string  `json:"hierarchyLevelId"`
	ReportsToID      *string `json:"reportsToId"` // Nil identifies a root user with no manager.
}

// UnmarshalJSON requires reportsToId to be present while allowing null for root users.
func (u *User) UnmarshalJSON(data []byte) error {
	type userAlias User
	var decoded struct {
		userAlias
		ReportsToID json.RawMessage `json:"reportsToId"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	if decoded.ReportsToID == nil {
		return errors.New("reportsToId is required")
	}
	if string(decoded.ReportsToID) != "null" {
		if err := json.Unmarshal(decoded.ReportsToID, &decoded.userAlias.ReportsToID); err != nil {
			return err
		}
	}
	*u = User(decoded.userAlias)
	return nil
}

// Team represents a team from an external organization data source.
type Team struct {
	ID                 string  `json:"id"`
	Name               string  `json:"name"`
	TeamLeadID         *string `json:"teamLeadId,omitempty"`         // Omitted preserves the existing THC-managed value.
	HealthCheckEnabled *bool   `json:"healthCheckEnabled,omitempty"` // tri-state: true=enabled, false=disabled, nil=not supplied.
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
}
