package orgsnapshot

// OwnedField identifies a contract field a provider may own.
type OwnedField string
 
const (
	// FieldTeamHealthCheckEnabled identifies Team.HealthCheckEnabled.
	FieldTeamHealthCheckEnabled OwnedField = "team.healthCheckEnabled"
)

// allowedOwnedFields contains the fields providers may declare.
var allowedOwnedFields = map[OwnedField]bool{
	FieldTeamHealthCheckEnabled: true,
}

// IsAllowedOwnedField reports whether field is allowlisted.
func IsAllowedOwnedField(field OwnedField) bool {
	return allowedOwnedFields[field]
}
