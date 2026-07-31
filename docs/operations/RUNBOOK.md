# Operations Runbook: Prospecção Brasil

**Status**: Stub
**Last Updated**: 2026-07-31

> Production deployment is a Non-Goal until a future spec. This file is a
> placeholder with the sections that will be needed when deployment becomes
> a goal. Do not fill in commands until the deployment spec is written.

## Health Checks

> TODO: Define health check endpoints and expected responses.
> Will be populated when the deployment spec (future SPEC-XX) is written.

- [ ] Application health endpoint
- [ ] Database connectivity check
- [ ] Migration version check

## DB Recovery

> TODO: Define database backup and recovery procedures.
> Will be populated when the deployment spec is written.

- [ ] Backup strategy (pg_dump, WAL archiving, or managed snapshot)
- [ ] Restore procedure (point-in-time recovery)
- [ ] Migration rollback procedure (dev only -- production is forward-only)

## Incident Response

> TODO: Define incident response procedures.
> Will be populated when the deployment spec is written.

- [ ] Severity levels and response times
- [ ] On-call rotation
- [ ] Postmortem template
- [ ] LGPD incident reporting procedure (ANPD notification within 48h for
  data breaches affecting personal data)

## References

- [Constitution](../../.specify/memory/constitution.md) -- Principle II
  (Security-First LGPD) governs all incident response.
- [PRD](../planning/PRD.md) -- Stack summary for understanding the system.
- [Architecture Decisions](../architecture/DECISIONS.md) -- ADRs that
  affect operations (e.g., single binary, forward-only migrations).
