package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDNSRecordTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		recordType DNSRecordType
		expected string
	}{
		{"A record", RecordTypeA, "A"},
		{"AAAA record", RecordTypeAAAA, "AAAA"},
		{"CNAME record", RecordTypeCNAME, "CNAME"},
		{"MX record", RecordTypeMX, "MX"},
		{"TXT record", RecordTypeTXT, "TXT"},
		{"NS record", RecordTypeNS, "NS"},
		{"SOA record", RecordTypeSOA, "SOA"},
		{"PTR record", RecordTypePTR, "PTR"},
		{"SRV record", RecordTypeSRV, "SRV"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.recordType) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.recordType)
			}
		})
	}
}

func TestDNSRecordJSONMarshaling(t *testing.T) {
	now := time.Now()
	record := DNSRecord{
		ID:        "rec_123456",
		Name:      "www",
		Type:      RecordTypeA,
		Value:     "192.168.1.1",
		TTL:       3600,
		Priority:  10,
		CreatedAt: now,
		UpdatedAt: now,
	}

	data, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("failed to marshal DNSRecord: %v", err)
	}

	var unmarshaled DNSRecord
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal DNSRecord: %v", err)
	}

	if unmarshaled.ID != record.ID {
		t.Errorf("ID mismatch: expected %s, got %s", record.ID, unmarshaled.ID)
	}
	if unmarshaled.Name != record.Name {
		t.Errorf("Name mismatch: expected %s, got %s", record.Name, unmarshaled.Name)
	}
	if unmarshaled.Type != record.Type {
		t.Errorf("Type mismatch: expected %s, got %s", record.Type, unmarshaled.Type)
	}
	if unmarshaled.Value != record.Value {
		t.Errorf("Value mismatch: expected %s, got %s", record.Value, unmarshaled.Value)
	}
	if unmarshaled.TTL != record.TTL {
		t.Errorf("TTL mismatch: expected %d, got %d", record.TTL, unmarshaled.TTL)
	}
	if unmarshaled.Priority != record.Priority {
		t.Errorf("Priority mismatch: expected %d, got %d", record.Priority, unmarshaled.Priority)
	}
}

func TestSOARecordJSONMarshaling(t *testing.T) {
	soa := SOARecord{
		MName:   "ns1.example.com.",
		RName:   "admin.example.com.",
		Serial:  2024011501,
		Refresh: 7200,
		Retry:   3600,
		Expire:  1209600,
		Minimum: 86400,
	}

	data, err := json.Marshal(soa)
	if err != nil {
		t.Fatalf("failed to marshal SOARecord: %v", err)
	}

	var unmarshaled SOARecord
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal SOARecord: %v", err)
	}

	if unmarshaled.MName != soa.MName {
		t.Errorf("MName mismatch: expected %s, got %s", soa.MName, unmarshaled.MName)
	}
	if unmarshaled.RName != soa.RName {
		t.Errorf("RName mismatch: expected %s, got %s", soa.RName, unmarshaled.RName)
	}
	if unmarshaled.Serial != soa.Serial {
		t.Errorf("Serial mismatch: expected %d, got %d", soa.Serial, unmarshaled.Serial)
	}
}

func TestDomainJSONMarshaling(t *testing.T) {
	now := time.Now()
	domain := Domain{
		Name: "example.com",
		Type: "master",
		File: "./zones/example.com.zone",
		SOA: SOARecord{
			MName:  "ns1.example.com.",
			RName:  "admin.example.com.",
			Serial: 2024011501,
		},
		Nameservers: []string{"ns1.example.com.", "ns2.example.com."},
		Records: []DNSRecord{
			{
				ID:    "rec_1",
				Name:  "@",
				Type:  RecordTypeA,
				Value: "192.168.1.1",
				TTL:   3600,
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	data, err := json.Marshal(domain)
	if err != nil {
		t.Fatalf("failed to marshal Domain: %v", err)
	}

	var unmarshaled Domain
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal Domain: %v", err)
	}

	if unmarshaled.Name != domain.Name {
		t.Errorf("Name mismatch: expected %s, got %s", domain.Name, unmarshaled.Name)
	}
	if unmarshaled.Type != domain.Type {
		t.Errorf("Type mismatch: expected %s, got %s", domain.Type, unmarshaled.Type)
	}
	if len(unmarshaled.Nameservers) != len(domain.Nameservers) {
		t.Errorf("Nameservers length mismatch: expected %d, got %d", len(domain.Nameservers), len(unmarshaled.Nameservers))
	}
	if len(unmarshaled.Records) != len(domain.Records) {
		t.Errorf("Records length mismatch: expected %d, got %d", len(domain.Records), len(unmarshaled.Records))
	}
}

func TestCreateDomainRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		request CreateDomainRequest
		valid   bool
	}{
		{
			name: "valid request",
			request: CreateDomainRequest{
				Name: "example.com",
			},
			valid: true,
		},
		{
			name: "empty name",
			request: CreateDomainRequest{
				Name: "",
			},
			valid: false,
		},
		{
			name: "with nameservers",
			request: CreateDomainRequest{
				Name:        "example.com",
				Nameservers: []string{"ns1.example.com."},
			},
			valid: true,
		},
		{
			name: "with SOA",
			request: CreateDomainRequest{
				Name: "example.com",
				SOA: SOARecord{
					MName: "ns1.example.com.",
					RName: "admin.example.com.",
				},
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.request)
			if err != nil {
				t.Fatalf("failed to marshal request: %v", err)
			}

			var unmarshaled CreateDomainRequest
			if err := json.Unmarshal(data, &unmarshaled); err != nil {
				t.Fatalf("failed to unmarshal request: %v", err)
			}

			if unmarshaled.Name != tt.request.Name {
				t.Errorf("Name mismatch: expected %s, got %s", tt.request.Name, unmarshaled.Name)
			}
		})
	}
}

func TestCreateRecordRequestJSONMarshaling(t *testing.T) {
	request := CreateRecordRequest{
		Name:     "www",
		Type:     RecordTypeA,
		Value:    "192.168.1.1",
		TTL:      3600,
		Priority: 0,
	}

	data, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("failed to marshal CreateRecordRequest: %v", err)
	}

	var unmarshaled CreateRecordRequest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal CreateRecordRequest: %v", err)
	}

	if unmarshaled.Name != request.Name {
		t.Errorf("Name mismatch: expected %s, got %s", request.Name, unmarshaled.Name)
	}
	if unmarshaled.Type != request.Type {
		t.Errorf("Type mismatch: expected %s, got %s", request.Type, unmarshaled.Type)
	}
	if unmarshaled.Value != request.Value {
		t.Errorf("Value mismatch: expected %s, got %s", request.Value, unmarshaled.Value)
	}
	if unmarshaled.TTL != request.TTL {
		t.Errorf("TTL mismatch: expected %d, got %d", request.TTL, unmarshaled.TTL)
	}
}

func TestUpdateRecordRequestJSONMarshaling(t *testing.T) {
	request := UpdateRecordRequest{
		Name:     "www",
		Type:     RecordTypeA,
		Value:    "192.168.1.100",
		TTL:      7200,
		Priority: 0,
	}

	data, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("failed to marshal UpdateRecordRequest: %v", err)
	}

	var unmarshaled UpdateRecordRequest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal UpdateRecordRequest: %v", err)
	}

	if unmarshaled.Value != request.Value {
		t.Errorf("Value mismatch: expected %s, got %s", request.Value, unmarshaled.Value)
	}
}

func TestAPIResponseJSONMarshaling(t *testing.T) {
	tests := []struct {
		name     string
		response APIResponse
	}{
		{
			name: "success response with data",
			response: APIResponse{
				Success: true,
				Message: "Operation completed",
				Data:    map[string]string{"key": "value"},
			},
		},
		{
			name: "error response",
			response: APIResponse{
				Success: false,
				Error:   "Something went wrong",
			},
		},
		{
			name: "success response without data",
			response: APIResponse{
				Success: true,
				Message: "Deleted successfully",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.response)
			if err != nil {
				t.Fatalf("failed to marshal APIResponse: %v", err)
			}

			var unmarshaled APIResponse
			if err := json.Unmarshal(data, &unmarshaled); err != nil {
				t.Fatalf("failed to unmarshal APIResponse: %v", err)
			}

			if unmarshaled.Success != tt.response.Success {
				t.Errorf("Success mismatch: expected %v, got %v", tt.response.Success, unmarshaled.Success)
			}
			if unmarshaled.Message != tt.response.Message {
				t.Errorf("Message mismatch: expected %s, got %s", tt.response.Message, unmarshaled.Message)
			}
			if unmarshaled.Error != tt.response.Error {
				t.Errorf("Error mismatch: expected %s, got %s", tt.response.Error, unmarshaled.Error)
			}
		})
	}
}

func TestHealthResponseJSONMarshaling(t *testing.T) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: "2024-01-15T10:30:00Z",
		Version:   "1.0.0",
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("failed to marshal HealthResponse: %v", err)
	}

	var unmarshaled HealthResponse
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal HealthResponse: %v", err)
	}

	if unmarshaled.Status != response.Status {
		t.Errorf("Status mismatch: expected %s, got %s", response.Status, unmarshaled.Status)
	}
	if unmarshaled.Timestamp != response.Timestamp {
		t.Errorf("Timestamp mismatch: expected %s, got %s", response.Timestamp, unmarshaled.Timestamp)
	}
	if unmarshaled.Version != response.Version {
		t.Errorf("Version mismatch: expected %s, got %s", response.Version, unmarshaled.Version)
	}
}

func TestDNSRecordWithOptionalPriority(t *testing.T) {
	// Test MX record with priority
	mxRecord := DNSRecord{
		ID:       "rec_mx",
		Name:     "@",
		Type:     RecordTypeMX,
		Value:    "mail.example.com.",
		TTL:      3600,
		Priority: 10,
	}

	data, err := json.Marshal(mxRecord)
	if err != nil {
		t.Fatalf("failed to marshal MX record: %v", err)
	}

	var unmarshaled DNSRecord
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal MX record: %v", err)
	}

	if unmarshaled.Priority != mxRecord.Priority {
		t.Errorf("Priority mismatch: expected %d, got %d", mxRecord.Priority, unmarshaled.Priority)
	}

	// Test A record without priority
	aRecord := DNSRecord{
		ID:       "rec_a",
		Name:     "www",
		Type:     RecordTypeA,
		Value:    "192.168.1.1",
		TTL:      3600,
		Priority: 0,
	}

	data, err = json.Marshal(aRecord)
	if err != nil {
		t.Fatalf("failed to marshal A record: %v", err)
	}

	var unmarshaledA DNSRecord
	if err := json.Unmarshal(data, &unmarshaledA); err != nil {
		t.Fatalf("failed to unmarshal A record: %v", err)
	}

	if unmarshaledA.Priority != 0 {
		t.Errorf("Priority should be 0 for A record, got %d", unmarshaledA.Priority)
	}
}
