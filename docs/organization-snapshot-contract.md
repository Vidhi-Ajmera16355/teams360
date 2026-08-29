# Organization Snapshot Contract v1.0

The organization snapshot is a complete, provider-neutral representation of externally managed users, teams, memberships, and reporting relationships.

```json
{
  "contractVersion": "1.0",
  "generatedAt": "2026-08-29T12:00:00Z",
  "users": [
    {
      "id": "user-1",
      "username": "alice",
      "displayName": "Alice",
      "email": "alice@example.com",
      "reportsToId": null,
      "hierarchyLevel": "manager"
    }
  ],
  "teams": [
    {
      "id": "team-1",
      "name": "Platform",
      "teamLeadId": "user-1",
      "healthCheckEnabled": true
    }
  ],
  "memberships": [
    { "userId": "user-1", "teamId": "team-1" }
  ]
}
```

## Semantics

- Provider user and team IDs are canonical IDs for externally managed records. They must be non-empty, stable, unique within their entity type, and never reused.
- `username`, `displayName`, `email`, `hierarchyLevel`, team `name`, `contractVersion`, and `generatedAt` are required.
- `reportsToId` is required in JSON. Use `null` for a root user; otherwise it must reference a user in the snapshot.
- `hierarchyLevel` must be one of `executive`, `director`, `manager`, `teamLead`, or `member`. Team Health Check converts these normalized values to its built-in hierarchy levels.
- `teamLeadId` is optional. When present, it references a user in the snapshot. When omitted, synchronization preserves the existing Team Health Check value.
- `healthCheckEnabled` is optional. Omission preserves the existing Team Health Check value; explicit `true` or `false` replaces it.
- The snapshot is complete for its configured provider. A later reconciliation service may deactivate externally managed records absent from a successfully validated snapshot, but must not alter local records or Team Health Check-owned data.
- Incompatible changes require a new contract version. Consumers of v1.0 reject unsupported versions.
