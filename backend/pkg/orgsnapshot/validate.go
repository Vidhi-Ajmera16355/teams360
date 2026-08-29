package orgsnapshot

import "fmt"

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
		} else if userIDs[u.ID] {
			errs = append(errs, ValidationError{
				Entity:     EntityUser,
				Identifier: u.ID,
				Field:      "id",
				Message:    "duplicate id among users",
			})
		} else {
			userIDs[u.ID] = true
		}
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
		if u.HierarchyLevelID == "" {
			errs = append(errs, ValidationError{
				Entity:     EntityUser,
				Identifier: u.ID,
				Field:      "hierarchyLevelId",
				Message:    "hierarchyLevelId is required and must be non-empty",
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
		} else if teamIDs[t.ID] {
			errs = append(errs, ValidationError{
				Entity:     EntityTeam,
				Identifier: t.ID,
				Field:      "id",
				Message:    "duplicate id among teams",
			})
		} else {
			teamIDs[t.ID] = true
		}
		if t.Name == "" {
			errs = append(errs, ValidationError{
				Entity:     EntityTeam,
				Identifier: t.ID,
				Field:      "name",
				Message:    "name is required and must be non-empty",
			})
		}
	}

	type membershipKey struct {
		userID string
		teamID string
	}
	membershipIDs := make(map[membershipKey]bool, len(s.Memberships))
	for _, m := range s.Memberships {
		identifier := fmt.Sprintf("%s->%s", m.UserID, m.TeamID)
		key := membershipKey{userID: m.UserID, teamID: m.TeamID}
		if membershipIDs[key] {
			errs = append(errs, ValidationError{
				Entity:     EntityMembership,
				Identifier: identifier,
				Field:      "userId,teamId",
				Message:    "duplicate membership",
			})
		} else {
			membershipIDs[key] = true
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
		leadID := *t.TeamLeadID
		if leadID == "" || !userIDs[leadID] {
			errs = append(errs, ValidationError{
				Entity:     EntityTeam,
				Identifier: t.ID,
				Field:      "teamLeadId",
				Message:    fmt.Sprintf("references unknown user id %q", leadID),
			})
			continue
		}
		if !membershipIDs[membershipKey{userID: leadID, teamID: t.ID}] {
			errs = append(errs, ValidationError{
				Entity:     EntityTeam,
				Identifier: t.ID,
				Field:      "teamLeadId",
				Message:    fmt.Sprintf("team lead %q must be a member of the team", leadID),
			})
		}
	}

	reportsTo := make(map[string]string, len(s.Users))
	for _, u := range s.Users {
		if u.ID != "" && u.ReportsToID != nil && *u.ReportsToID != "" && userIDs[*u.ReportsToID] {
			reportsTo[u.ID] = *u.ReportsToID
		}
	}
	for userID := range reportsTo {
		seen := make(map[string]bool)
		for current := userID; current != ""; current = reportsTo[current] {
			if seen[current] {
				errs = append(errs, ValidationError{
					Entity:     EntityUser,
					Identifier: userID,
					Field:      "reportsToId",
					Message:    "reporting hierarchy contains a cycle",
				})
				break
			}
			seen[current] = true
		}
	}

	return errs
}
