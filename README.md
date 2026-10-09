# ligneous-gedcom-lib-api

A small Go HTTP service that parses, validates, enriches, exports and reconciles GEDCOM genealogy files for the Ligneous web apps.

Part of [Ligneous](https://github.com/lesfleursdelanuitdev/Ligneous).

## What it does

The service is a thin HTTP layer over [ligneous-gedcom-lib](https://github.com/lesfleursdelanuitdev/ligneous-gedcom-lib) (Go module `github.com/lesfleursdelanuitdev/ligneous-gedcom-lib`). The GEDCOM logic lives in the library. This repo adds routing, upload handling, JSON responses, CORS and deployment files.

Front ends call it to:

- turn an uploaded `.ged` file into JSON tables they can insert into Postgres
- check a file and show validation findings before import
- export a tree from their database back to GEDCOM, JSON or CSV
- compare two versions of a tree and get a merge plan

It has no database and keeps no state between requests.

## Why Go

The library is written in Go for its parsing strengths, explained in the [library README](https://github.com/lesfleursdelanuitdev/ligneous-gedcom-lib#why-go). For this service, Go gives two practical things:

- **Concurrent requests with no extra code.** Go's `net/http` server handles each request on its own goroutine. The handlers have no shared state (`type Handlers struct{}` in `internal/handlers/handlers.go`), and the library calls only touch their own input, so uploads are processed in parallel without locks. The service does not start goroutines of its own.
- **One static binary.** The Dockerfile builds with `CGO_ENABLED=0` and copies the binary into an `alpine:3.19` image. The same binary also runs directly under systemd.

## The intermediate representation over HTTP

The library produces two JSON-serializable IRs (details in the library's [docs/INTERMEDIATE_FORMAT.md](https://github.com/lesfleursdelanuitdev/ligneous-gedcom-lib/blob/master/docs/INTERMEDIATE_FORMAT.md)):

- `document`: a lossless `GedcomDocument` tree, one node per GEDCOM line
- `enriched`: an `EnrichedDocument`, flat deduplicated tables (people, families, dates, places, events, edges and link tables) that mirror the front end's `gedcom_*_v2` Postgres tables

This service is how that IR reaches the front ends:

```
 browser upload
      |
      v
 front-end server route (Next.js)
      |  multipart "file"
      v
 POST /api/v1/parse-validate-enrich?generateIds=true      <- this service
      |  JSON: document, warnings, validation, enriched, stats
      v
 ligneous-frontend: inserts `enriched` into Postgres (Prisma, one transaction)
 the-gonsalves-family-admin: stores the JSON as a pending import for review

 front-end export
      |  JSON: { format, filename, enriched }   (EnrichedDocument rebuilt from Postgres)
      v
 POST /api/v1/export  ->  .ged / .json / .csv download
```

Which endpoints produce or consume which IR:

| Endpoint | Input | Returns |
|---|---|---|
| `/api/v1/parse` | `.ged` upload | `document`, `warnings` |
| `/api/v1/validate` | `.ged` upload | `valid`, `errors` (all findings), `counts` (errors, warnings, hints) |
| `/api/v1/parse-validate` | `.ged` upload | `document`, `warnings`, `validation` |
| `/api/v1/enrich` | `.ged` upload | `enriched`, `stats` |
| `/api/v1/parse-validate-enrich` | `.ged` upload | `document`, `warnings`, `validation`, `enriched`, `stats` |
| `/api/v1/export` | JSON with `enriched` | GEDCOM, JSON or CSV file |
| `/api/v1/reconcile/merge-plan` | JSON `{ left, right, options? }`, two `EnrichedDocument`s | `mergePlan` |
| `/api/v1/reconcile/session` | same as merge-plan | `mergePlan`, `session` |

Add `?generateIds=true` to `enrich` or `parse-validate-enrich` to get a UUID on every row, ready for database insert.

Trimmed real response from `parse-validate-enrich` for a two-person file:

```json
{
  "warnings": null,
  "validation": {
    "valid": true,
    "errors": [
      { "Severity": 1, "Code": "HEADER_MISSING_SUBM",
        "Message": "GEDCOM 5.5.1 header should include a SUBM pointer to a submitter record",
        "Xref": "", "Details": { "gedcom_version": "5.5.1" } }
    ]
  },
  "enriched": {
    "Individuals": [
      { "xref": "@I1@", "full_name": "John /Doe/", "sex": "M",
        "birth_date_index": 0, "birth_place_index": 0,
        "death_date_index": -1, "death_place_index": -1 }
    ],
    "Dates":  [ { "original": "ABT 1950", "type": "ABOUT", "calendar": "GREGORIAN", "year": 1950 } ],
    "ParentChild": [ { "parent_xref": "@I2@", "child_xref": "@I1@", "family_xref": "@F1@",
                       "parent_type": "mother", "relationship_type": "biological" } ]
  },
  "stats": { "individuals": 2, "families": 1, "dates": 1, "places": 1, "events": 1 }
}
```

Notes for consumers:

- `Severity` is a number: 0 = hint, 1 = warning, 2 = error. `valid` is `false` only when there is at least one error.
- Top-level `enriched` keys are Go field names (`Individuals`, `ParentChild`). Row fields are snake_case.
- Cross-references to dates, places and other lookup rows are slice indexes, with `-1` for none. People and families are referenced by xref.
- Empty tables come back as `null`, not `[]`.

## Endpoints

| Method | Path | Body |
|---|---|---|
| GET | `/health` | none; returns `{"status":"ok"}` |
| POST | `/api/v1/parse` | multipart, field `file` |
| POST | `/api/v1/validate` | multipart, field `file` |
| POST | `/api/v1/enrich` | multipart, field `file` |
| POST | `/api/v1/parse-validate` | multipart, field `file` |
| POST | `/api/v1/parse-validate-enrich` | multipart, field `file` |
| POST | `/api/v1/export` | JSON `{ "format": "gedcom" \| "json" \| "csv", "filename": "...", "enriched": {...} }` |
| POST | `/api/v1/reconcile/merge-plan` | JSON `{ "left": {...}, "right": {...}, "options": {...} }` |
| POST | `/api/v1/reconcile/session` | same as merge-plan |

Limits: multipart uploads are parsed with a 10 MB in-memory limit (`internal/handlers/helpers.go`). Reconcile bodies are capped at 25 MB. All responses carry `Access-Control-Allow-Origin: *`. There is no authentication; the deploy files bind the service to `127.0.0.1` behind nginx.

More detail, including associate (`ASSO`/`RELA`) handling, is in [API_REFERENCE.md](API_REFERENCE.md).

```bash
curl -sS -F "file=@family.ged" "http://127.0.0.1:8091/api/v1/parse-validate-enrich?generateIds=true"
```

## Layout

| Path | Purpose |
|---|---|
| `main.go` | Reads `-port`, `PORT` and `LISTEN_ADDR`, starts the server |
| `internal/server/server.go` | Routes (Go 1.22+ method patterns on `http.ServeMux`) and CORS middleware |
| `internal/handlers/` | One file per endpoint, plus upload and JSON helpers |
| `internal/handlers/associates_test.go` | HTTP tests for associates in enrich, pipeline and GEDCOM export |
| `vendor/` | Vendored copy of `ligneous-gedcom-lib` and `google/uuid`, so the repo builds on its own |
| `Dockerfile`, `docker-compose.yml` | Container build and run on port 8091 |
| `deploy/` | systemd unit, nginx example and deployment notes |

Size: 497 lines of Go outside tests and `vendor/`.

## Build, run and test

Requires Go 1.23.5 or later (`go.mod`). The library is vendored, so build with `-mod=vendor`:

```bash
go build -mod=vendor -o ligneous-gedcom-lib-api .
./ligneous-gedcom-lib-api                       # listens on :8091
LISTEN_ADDR=127.0.0.1:8091 ./ligneous-gedcom-lib-api

go test -mod=vendor ./...
```

The three handler tests pass (checked 2026-10-09 with Go 1.24.4).

`go.mod` also has `replace github.com/lesfleursdelanuitdev/ligneous-gedcom-lib => ../ligneous-gedcom-lib`. To build against a local checkout of the library instead of `vendor/`, clone it next to this repo **into a folder named `ligneous-gedcom-lib`** (the GitHub repo name is spelled `ligneous-gedcom-lib`), then run `go mod vendor` to refresh the vendored copy.

With Docker Compose:

```bash
docker compose up -d
curl -fsS http://127.0.0.1:8091/health
```

| Variable | Meaning |
|---|---|
| `PORT` | Port when `LISTEN_ADDR` is unset (binds `:<PORT>`). Default 8091. |
| `LISTEN_ADDR` | Full bind address such as `127.0.0.1:8091`. Overrides `PORT` and `-port`. |

For systemd plus nginx, see [deploy/README.md](deploy/README.md).

## Related repositories

- [Ligneous](https://github.com/lesfleursdelanuitdev/Ligneous): the umbrella project
- [ligneous-gedcom-lib](https://github.com/lesfleursdelanuitdev/ligneous-gedcom-lib): the Go library this service wraps
- `ligneous-frontend` (private): calls `parse-validate-enrich` on upload and `export` on download
- `the-gonsalves-family-admin` (private): calls `parse-validate-enrich`, `export` and `reconcile/merge-plan`
- `the-gonsalves-family` (private): the public family site managed by that admin
- [ligneous-gedcom-flask-api](https://github.com/lesfleursdelanuitdev/ligneous-gedcom-flask-api): Python research and analytics API over the same Postgres data

## License

No license file: all rights reserved.
