package orgsnapshot

import "fmt"

var supportedHierarchyLevels = map[HierarchyLevel]bool{
	HierarchyLevelExecutive: true,
	HierarchyLevelDirector:  true,
	HierarchyLevelManager:   true,
	HierarchyLevelTeamLead:  true,
	HierarchyLevelMember:    true,
}

// EntityType identifies the record type referenced by a ValidationError.
type EntityType string

const (
	EntityUser       EntityType = "user"
	EntityTeam       EntityType = "team"
	EntityMembership EntityType = "membership"
	EntitySnapshot   EntityType = "snapshot"
)

// ValidationError represents a single validation failure.
type ValidationError struct {
	Entity     EntityType `json:"entity"`
	Identifier string     `json:"identifier"`
	Field      string     `json:"field"`
	Message    string     `json:"message"`
}

func (e ValidationError) Error() string {
	if e.Identifier != "" {
		return fmt.Sprintf("%s %q: %s: %s", e.Entity, e.Identifier, e.Field, e.Message)
	}
	return fmt.Sprintf("%s: %s: %s", e.Entity, e.Field, e.Message)
}

// Validate checks the snapshot against contract-level rules and returns
// all validation errors found.
func (s Snapshot) Validate() []ValidationError {
	var errs []ValidationError
	if s.ContractVersion != ContractVersion {
		errs = append(errs, ValidationError{
			Entity:  EntitySnapshot,
			Field:   "contractVersion",
			Message: fmt.Sprintf("unsupported contract version %q; expected %q", s.ContractVersion, ContractVersion),
		})
	}
	if s.GeneratedAt.IsZero() {
		errs = append(errs, ValidationError{
			Entity:  EntitySnapshot,
			Field:   "generatedAt",
			Message: "generatedAt is required",
		})
	}

	userIDs := make(map[string]bool, len(s.Users))
	for i, u := range s.Users {
		if u.ID == "" {
			errs = append(errs, ValidationError{
				Entity:     EntityUser,
				Identifier: fmt.Sprintf("index %d", i),
				Field:      "id",
				Message:    "id is required and must be non-empty",
			})
			continue
		}
		if userIDs[u.ID] {
			errs = append(errs, ValidationError{
				Entity:     EntityUser,
				Identifier: u.ID,
				Field:      "id",
				Message:    "duplicate id among users",
			})
			continue
		}
		userIDs[u.ID] = true
		if u.Username == "" {
			errs = append(errs, ValidationError{
				Entity:     EntityUser,
				Identifier: u.ID,
				Field:      "username",
				Message:    "username is required and must be non-empty",
			})
		}
		if u.DisplayName == "" {
			errs = append(errs, ValidationError{
				Entity:     EntityUser,
				Identifier: u.ID,
				Field:      "displayName",
				Message:    "displayName is required and must be non-empty",
			})
		}
		if !supportedHierarchyLevels[u.HierarchyLevel] {
			errs = append(errs, ValidationError{
				Entity:     EntityUser,
				Identifier: u.ID,
				Field:      "hierarchyLevel",
				Message:    fmt.Sprintf("unsupported hierarchy level %q", u.HierarchyLevel),
			})
		}
		if u.Email == "" {
			errs = append(errs, ValidationError{
				Entity:     EntityUser,
				Identifier: u.ID,
				Field:      "email",
				Message:    "email is required and must be non-empty",
			})
		}
	}

	teamIDs := make(map[string]bool, len(s.Teams))
	for i, t := range s.Teams {
		if t.ID == "" {
			errs = append(errs, ValidationError{
				Entity:     EntityTeam,
				Identifier: fmt.Sprintf("index %d", i),
				Field:      "id",
				Message:    "id is required and must be non-empty",
			})
			continue
		}
		if teamIDs[t.ID] {
			errs = append(errs, ValidationError{
				Entity:     EntityTeam,
				Identifier: t.ID,
				Field:      "id",
				Message:    "duplicate id among teams",
			})
			continue
		}
		teamIDs[t.ID] = true
		if t.Name == "" {
			errs = append(errs, ValidationError{
				Entity:     EntityTeam,
				Identifier: t.ID,
				Field:      "name",
				Message:    "name is required and must be non-empty",
			})
		}
	}

	membershipIDs := make(map[string]bool, len(s.Memberships))
	for _, m := range s.Memberships {
		identifier := fmt.Sprintf("%s->%s", m.UserID, m.TeamID)
		if membershipIDs[identifier] {
			errs = append(errs, ValidationError{
				Entity:     EntityMembership,
				Identifier: identifier,
				Field:      "userId,teamId",
				Message:    "duplicate membership",
			})
		} else {
			membershipIDs[identifier] = true
		}

		if !userIDs[m.UserID] {
			errs = append(errs, ValidationError{
				Entity:     EntityMembership,
				Identifier: identifier,
				Field:      "userId",
				Message: fmt.Sprintf(
					"references unknown user id %q",
					m.UserID,
				),
			})
		}

		if !teamIDs[m.TeamID] {
			errs = append(errs, ValidationError{
				Entity:     EntityMembership,
				Identifier: identifier,
				Field:      "teamId",
				Message: fmt.Sprintf(
					"references unknown team id %q",
					m.TeamID,
				),
			})
		}
	}

	for _, u := range s.Users {
		if u.ReportsToID == nil {
			continue
		}
		managerID := *u.ReportsToID
		if managerID == "" {
			errs = append(errs, ValidationError{
				Entity:     EntityUser,
				Identifier: u.ID,
				Field:      "reportsToId",
				Message:    "reportsToId must be null or a non-empty user id",
			})
			continue
		}

		if managerID == u.ID {
			errs = append(errs, ValidationError{
				Entity:     EntityUser,
				Identifier: u.ID,
				Field:      "reportsToId",
				Message:    "user cannot be their own manager",
			})
			continue
		}

		if !userIDs[managerID] {
			errs = append(errs, ValidationError{
				Entity:     EntityUser,
				Identifier: u.ID,
				Field:      "reportsToId",
				Message: fmt.Sprintf(
					"references unknown user id %q",
					managerID,
				),
			})
		}
	}

	for _, t := range s.Teams {
		if t.TeamLeadID == nil {
			continue
		}
		if *t.TeamLeadID == "" || !userIDs[*t.TeamLeadID] {
			errs = append(errs, ValidationError{
				Entity:     EntityTeam,
				Identifier: t.ID,
				Field:      "teamLeadId",
				Message:    fmt.Sprintf("references unknown user id %q", *t.TeamLeadID),
			})
		}
	}

	return errs
}
