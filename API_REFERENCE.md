# ligneous-gedcom-lib-api Reference

This service is a thin HTTP shell over `ligneous-gedcom-lib`:
- parse/validate/enrich delegate to library packages
- GEDCOM business logic (including `ASSO`/`RELA`) lives in the library

## Endpoints

- `POST /api/v1/parse`
- `POST /api/v1/validate`
- `POST /api/v1/enrich`
- `POST /api/v1/parse-validate`
- `POST /api/v1/parse-validate-enrich`
- `POST /api/v1/export`

## Multipart Upload Endpoints

The following endpoints expect a multipart form upload with field name `file`:
- `/api/v1/parse`
- `/api/v1/validate`
- `/api/v1/enrich`
- `/api/v1/parse-validate`
- `/api/v1/parse-validate-enrich`

Example:

```bash
curl -sS -X POST \
  -F "file=@/path/to/input.ged" \
  "http://127.0.0.1:8091/api/v1/enrich"
```

## Associates (`ASSO` / `RELA`) Support

When the GEDCOM input contains associates, enrich and pipeline responses include:
- `enriched.Associates[]`

Each associate edge has:
- `owner_xref` (record containing `ASSO`, e.g. `@I1@` or `@F1@`)
- `owner_type` (`INDI` or `FAM`)
- `associate_xref` (target individual xref, e.g. `@I2@`)
- `relationship` (from subordinate `RELA`)
- `source_tag` (`ASSO`)
- `owner_event_type` (optional, set when `ASSO` is under an event like `BIRT`)

### Example: `/api/v1/enrich` response fragment

```json
{
  "enriched": {
    "Individuals": [
      { "xref": "@I1@", "full_name": "John /Doe/" },
      { "xref": "@I2@", "full_name": "Jane /Smith/" }
    ],
    "Associates": [
      {
        "owner_xref": "@I1@",
        "owner_type": "INDI",
        "associate_xref": "@I2@",
        "relationship": "Neighbor",
        "source_tag": "ASSO"
      }
    ]
  }
}
```

## Export Endpoint

`POST /api/v1/export` accepts JSON:

```json
{
  "format": "gedcom",
  "filename": "export",
  "enriched": {}
}
```

`format` options:
- `gedcom`
- `json`
- `csv`

If `enriched.Associates` is populated, GEDCOM export emits association lines:
- `1 ASSO @I2@`
- `2 RELA Neighbor`

JSON export includes:
- `individuals[].associates[]`
- `families[].associates[]`

## Notes for Consumers

- Field casing in enriched payloads follows Go struct JSON tags from the library.
- If `generateIds=true` is set on enrich/pipeline endpoints, UUIDs are added to entities/edges, including associates.
