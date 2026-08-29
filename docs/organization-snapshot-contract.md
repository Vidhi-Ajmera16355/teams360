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
      "hierarchyLevelId": "level-3"
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
- `username`, `displayName`, `email`, `hierarchyLevelId`, team `name`, `contractVersion`, and `generatedAt` are required.
- `reportsToId` is required in JSON. Use `null` for a root user; otherwise it must reference a user in the snapshot.
- `hierarchyLevelId` must identify a hierarchy level configured in the target Team Health Check deployment. Existence is checked during synchronization because the snapshot does not contain hierarchy-level definitions.
- `teamLeadId` is optional. When present, it must reference a user who is also a member of that team. When omitted or `null`, synchronization preserves the existing Team Health Check value; v1.0 does not define an explicit clear operation.
- `healthCheckEnabled` is optional. Omission preserves the existing Team Health Check value; explicit `true` or `false` replaces it.
- The snapshot is complete for its configured provider. A later reconciliation service deactivates externally managed records absent from a successfully validated snapshot, but does not alter local records or Team Health Check-owned data.
- Synchronization must track provider provenance so externally managed records can be distinguished from local records.
- Incompatible changes require a new contract version. Consumers of v1.0 reject unsupported versions.
