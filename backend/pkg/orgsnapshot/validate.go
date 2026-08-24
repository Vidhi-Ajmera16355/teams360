package orgsnapshot

import "fmt"

// EntityType identifies the record type referenced by a ValidationError.
type EntityType string

const (
	EntityUser                  EntityType = "user"
	EntityTeam                  EntityType = "team"
	EntityMembership            EntityType = "membership"
	EntityReportingRelationship EntityType = "reportingRelationship"
	EntitySnapshot              EntityType = "snapshot"
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
	}

	for _, t := range s.Teams {
		if t.ID == "" {
			continue
		}
		if t.ParentTeamID != "" && !teamIDs[t.ParentTeamID] {
			errs = append(errs, ValidationError{
				Entity:     EntityTeam,
				Identifier: t.ID,
				Field:      "parentTeamId",
				Message: fmt.Sprintf(
					"references unknown team id %q",
					t.ParentTeamID,
				),
			})
		}
	}

	for _, m := range s.Memberships {
		identifier := fmt.Sprintf("%s->%s", m.UserID, m.TeamID)

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

	for _, r := range s.ReportingRelationships {
		identifier := fmt.Sprintf("%s->%s", r.ManagerID, r.ReportID)

		if !userIDs[r.ManagerID] {
			errs = append(errs, ValidationError{
				Entity:     EntityReportingRelationship,
				Identifier: identifier,
				Field:      "managerId",
				Message: fmt.Sprintf(
					"references unknown user id %q",
					r.ManagerID,
				),
			})
		}

		if !userIDs[r.ReportID] {
			errs = append(errs, ValidationError{
				Entity:     EntityReportingRelationship,
				Identifier: identifier,
				Field:      "reportId",
				Message: fmt.Sprintf(
					"references unknown user id %q",
					r.ReportID,
				),
			})
		}
	}

	for _, field := range s.OwnedFields {
		if !IsAllowedOwnedField(field) {
			errs = append(errs, ValidationError{
				Entity:     EntitySnapshot,
				Identifier: string(field),
				Field:      "ownedFields",
				Message: fmt.Sprintf(
					"provider claims ownership of unsupported field %q",
					field,
				),
			})
		}
	}

	return errs
}
