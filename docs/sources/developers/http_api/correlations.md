---
aliases:
  - /docs/grafana/latest/developers/http_api/correlations/
  - /docs/grafana/latest/http_api/correlations/
description: Grafana Correlations HTTP API
keywords:
  - grafana
  - http
  - documentation
  - api
  - correlations
  - Glue
title: 'Correlations HTTP API '
---

# Correlations API

This API can be used to define correlations between data sources.

## Create correlations

`POST /api/datasources/uid/:sourceUid/correlations`

Creates a correlation between two data sources - the source data source indicated by the path UID, and the target data source which is specified in the body.

**Example request:**

```http
POST /api/datasources/uid/uyBf2637k/correlations HTTP/1.1
Accept: application/json
Content-Type: application/json
Authorization: Bearer eyJrIjoiT0tTcG1pUlY2RnVKZTFVaDFsNFZXdE9ZWmNrMkZYbk
{
	"targetUid": "PDDA8E780A17E7EF1",
	"label": "My Label",
	"description": "Logs to Traces",
}
```

JSON body schema:

- **targetUid** – Target data source uid.
- **label** – A label for the correlation.
- **description** – A description for the correlation.

**Example response:**

```http
HTTP/1.1 200
Content-Type: application/json
{
  "message": "Correlation created",
  "result": {
    "description": "Logs to Traces",
    "label": "My Label",
    "sourceUid": "uyBf2637k",
    "targetUid": "PDDA8E780A17E7EF1",
    "uid": "50xhMlg9k"
  }
}
```

Status codes:

- **200** – OK
- **400** - Errors (invalid JSON, missing or invalid fields)
- **401** – Unauthorized
- **403** – Forbidden, source data source is read-only
- **404** – Not found, either source or target data source could not be found
- **500** – Internal error

## Delete correlations

`DELETE /api/datasources/uid/:uid/correlations/:correlationUid`

Deletes a correlation.

**Example request:**

```http
DELETE /api/datasources/uid/uyBf2637k/correlations/J6gn7d31L HTTP/1.1
Accept: application/json
Content-Type: application/json
Authorization: Bearer eyJrIjoiT0tTcG1pUlY2RnVKZTFVaDFsNFZXdE9ZWmNrMkZYbk
```

**Example response:**

```http
HTTP/1.1 200
Content-Type: application/json
{
  "message": "Correlation deleted"
}
```

Status codes:

- **200** – OK
- **401** – Unauthorized
- **403** – Forbidden, data source is read-only
- **404** – Correlation not found
- **500** – Internal error

## Update correlations

`POST /api/datasources/uid/:uid/correlations/:correlationUis`

Updates a correlation.

**Example request:**

```http
POST /api/datasources/uid/uyBf2637k/correlations/J6gn7d31L HTTP/1.1
Accept: application/json
Content-Type: application/json
Authorization: Bearer eyJrIjoiT0tTcG1pUlY2RnVKZTFVaDFsNFZXdE9ZWmNrMkZYbk
{
	"label": "My Label",
	"description": "Logs to Traces",
}
```

JSON body schema:

- **label** – A label for the correlation.
- **description** – A description for the correlation.

**Example response:**

```http
HTTP/1.1 200
Content-Type: application/json
{
  "message": "Correlation updated",
  "result": {
    "description": "Logs to Traces",
    "label": "My Label",
    "sourceUid": "uyBf2637k",
    "targetUid": "PDDA8E780A17E7EF1",
    "uid": "J6gn7d31L"
  }
}
```

Status codes:

- **200** – OK
- **401** – Unauthorized
- **403** – Forbidden, source data source is read-only
- **404** – Not found, either source or target data source could not be found
- **500** – Internal error

## Get single correlation

`POST /api/datasources/uid/:uid/correlations/:correlationUid`

Gets a single correlation.

**Example request:**

```http
GET /api/datasources/uid/uyBf2637k/correlations/J6gn7d31L HTTP/1.1
Accept: application/json
Authorization: Bearer eyJrIjoiT0tTcG1pUlY2RnVKZTFVaDFsNFZXdE9ZWmNrMkZYbk
```

**Example response:**

```http
HTTP/1.1 200
Content-Type: application/json
{
  "description": "Logs to Traces",
  "label": "My Label",
  "sourceUid": "uyBf2637k",
  "targetUid": "PDDA8E780A17E7EF1",
  "uid": "J6gn7d31L"
}
```

Status codes:

- **200** – OK
- **401** – Unauthorized
- **404** – Not found, either source data source or correlation were not found
- **500** – Internal error

## Get all correlations originating from a given data source

`POST /api/datasources/uid/:uid/correlations`

Get all correlations originating from the data source identified by the given UID in the path.

**Example request:**

```http
GET /api/datasources/uid/uyBf2637k/correlations HTTP/1.1
Accept: application/json
Authorization: Bearer eyJrIjoiT0tTcG1pUlY2RnVKZTFVaDFsNFZXdE9ZWmNrMkZYbk
```

**Example response:**

```http
HTTP/1.1 200
Content-Type: application/json
[
  {
    "description": "Logs to Traces",
    "label": "My Label",
    "sourceUid": "uyBf2637k",
    "targetUid": "PDDA8E780A17E7EF1",
    "uid": "J6gn7d31L"
  },
  {
    "description": "Logs to Metrics",
    "label": "Another Label",
    "sourceUid": "uyBf2637k",
    "targetUid": "P15396BDD62B2BE29",
    "uid": "uWCpURgVk"
  }
]
```

Status codes:

- **200** – OK
- **401** – Unauthorized
- **404** – Not found, either source data source is not found or no correlation exists for the given data source
- **500** – Internal error

## Get all correlations

`POST /api/datasources/correlations`

Get all correlations.

**Example request:**

```http
GET /api/datasources/correlations HTTP/1.1
Accept: application/json
Authorization: Bearer eyJrIjoiT0tTcG1pUlY2RnVKZTFVaDFsNFZXdE9ZWmNrMkZYbk
```

**Example response:**

```http
HTTP/1.1 200
Content-Type: application/json
[
  {
    "description": "Prometheus to Loki",
    "label": "My Label",
    "sourceUid": "uyBf2637k",
    "targetUid": "PDDA8E780A17E7EF1",
    "uid": "J6gn7d31L"
  },
  {
    "description": "Loki to Tempo",
    "label": "Another Label",
    "sourceUid": "PDDA8E780A17E7EF1",
    "targetUid": "P15396BDD62B2BE29",
    "uid": "uWCpURgVk"
  }
]
```

Status codes:

- **200** – OK
- **401** – Unauthorized
- **404** – Not found, no correlations is found
- **500** – Internal error
