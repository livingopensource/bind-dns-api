package models

import "time"

// DNSRecordType represents DNS record types
type DNSRecordType string

const (
	RecordTypeA     DNSRecordType = "A"
	RecordTypeAAAA  DNSRecordType = "AAAA"
	RecordTypeCNAME DNSRecordType = "CNAME"
	RecordTypeMX    DNSRecordType = "MX"
	RecordTypeTXT   DNSRecordType = "TXT"
	RecordTypeNS    DNSRecordType = "NS"
	RecordTypeSOA   DNSRecordType = "SOA"
	RecordTypePTR   DNSRecordType = "PTR"
	RecordTypeSRV   DNSRecordType = "SRV"
)

// DNSRecord represents a single DNS record
type DNSRecord struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Type      DNSRecordType `json:"type"`
	Value     string        `json:"value"`
	TTL       int           `json:"ttl"`
	Priority  int           `json:"priority,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// SOARecord represents SOA record data
type SOARecord struct {
	MName     string `json:"mname"`
	RName     string `json:"rname"`
	Serial    int64  `json:"serial"`
	Refresh   int    `json:"refresh"`
	Retry     int    `json:"retry"`
	Expire    int    `json:"expire"`
	Minimum   int    `json:"minimum"`
}

// Domain represents a DNS domain/zone
type Domain struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"` // master, slave, hint
	File        string      `json:"file"`
	SOA         SOARecord   `json:"soa"`
	Nameservers []string    `json:"nameservers"`
	Records     []DNSRecord `json:"records"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// CreateDomainRequest represents the request body for creating a domain
type CreateDomainRequest struct {
	Name        string   `json:"name" binding:"required"`
	Type        string   `json:"type"`
	Nameservers []string `json:"nameservers"`
	SOA         SOARecord `json:"soa"`
}

// CreateRecordRequest represents the request body for creating a DNS record
type CreateRecordRequest struct {
	Name     string        `json:"name" binding:"required"`
	Type     DNSRecordType `json:"type" binding:"required"`
	Value    string        `json:"value" binding:"required"`
	TTL      int           `json:"ttl"`
	Priority int           `json:"priority"`
}

// UpdateRecordRequest represents the request body for updating a DNS record
type UpdateRecordRequest struct {
	Name     string        `json:"name"`
	Type     DNSRecordType `json:"type"`
	Value    string        `json:"value"`
	TTL      int           `json:"ttl"`
	Priority int           `json:"priority"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
}
