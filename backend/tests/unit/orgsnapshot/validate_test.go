package orgsnapshot_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/agopalakrishnan/teams360/backend/pkg/orgsnapshot"
)

func boolPtr(b bool) *bool       { return &b }
func stringPtr(s string) *string { return &s }

func validSnapshot() orgsnapshot.Snapshot {
	return orgsnapshot.Snapshot{
		ContractVersion: orgsnapshot.ContractVersion,
		GeneratedAt:     time.Unix(1, 0).UTC(),
		Teams: []orgsnapshot.Team{
			{ID: "team-1", Name: "Platform"},
			{ID: "team-2", Name: "Growth", HealthCheckEnabled: boolPtr(true)},
			{ID: "team-3", Name: "Data", HealthCheckEnabled: boolPtr(false)},
			{ID: "team-4", Name: "Infra"}, // HealthCheckEnabled omitted (nil)
		},
		Users: []orgsnapshot.User{
			{ID: "user-1", Username: "alice", DisplayName: "Alice", Email: "alice@example.com", HierarchyLevel: orgsnapshot.HierarchyLevelManager},
			{ID: "user-2", Username: "bob", DisplayName: "Bob", Email: "bob@example.com", ReportsToID: stringPtr("user-1"), HierarchyLevel: orgsnapshot.HierarchyLevelMember},
		},
		Memberships: []orgsnapshot.Membership{
			{UserID: "user-1", TeamID: "team-1"},
			{UserID: "user-2", TeamID: "team-2"},
		},
	}
}

func TestValidate_HappyPath(t *testing.T) {
	errs := validSnapshot().Validate()
	if len(errs) != 0 {
		t.Fatalf("expected no validation errors, got %+v", errs)
	}
}

func TestValidate_HealthCheckEnabledTriState(t *testing.T) {
	data, err := json.Marshal(validSnapshot())
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}
	var snap orgsnapshot.Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		t.Fatalf("unmarshal snapshot: %v", err)
	}

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

func TestUserJSON_RequiresReportsToID(t *testing.T) {
	var user orgsnapshot.User
	if err := json.Unmarshal([]byte(`{"id":"user-1","username":"alice","displayName":"Alice","email":"alice@example.com","hierarchyLevel":"manager"}`), &user); err == nil {
		t.Fatal("expected missing reportsToId to fail JSON decoding")
	}
	if err := json.Unmarshal([]byte(`{"id":"user-1","username":"alice","displayName":"Alice","email":"alice@example.com","hierarchyLevel":"manager","reportsToId":null}`), &user); err != nil {
		t.Fatalf("expected null reportsToId to identify a root user: %v", err)
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

func TestValidate_ReportsToUnknownManager(t *testing.T) {
	snap := validSnapshot()
	snap.Users = append(snap.Users, orgsnapshot.User{ID: "user-3", Username: "cara", DisplayName: "Cara", Email: "cara@example.com", HierarchyLevel: orgsnapshot.HierarchyLevelMember, ReportsToID: stringPtr("ghost-manager")})

	errs := snap.Validate()
	if !hasError(errs, orgsnapshot.EntityUser, "reportsToId", "unknown user") {
		t.Fatalf("expected a reportsTo unknown-manager error, got %+v", errs)
	}
}

func TestValidate_ReportsToSelf(t *testing.T) {
	snap := validSnapshot()
	snap.Users = append(snap.Users, orgsnapshot.User{ID: "user-3", Username: "cara", DisplayName: "Cara", Email: "cara@example.com", HierarchyLevel: orgsnapshot.HierarchyLevelMember, ReportsToID: stringPtr("user-3")})

	errs := snap.Validate()
	if !hasError(errs, orgsnapshot.EntityUser, "reportsToId", "own manager") {
		t.Fatalf("expected a self-manager error, got %+v", errs)
	}
}

func TestValidate_UnsupportedContractVersion(t *testing.T) {
	snap := validSnapshot()
	snap.ContractVersion = "2.0"

	if !hasError(snap.Validate(), orgsnapshot.EntitySnapshot, "contractVersion", "unsupported") {
		t.Fatal("expected unsupported contract version error")
	}
}

func TestValidate_DuplicateMembership(t *testing.T) {
	snap := validSnapshot()
	snap.Memberships = append(snap.Memberships, snap.Memberships[0])

	if !hasError(snap.Validate(), orgsnapshot.EntityMembership, "userId,teamId", "duplicate") {
		t.Fatal("expected duplicate membership error")
	}
}

func TestValidate_UnknownTeamLead(t *testing.T) {
	snap := validSnapshot()
	snap.Teams[0].TeamLeadID = stringPtr("ghost-user")

	if !hasError(snap.Validate(), orgsnapshot.EntityTeam, "teamLeadId", "unknown user") {
		t.Fatal("expected unknown team lead error")
	}
}

func TestValidate_UnsupportedHierarchyLevel(t *testing.T) {
	snap := validSnapshot()
	snap.Users[0].HierarchyLevel = "engineering-manager"

	errs := snap.Validate()
	if !hasError(errs, orgsnapshot.EntityUser, "hierarchyLevel", "unsupported") {
		t.Fatalf("expected an unsupported hierarchy-level error, got %+v", errs)
	}
}

func TestValidate_RequiredUserFields(t *testing.T) {
	tests := []struct {
		name  string
		field string
		clear func(*orgsnapshot.User)
	}{
		{name: "username", field: "username", clear: func(u *orgsnapshot.User) { u.Username = "" }},
		{name: "display name", field: "displayName", clear: func(u *orgsnapshot.User) { u.DisplayName = "" }},
		{name: "email", field: "email", clear: func(u *orgsnapshot.User) { u.Email = "" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snap := validSnapshot()
			tt.clear(&snap.Users[0])
			if !hasError(snap.Validate(), orgsnapshot.EntityUser, tt.field, "required") {
				t.Fatalf("expected required %s error", tt.field)
			}
		})
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
