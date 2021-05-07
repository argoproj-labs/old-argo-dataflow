# Security

## Supply Chain

* For base images prefer `scratch` then `distroless` then `alpine`.
* Snyk is used to scan images.
* Synk is used to scan imported Go modules.

## Configuration

* Step pods `runAsNonRoot: true` with user `9653`.
* Step pods have `automountServiceAccountToken: true`, but the `pipeline` service account has only `get secrects` and `patch steps/status`.