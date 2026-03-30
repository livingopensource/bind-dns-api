# BIND DNS Domain Management API

A Go-based RESTful API service for managing DNS domains and zones using BIND DNS tools.

## Features

- Create, read, update, and delete DNS domains/zones
- Manage DNS records (A, AAAA, CNAME, MX, TXT, NS, PTR, SRV)
- Zone file generation and parsing
- BIND `rndc` integration for zone reloading
- RESTful JSON API

## Project Structure

```
bind/
├── cmd/
│   └── server/
│       └── main.go          # Application entry point
├── internal/
│   ├── api/
│   │   └── handler.go       # HTTP handlers and routes
│   ├── bind/
│   │   └── manager.go       # BIND integration (zone files, rndc)
│   ├── config/
│   │   └── config.go        # Configuration management
│   └── models/
│       └── models.go        # Data structures
├── zones/                    # Zone files directory
├── config.json              # Configuration file
├── go.mod
└── README.md
```

## Quick Start

### Build

```bash
go build -o bind-dns-api ./cmd/server
```

### Run

```bash
./bind-dns-api -config config.json
```

Or with default configuration:

```bash
go run ./cmd/server/main.go
```

## Configuration

Edit `config.json` to customize:

| Field | Description | Default |
|-------|-------------|---------|
| `server.host` | HTTP server bind address | `0.0.0.0` |
| `server.port` | HTTP server port | `8080` |
| `bind.zone_directory` | Directory for zone files | `./zones` |
| `bind.rndc_path` | Path to rndc binary | `/usr/sbin/rndc` |
| `bind.default_ttl` | Default TTL for records | `3600` |

## API Endpoints

### Health Check

```
GET /api/v1/health
```

### Domains

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/domains` | List all domains |
| GET | `/api/v1/domains/:name` | Get domain details |
| POST | `/api/v1/domains` | Create new domain |
| PUT | `/api/v1/domains/:name` | Update domain |
| DELETE | `/api/v1/domains/:name` | Delete domain |

### DNS Records

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/domains/:name/records` | List all records |
| POST | `/api/v1/domains/:name/records` | Add DNS record |
| PUT | `/api/v1/domains/:name/records/:recordName/:recordType` | Update record |
| DELETE | `/api/v1/domains/:name/records/:recordName/:recordType` | Delete record |

### Zone Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/domains/:name/reload` | Reload specific zone |
| POST | `/api/v1/reload` | Reload all zones |

## Usage Examples

### Create a Domain

```bash
curl -X POST http://localhost:8080/api/v1/domains \
  -H "Content-Type: application/json" \
  -d '{
    "name": "example.com",
    "nameservers": ["ns1.example.com.", "ns2.example.com."],
    "soa": {
      "mname": "ns1.example.com.",
      "rname": "admin.example.com."
    }
  }'
```

### Add a DNS Record

```bash
curl -X POST http://localhost:8080/api/v1/domains/example.com/records \
  -H "Content-Type: application/json" \
  -d '{
    "name": "mail",
    "type": "A",
    "value": "192.168.1.100",
    "ttl": 3600
  }'
```

### Add MX Record

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

### List All Records

```bash
curl http://localhost:8080/api/v1/domains/example.com/records
```

### Reload Zone

```bash
curl -X POST http://localhost:8080/api/v1/domains/example.com/reload
```

## Requirements

- Go 1.21+
- BIND9 (optional, for rndc operations)
- Zone file write permissions

## Notes

- Zone files are stored in the configured `zone_directory`
- The service generates standard BIND zone file format
- `rndc` commands require proper BIND configuration and permissions
- For production use, configure BIND to include the zones directory

## License

MIT
