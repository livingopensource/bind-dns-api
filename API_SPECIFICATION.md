# BIND DNS API - Complete API Specification

**Version:** 1.0.0  
**Base URL:** `http://localhost:8080/api/v1`  
**Content-Type:** `application/json`

---

## Table of Contents

1. [Overview](#overview)
2. [Authentication](#authentication)
3. [Response Format](#response-format)
4. [Error Codes](#error-codes)
5. [Endpoints](#endpoints)
   - [Health Check](#health-check)
   - [Domains](#domains)
   - [DNS Records](#dns-records)
   - [Zone Management](#zone-management)
6. [Data Models](#data-models)
7. [Examples](#examples)

---

## Overview

This API provides RESTful endpoints for managing DNS domains and records using BIND DNS zone files. All operations are persisted to zone files in the configured directory.

**Base URL:** `/api/v1`

---

## Authentication

Currently, the API does not require authentication. For production use, implement authentication middleware.

---

## Response Format

All responses follow this standard format:

### Success Response
```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": { ... }
}
```

### Error Response
```json
{
  "success": false,
  "error": "Error description"
}
```

---

## Error Codes

| HTTP Status | Meaning |
|-------------|---------|
| 200 | OK - Request succeeded |
| 201 | Created - Resource created successfully |
| 400 | Bad Request - Invalid request body |
| 404 | Not Found - Resource does not exist |
| 409 | Conflict - Resource already exists |
| 500 | Internal Server Error - Server error |

---

## Endpoints

### Health Check

#### GET /health

Returns the health status of the API.

**Request:**
```
GET /api/v1/health
```

**Response (200 OK):**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0.0"
}
```

---

### Domains

#### GET /domains

Lists all managed domains.

**Request:**
```
GET /api/v1/domains
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": [
    "example.com",
    "test.org",
    "mydomain.net"
  ]
}
```

---

#### GET /domains/:name

Retrieves detailed information about a specific domain.

**Request:**
```
GET /api/v1/domains/:name
```

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| name | string | Yes | Domain name (e.g., `example.com`) |

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "name": "example.com",
    "type": "master",
    "file": "./zones/example.com.zone",
    "soa": {
      "mname": "ns1.example.com.",
      "rname": "admin.example.com.",
      "serial": 1705312200,
      "refresh": 7200,
      "retry": 3600,
      "expire": 1209600,
      "minimum": 86400
    },
    "nameservers": [
      "ns1.example.com.",
      "ns2.example.com."
    ],
    "records": [
      {
        "id": "rec_1705312200123456789",
        "name": "@",
        "type": "A",
        "value": "127.0.0.1",
        "ttl": 3600,
        "created_at": "2024-01-15T10:30:00Z",
        "updated_at": "2024-01-15T10:30:00Z"
      }
    ],
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

**Response (404 Not Found):**
```json
{
  "success": false,
  "error": "failed to read zone file: open ./zones/example.com.zone: no such file or directory"
}
```

---

#### POST /domains

Creates a new domain/zone.

**Request:**
```
POST /api/v1/domains
Content-Type: application/json
```

**Body:**
```json
{
  "name": "example.com",
  "type": "master",
  "nameservers": [
    "ns1.example.com.",
    "ns2.example.com."
  ],
  "soa": {
    "mname": "ns1.example.com.",
    "rname": "admin.example.com.",
    "serial": 0,
    "refresh": 0,
    "retry": 0,
    "expire": 0,
    "minimum": 0
  }
}
```

**Body Parameters:**
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| name | string | Yes | - | Domain name (must be valid DNS name) |
| type | string | No | `master` | Zone type (`master` or `slave`) |
| nameservers | string[] | No | `[ns1.<domain>.]` | List of authoritative nameservers |
| soa | SOA object | No | Auto-generated | SOA record configuration |

**SOA Object:**
| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| mname | string | No | `ns1.<domain>.` | Primary nameserver |
| rname | string | No | `admin.<domain>.` | Responsible email (with `.` instead of `@`) |
| serial | integer | No | Current timestamp | Zone serial number |
| refresh | integer | No | 7200 | Refresh interval (seconds) |
| retry | integer | No | 3600 | Retry interval (seconds) |
| expire | integer | No | 1209600 | Expire time (seconds) |
| minimum | integer | No | 86400 | Minimum TTL (seconds) |

**Response (201 Created):**
```json
{
  "success": true,
  "message": "Domain created successfully"
}
```

**Response (409 Conflict):**
```json
{
  "success": false,
  "error": "domain example.com already exists"
}
```

---

#### PUT /domains/:name

Updates an existing domain's configuration.

**Request:**
```
PUT /api/v1/domains/:name
Content-Type: application/json
```

**Body:** Same as POST /domains

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Domain updated successfully"
}
```

**Response (404 Not Found):**
```json
{
  "success": false,
  "error": "domain example.com does not exist"
}
```

---

#### DELETE /domains/:name

Deletes a domain and its zone file.

**Request:**
```
DELETE /api/v1/domains/:name
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Domain deleted successfully"
}
```

**Response (404 Not Found):**
```json
{
  "success": false,
  "error": "domain example.com does not exist"
}
```

---

### DNS Records

#### GET /domains/:name/records

Lists all DNS records for a domain.

**Request:**
```
GET /api/v1/domains/:name/records
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": [
    {
      "id": "rec_1705312200123456789",
      "name": "@",
      "type": "A",
      "value": "192.168.1.1",
      "ttl": 3600,
      "priority": 0,
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    },
    {
      "id": "rec_1705312200987654321",
      "name": "@",
      "type": "MX",
      "value": "mail.example.com.",
      "ttl": 3600,
      "priority": 10,
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ]
}
```

---

#### POST /domains/:name/records

Adds a new DNS record to a domain.

**Request:**
```
POST /api/v1/domains/:name/records
Content-Type: application/json
```

**Body:**
```json
{
  "name": "www",
  "type": "A",
  "value": "192.168.1.100",
  "ttl": 3600,
  "priority": 0
}
```

**Body Parameters:**
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| name | string | Yes | - | Record name (`@` for domain root) |
| type | string | Yes | - | Record type (A, AAAA, CNAME, MX, TXT, NS, PTR, SRV) |
| value | string | Yes | - | Record value/data |
| ttl | integer | No | 3600 | Time to live in seconds |
| priority | integer | No | 0 | Priority (for MX and SRV records) |

**Supported Record Types:**
| Type | Value Format | Example |
|------|--------------|---------|
| A | IPv4 address | `192.168.1.1` |
| AAAA | IPv6 address | `2001:db8::1` |
| CNAME | Domain name | `www.example.com.` |
| MX | Domain name | `mail.example.com.` |
| TXT | Text string | `"v=spf1 include:_spf.google.com ~all"` |
| NS | Domain name | `ns1.example.com.` |
| PTR | Domain name | `example.com.` |
| SRV | Priority weight port target | `10 5 5060 sip.example.com.` |

**Response (201 Created):**
```json
{
  "success": true,
  "message": "Record added successfully"
}
```

---

#### PUT /domains/:name/records/:recordName/:recordType

Updates an existing DNS record.

**Request:**
```
PUT /api/v1/domains/:name/records/:recordName/:recordType
Content-Type: application/json
```

**Path Parameters:**
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| name | string | Yes | Domain name |
| recordName | string | Yes | Record name to update |
| recordType | string | Yes | Record type (A, AAAA, CNAME, MX, etc.) |

**Body:**
```json
{
  "name": "www",
  "type": "A",
  "value": "192.168.1.200",
  "ttl": 7200,
  "priority": 0
}
```

**Body Parameters:** Same as POST (all optional)

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Record updated successfully"
}
```

**Response (404 Not Found):**
```json
{
  "success": false,
  "error": "record www of type A not found"
}
```

---

#### DELETE /domains/:name/records/:recordName/:recordType

Deletes a DNS record.

**Request:**
```
DELETE /api/v1/domains/:name/records/:recordName/:recordType
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Record deleted successfully"
}
```

**Response (404 Not Found):**
```json
{
  "success": false,
  "error": "record www of type A not found"
}
```

---

### Zone Management

#### POST /domains/:name/reload

Reloads a specific zone using BIND's `rndc`.

**Request:**
```
POST /api/v1/domains/:name/reload
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Zone reloaded successfully"
}
```

**Response (500 Internal Server Error):**
```json
{
  "success": false,
  "error": "rndc reload failed: exit status 1, output: rndc: connect failed: connection refused"
}
```

---

#### POST /reload

Reloads all zones using BIND's `rndc`.

**Request:**
```
POST /api/v1/reload
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "All zones reloaded successfully"
}
```

---

## Data Models

### Domain
```json
{
  "name": "string",
  "type": "master|slave",
  "file": "string",
  "soa": "SOARecord",
  "nameservers": ["string"],
  "records": ["DNSRecord"],
  "created_at": "ISO8601 timestamp",
  "updated_at": "ISO8601 timestamp"
}
```

### DNSRecord
```json
{
  "id": "string",
  "name": "string",
  "type": "A|AAAA|CNAME|MX|TXT|NS|PTR|SRV",
  "value": "string",
  "ttl": "integer",
  "priority": "integer",
  "created_at": "ISO8601 timestamp",
  "updated_at": "ISO8601 timestamp"
}
```

### SOARecord
```json
{
  "mname": "string",
  "rname": "string",
  "serial": "integer",
  "refresh": "integer",
  "retry": "integer",
  "expire": "integer",
  "minimum": "integer"
}
```

### APIResponse
```json
{
  "success": "boolean",
  "message": "string (optional)",
  "data": "any (optional)",
  "error": "string (optional)"
}
```

---

## Examples

### Complete Workflow

#### 1. Create a domain
```bash
curl -X POST http://localhost:8080/api/v1/domains \
  -H "Content-Type: application/json" \
  -d '{
    "name": "example.com",
    "nameservers": ["ns1.example.com.", "ns2.example.com."]
  }'
```

#### 2. Add an A record
```bash
curl -X POST http://localhost:8080/api/v1/domains/example.com/records \
  -H "Content-Type: application/json" \
  -d '{
    "name": "api",
    "type": "A",
    "value": "192.168.1.100",
    "ttl": 3600
  }'
```

#### 3. Add an MX record
```bash
curl -X POST http://localhost:8080/api/v1/domains/example.com/records \
  -H "Content-Type: application/json" \
  -d '{
    "name": "@",
    "type": "MX",
    "value": "mail.example.com.",
    "priority": 10,
    "ttl": 3600
  }'
```

#### 4. Add a TXT record (SPF)
```bash
curl -X POST http://localhost:8080/api/v1/domains/example.com/records \
  -H "Content-Type: application/json" \
  -d '{
    "name": "@",
    "type": "TXT",
    "value": "\"v=spf1 include:_spf.google.com ~all\"",
    "ttl": 3600
  }'
```

#### 5. List all records
```bash
curl http://localhost:8080/api/v1/domains/example.com/records
```

#### 6. Update a record
```bash
curl -X PUT http://localhost:8080/api/v1/domains/example.com/records/api/A \
  -H "Content-Type: application/json" \
  -d '{
    "value": "192.168.1.200",
    "ttl": 7200
  }'
```

#### 7. Reload the zone
```bash
curl -X POST http://localhost:8080/api/v1/domains/example.com/reload
```

#### 8. Delete a record
```bash
curl -X DELETE http://localhost:8080/api/v1/domains/example.com/records/api/A
```

#### 9. Get domain details
```bash
curl http://localhost:8080/api/v1/domains/example.com
```

#### 10. Delete the domain
```bash
curl -X DELETE http://localhost:8080/api/v1/domains/example.com
```

---

## Rate Limiting

Currently, no rate limiting is implemented. For production use, consider adding rate limiting middleware.

---

## CORS

Currently, no CORS headers are set. For production use with browser clients, configure CORS in the API handler.

---

## Changelog

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2024-01-15 | Initial release |
