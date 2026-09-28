# Tasks: M6 — Operations, OpenAPI, Checks, and Import

## 1. OpenAPI
- [ ] 1.1 Author `internal/api/openapi.yaml` covering all `/api/v1` endpoints
- [ ] 1.2 Serve it at `/api/v1/openapi.json`
- [ ] 1.3 Generate the TypeScript client into `web/src/lib/api/generated`
- [ ] 1.4 Replace M5's hand-written client with the generated one
- [ ] 1.5 CI check that regeneration produces no diff
- [ ] 1.6 Contract tests asserting handlers match the spec

## 2. TCP and DNS checks
- [ ] 2.1 `TCPChecker` implementing the M2 `Checker` interface
- [ ] 2.2 `DNSChecker` implementing the M2 `Checker` interface
- [ ] 2.3 Validate type-specific required fields
- [ ] 2.4 Unit tests for success, timeout, and refusal per type
- [ ] 2.5 Verify a TCP and a DNS service in the dashboard

## 3. Metrics
- [ ] 3.1 Expose the counters from design.md
- [ ] 3.2 Verify labels and values after a real down/up cycle

## 4. Packaging and docs
- [ ] 4.1 Multi-stage `Dockerfile` with CA certs
- [ ] 4.2 `slack-app-manifest.yaml` at the repo root
- [ ] 4.3 `docs/setup.md` covering Google OAuth, Slack app, env, first admin
- [ ] 4.4 Verify `docker run` yields a working single-binary deployment

## 5. Kuma import
- [ ] 5.1 Parse a Kuma export into a typed intermediate model
- [ ] 5.2 Map groups to projects and monitors to services
- [ ] 5.3 Map Slack notifications to channels and `notificationIDList` to
      service overrides
- [ ] 5.4 Skip unsupported types and list them in the report with reasons
- [ ] 5.5 Ignore `parent` and do not import history
- [ ] 5.6 Encrypt imported secrets and never log them
- [ ] 5.7 `POST /import/kuma/preview` returns the report with no writes
- [ ] 5.8 `POST /import/kuma/apply` applies atomically and returns the report
- [ ] 5.9 Idempotent-safe: applying the same export twice creates no duplicates
- [ ] 5.10 Verify with the sample version 1.23.17 export

## 6. Verification
- [ ] 6.1 Run `go test ./...` and the frontend suite
- [ ] 6.2 Full end-to-end: import → alert → dashboard → delivery view
- [ ] 6.3 Confirm no secret appears in logs or API responses
