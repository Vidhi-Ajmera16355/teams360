package orgsnapshot_test

import (
	"strings"
	"testing"
	"time"

	"github.com/agopalakrishnan/teams360/backend/pkg/orgsnapshot"
)

func boolPtr(b bool) *bool { return &b }

func validSnapshot() orgsnapshot.Snapshot {
	return orgsnapshot.Snapshot{
		ContractVersion: orgsnapshot.ContractVersion,
		GeneratedAt:     time.Unix(0, 0).UTC(),
		Teams: []orgsnapshot.Team{
			{ID: "team-1", Name: "Platform"},
			{ID: "team-2", Name: "Growth", ParentTeamID: "team-1", HealthCheckEnabled: boolPtr(true)},
			{ID: "team-3", Name: "Data", HealthCheckEnabled: boolPtr(false)},
			{ID: "team-4", Name: "Infra"}, // HealthCheckEnabled omitted (nil)
		},
		Users: []orgsnapshot.User{
			{ID: "user-1", DisplayName: "Alice", Email: "alice@example.com"},
			{ID: "user-2", DisplayName: "Bob"},
		},
		Memberships: []orgsnapshot.Membership{
			{UserID: "user-1", TeamID: "team-1"},
			{UserID: "user-2", TeamID: "team-2"},
		},
		ReportingRelationships: []orgsnapshot.ReportingRelationship{
			{ManagerID: "user-1", ReportID: "user-2"},
		},
		OwnedFields: []orgsnapshot.OwnedField{orgsnapshot.FieldTeamHealthCheckEnabled},
	}
}

func TestValidate_HappyPath(t *testing.T) {
	errs := validSnapshot().Validate()
	if len(errs) != 0 {
		t.Fatalf("expected no validation errors, got %+v", errs)
	}
}

func TestValidate_HealthCheckEnabledTriState(t *testing.T) {
	snap := validSnapshot()

	var enabled, disabled, omitted *bool
	for i := range snap.Teams {
		switch snap.Teams[i].ID {
		case "team-2":
			enabled = snap.Teams[i].HealthCheckEnabled
		case "team-3":
			disabled = snap.Teams[i].HealthCheckEnabled
		case "team-4":
			omitted = snap.Teams[i].HealthCheckEnabled
		}
	}

	if enabled == nil || *enabled != true {
		t.Fatalf("expected team-2 HealthCheckEnabled to be explicit true, got %v", enabled)
	}
	if disabled == nil || *disabled != false {
		t.Fatalf("expected team-3 HealthCheckEnabled to be explicit false, got %v", disabled)
	}
	if omitted != nil {
		t.Fatalf("expected team-4 HealthCheckEnabled to be nil (omitted), got %v", *omitted)
	}
}

func TestValidate_MissingUserID(t *testing.T) {
	snap := validSnapshot()
	snap.Users = append(snap.Users, orgsnapshot.User{DisplayName: "No ID"})

	errs := snap.Validate()
	if !hasError(errs, orgsnapshot.EntityUser, "id", "required") {
		t.Fatalf("expected a missing-id user error, got %+v", errs)
	}
}

func TestValidate_MissingTeamID(t *testing.T) {
	snap := validSnapshot()
	snap.Teams = append(snap.Teams, orgsnapshot.Team{Name: "No ID"})

	errs := snap.Validate()
	if !hasError(errs, orgsnapshot.EntityTeam, "id", "required") {
		t.Fatalf("expected a missing-id team error, got %+v", errs)
	}
}

func TestValidate_DuplicateUserID(t *testing.T) {
	snap := validSnapshot()
	snap.Users = append(snap.Users, orgsnapshot.User{ID: "user-1", DisplayName: "Duplicate Alice"})

	errs := snap.Validate()
	if !hasError(errs, orgsnapshot.EntityUser, "id", "duplicate") {
		t.Fatalf("expected a duplicate user id error, got %+v", errs)
	}
}

func TestValidate_DuplicateTeamID(t *testing.T) {
	snap := validSnapshot()
	snap.Teams = append(snap.Teams, orgsnapshot.Team{ID: "team-1", Name: "Duplicate Platform"})

	errs := snap.Validate()
	if !hasError(errs, orgsnapshot.EntityTeam, "id", "duplicate") {
		t.Fatalf("expected a duplicate team id error, got %+v", errs)
	}
}

func TestValidate_MembershipUnknownUser(t *testing.T) {
	snap := validSnapshot()
	snap.Memberships = append(snap.Memberships, orgsnapshot.Membership{UserID: "ghost-user", TeamID: "team-1"})

	errs := snap.Validate()
	if !hasError(errs, orgsnapshot.EntityMembership, "userId", "unknown user") {
		t.Fatalf("expected a membership unknown-user error, got %+v", errs)
	}
}

func TestValidate_MembershipUnknownTeam(t *testing.T) {
	snap := validSnapshot()
	snap.Memberships = append(snap.Memberships, orgsnapshot.Membership{UserID: "user-1", TeamID: "ghost-team"})

	errs := snap.Validate()
	if !hasError(errs, orgsnapshot.EntityMembership, "teamId", "unknown team") {
		t.Fatalf("expected a membership unknown-team error, got %+v", errs)
	}
}

func TestValidate_ReportingRelationshipUnknownEndpoints(t *testing.T) {
	snap := validSnapshot()
	snap.ReportingRelationships = append(snap.ReportingRelationships, orgsnapshot.ReportingRelationship{
		ManagerID: "ghost-manager",
		ReportID:  "ghost-report",
	})

	errs := snap.Validate()
	if !hasError(errs, orgsnapshot.EntityReportingRelationship, "managerId", "unknown user") {
		t.Fatalf("expected a reporting-relationship unknown-manager error, got %+v", errs)
	}
	if !hasError(errs, orgsnapshot.EntityReportingRelationship, "reportId", "unknown user") {
		t.Fatalf("expected a reporting-relationship unknown-report error, got %+v", errs)
	}
}

func TestValidate_TeamParentUnknown(t *testing.T) {
	snap := validSnapshot()
	snap.Teams = append(snap.Teams, orgsnapshot.Team{ID: "team-5", Name: "Orphan", ParentTeamID: "ghost-team"})

	errs := snap.Validate()
	if !hasError(errs, orgsnapshot.EntityTeam, "parentTeamId", "unknown team") {
		t.Fatalf("expected a team parent-reference error, got %+v", errs)
	}
}

func TestValidate_UnsupportedOwnedField(t *testing.T) {
	snap := validSnapshot()
	snap.OwnedFields = append(snap.OwnedFields, orgsnapshot.OwnedField("team.someUnsupportedField"))

	errs := snap.Validate()
	if !hasError(errs, orgsnapshot.EntitySnapshot, "ownedFields", "unsupported field") {
		t.Fatalf("expected an unsupported-owned-field error, got %+v", errs)
	}
}

func TestValidate_RecordLevelErrorsIdentifyEntityAndField(t *testing.T) {
	snap := orgsnapshot.Snapshot{
		ContractVersion: orgsnapshot.ContractVersion,
		Users:           []orgsnapshot.User{{ID: ""}},
	}

	errs := snap.Validate()
	if len(errs) == 0 {
		t.Fatalf("expected at least one error")
	}
	e := errs[0]
	if e.Entity == "" || e.Field == "" || e.Message == "" {
		t.Fatalf("expected error to identify entity, field, and message, got %+v", e)
	}
}

func hasError(errs []orgsnapshot.ValidationError, entity orgsnapshot.EntityType, field, messageContains string) bool {
	for _, e := range errs {
		if e.Entity == entity && e.Field == field && strings.Contains(e.Message, messageContains) {
			return true
		}
	}
	return false
}
