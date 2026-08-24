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
		if u.ExternalID == "" {
			errs = append(errs, ValidationError{
				Entity:     EntityUser,
				Identifier: fmt.Sprintf("index %d", i),
				Field:      "externalId",
				Message:    "externalId is required and must be non-empty",
			})
			continue
		}
		if userIDs[u.ExternalID] {
			errs = append(errs, ValidationError{
				Entity:     EntityUser,
				Identifier: u.ExternalID,
				Field:      "externalId",
				Message:    "duplicate externalId among users",
			})
			continue
		}
		userIDs[u.ExternalID] = true
	}

	teamIDs := make(map[string]bool, len(s.Teams))
	for i, t := range s.Teams {
		if t.ExternalID == "" {
			errs = append(errs, ValidationError{
				Entity:     EntityTeam,
				Identifier: fmt.Sprintf("index %d", i),
				Field:      "externalId",
				Message:    "externalId is required and must be non-empty",
			})
			continue
		}
		if teamIDs[t.ExternalID] {
			errs = append(errs, ValidationError{
				Entity:     EntityTeam,
				Identifier: t.ExternalID,
				Field:      "externalId",
				Message:    "duplicate externalId among teams",
			})
			continue
		}
		teamIDs[t.ExternalID] = true
	}

	for _, t := range s.Teams {
		if t.ExternalID == "" {
			continue
		}
		if t.ParentTeamExternalID != "" && !teamIDs[t.ParentTeamExternalID] {
			errs = append(errs, ValidationError{
				Entity:     EntityTeam,
				Identifier: t.ExternalID,
				Field:      "parentTeamExternalId",
				Message: fmt.Sprintf(
					"references unknown team externalId %q",
					t.ParentTeamExternalID,
				),
			})
		}
	}

	for _, m := range s.Memberships {
		identifier := fmt.Sprintf("%s->%s", m.UserExternalID, m.TeamExternalID)

		if !userIDs[m.UserExternalID] {
			errs = append(errs, ValidationError{
				Entity:     EntityMembership,
				Identifier: identifier,
				Field:      "userExternalId",
				Message: fmt.Sprintf(
					"references unknown user externalId %q",
					m.UserExternalID,
				),
			})
		}

		if !teamIDs[m.TeamExternalID] {
			errs = append(errs, ValidationError{
				Entity:     EntityMembership,
				Identifier: identifier,
				Field:      "teamExternalId",
				Message: fmt.Sprintf(
					"references unknown team externalId %q",
					m.TeamExternalID,
				),
			})
		}
	}

	for _, r := range s.ReportingRelationships {
		identifier := fmt.Sprintf("%s->%s", r.ManagerExternalID, r.ReportExternalID)

		if !userIDs[r.ManagerExternalID] {
			errs = append(errs, ValidationError{
				Entity:     EntityReportingRelationship,
				Identifier: identifier,
				Field:      "managerExternalId",
				Message: fmt.Sprintf(
					"references unknown user externalId %q",
					r.ManagerExternalID,
				),
			})
		}

		if !userIDs[r.ReportExternalID] {
			errs = append(errs, ValidationError{
				Entity:     EntityReportingRelationship,
				Identifier: identifier,
				Field:      "reportExternalId",
				Message: fmt.Sprintf(
					"references unknown user externalId %q",
					r.ReportExternalID,
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
