package bind

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/livingopensource/bind-dns-api/internal/config"
	"github.com/livingopensource/bind-dns-api/internal/models"
)

func createTestManager(t *testing.T) (*Manager, string) {
	tmpDir := t.TempDir()

	cfg := &config.BINDConfig{
		ZoneDirectory:  tmpDir,
		DefaultTTL:     3600,
		DefaultRefresh: 7200,
		DefaultRetry:   3600,
		DefaultExpire:  1209600,
		DefaultMinimum: 86400,
	}

	return NewManager(cfg), tmpDir
}

func TestNewManager(t *testing.T) {
	cfg := &config.BINDConfig{
		ZoneDirectory: "./zones",
		DefaultTTL:    3600,
	}

	manager := NewManager(cfg)
	if manager == nil {
		t.Fatal("NewManager returned nil")
	}
	if manager.config != cfg {
		t.Error("Manager config not set correctly")
	}
}

func TestGetZoneFilePath(t *testing.T) {
	manager, _ := createTestManager(t)

	tests := []struct {
		domain   string
		expected string
	}{
		{"example.com", filepath.Join(manager.config.ZoneDirectory, "example.com.zone")},
		{"test.org", filepath.Join(manager.config.ZoneDirectory, "test.org.zone")},
		{"sub.domain.com", filepath.Join(manager.config.ZoneDirectory, "sub.domain.com.zone")},
	}

	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			result := manager.getZoneFilePath(tt.domain)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestZoneExists(t *testing.T) {
	manager, tmpDir := createTestManager(t)

	// Zone should not exist initially
	if manager.ZoneExists("example.com") {
		t.Error("Zone should not exist initially")
	}

	// Create a zone file
	zoneFile := filepath.Join(tmpDir, "example.com.zone")
	if err := os.WriteFile(zoneFile, []byte("; test zone"), 0644); err != nil {
		t.Fatalf("failed to create zone file: %v", err)
	}

	// Now it should exist
	if !manager.ZoneExists("example.com") {
		t.Error("Zone should exist after creation")
	}
}

func TestListDomains(t *testing.T) {
	manager, tmpDir := createTestManager(t)

	// Initially empty
	domains, err := manager.ListDomains()
	if err != nil {
		t.Fatalf("ListDomains failed: %v", err)
	}
	if len(domains) != 0 {
		t.Errorf("expected 0 domains, got %d", len(domains))
	}

	// Create some zone files
	zones := []string{"example.com.zone", "test.org.zone", "mydomain.net.zone"}
	for _, zone := range zones {
		if err := os.WriteFile(filepath.Join(tmpDir, zone), []byte("; test"), 0644); err != nil {
			t.Fatalf("failed to create zone file: %v", err)
		}
	}

	domains, err = manager.ListDomains()
	if err != nil {
		t.Fatalf("ListDomains failed: %v", err)
	}
	if len(domains) != 3 {
		t.Errorf("expected 3 domains, got %d", len(domains))
	}
}

func TestCreateDomain(t *testing.T) {
	manager, _ := createTestManager(t)

	req := models.CreateDomainRequest{
		Name: "example.com",
		Nameservers: []string{"ns1.example.com.", "ns2.example.com."},
		SOA: models.SOARecord{
			MName: "ns1.example.com.",
			RName: "admin.example.com.",
		},
	}

	err := manager.CreateDomain("example.com", req)
	if err != nil {
		t.Fatalf("CreateDomain failed: %v", err)
	}

	// Verify zone file was created
	if !manager.ZoneExists("example.com") {
		t.Error("Zone file was not created")
	}

	// Try to create again - should fail
	err = manager.CreateDomain("example.com", req)
	if err == nil {
		t.Error("CreateDomain should fail for existing domain")
	}
}

func TestCreateDomainWithDefaults(t *testing.T) {
	manager, _ := createTestManager(t)

	req := models.CreateDomainRequest{
		Name: "test.com",
	}

	err := manager.CreateDomain("test.com", req)
	if err != nil {
		t.Fatalf("CreateDomain failed: %v", err)
	}

	// Read and verify zone file content
	domain, err := manager.GetDomain("test.com")
	if err != nil {
		t.Fatalf("GetDomain failed: %v", err)
	}

	// Should have default nameserver
	if len(domain.Nameservers) == 0 {
		t.Error("Should have default nameserver")
	}

	// SOA should have defaults
	if domain.SOA.Refresh == 0 {
		t.Error("SOA Refresh should have default value")
	}
	if domain.SOA.Retry == 0 {
		t.Error("SOA Retry should have default value")
	}
}

func TestGetDomain(t *testing.T) {
	manager, _ := createTestManager(t)

	// Create a domain first
	req := models.CreateDomainRequest{
		Name: "example.com",
		Nameservers: []string{"ns1.example.com.", "ns2.example.com."},
	}
	if err := manager.CreateDomain("example.com", req); err != nil {
		t.Fatalf("CreateDomain failed: %v", err)
	}

	// Get the domain
	domain, err := manager.GetDomain("example.com")
	if err != nil {
		t.Fatalf("GetDomain failed: %v", err)
	}

	if domain.Name != "example.com" {
		t.Errorf("expected name example.com, got %s", domain.Name)
	}
	if domain.Type != "master" {
		t.Errorf("expected type master, got %s", domain.Type)
	}
	if len(domain.Nameservers) != 2 {
		t.Errorf("expected 2 nameservers, got %d", len(domain.Nameservers))
	}
}

func TestGetDomainNotFound(t *testing.T) {
	manager, _ := createTestManager(t)

	_, err := manager.GetDomain("nonexistent.com")
	if err == nil {
		t.Error("GetDomain should return error for non-existent domain")
	}
}

func TestUpdateDomain(t *testing.T) {
	manager, _ := createTestManager(t)

	// Create domain first
	createReq := models.CreateDomainRequest{Name: "example.com"}
	if err := manager.CreateDomain("example.com", createReq); err != nil {
		t.Fatalf("CreateDomain failed: %v", err)
	}

	// Update domain
	updateReq := models.CreateDomainRequest{
		Name: "example.com",
		Nameservers: []string{"ns1.updated.com.", "ns2.updated.com.", "ns3.updated.com."},
	}
	err := manager.UpdateDomain("example.com", updateReq)
	if err != nil {
		t.Fatalf("UpdateDomain failed: %v", err)
	}

	// Verify update
	domain, err := manager.GetDomain("example.com")
	if err != nil {
		t.Fatalf("GetDomain failed: %v", err)
	}
	if len(domain.Nameservers) != 3 {
		t.Errorf("expected 3 nameservers after update, got %d", len(domain.Nameservers))
	}
}

func TestUpdateDomainNotFound(t *testing.T) {
	manager, _ := createTestManager(t)

	req := models.CreateDomainRequest{Name: "nonexistent.com"}
	err := manager.UpdateDomain("nonexistent.com", req)
	if err == nil {
		t.Error("UpdateDomain should return error for non-existent domain")
	}
}

func TestDeleteDomain(t *testing.T) {
	manager, _ := createTestManager(t)

	// Create domain first
	req := models.CreateDomainRequest{Name: "example.com"}
	if err := manager.CreateDomain("example.com", req); err != nil {
		t.Fatalf("CreateDomain failed: %v", err)
	}

	// Delete domain
	err := manager.DeleteDomain("example.com")
	if err != nil {
		t.Fatalf("DeleteDomain failed: %v", err)
	}

	// Verify deletion
	if manager.ZoneExists("example.com") {
		t.Error("Zone should not exist after deletion")
	}
}

func TestDeleteDomainNotFound(t *testing.T) {
	manager, _ := createTestManager(t)

	err := manager.DeleteDomain("nonexistent.com")
	if err == nil {
		t.Error("DeleteDomain should return error for non-existent domain")
	}
}

func TestAddRecord(t *testing.T) {
	manager, _ := createTestManager(t)

	// Create domain first
	createReq := models.CreateDomainRequest{Name: "example.com"}
	if err := manager.CreateDomain("example.com", createReq); err != nil {
		t.Fatalf("CreateDomain failed: %v", err)
	}

	// Add A record
	recordReq := models.CreateRecordRequest{
		Name:  "api",
		Type:  models.RecordTypeA,
		Value: "192.168.1.100",
		TTL:   3600,
	}
	err := manager.AddRecord("example.com", recordReq)
	if err != nil {
		t.Fatalf("AddRecord failed: %v", err)
	}

	// Verify record was added
	records, err := manager.ListRecords("example.com")
	if err != nil {
		t.Fatalf("ListRecords failed: %v", err)
	}

	found := false
	for _, rec := range records {
		if rec.Name == "api" && rec.Type == models.RecordTypeA && rec.Value == "192.168.1.100" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Record was not added")
	}
}

func TestAddRecordWithDefaultTTL(t *testing.T) {
	manager, _ := createTestManager(t)

	createReq := models.CreateDomainRequest{Name: "example.com"}
	if err := manager.CreateDomain("example.com", createReq); err != nil {
		t.Fatalf("CreateDomain failed: %v", err)
	}

	recordReq := models.CreateRecordRequest{
		Name:  "www",
		Type:  models.RecordTypeA,
		Value: "192.168.1.1",
		TTL:   0, // Should use default
	}
	err := manager.AddRecord("example.com", recordReq)
	if err != nil {
		t.Fatalf("AddRecord failed: %v", err)
	}

	records, err := manager.ListRecords("example.com")
	if err != nil {
		t.Fatalf("ListRecords failed: %v", err)
	}

	for _, rec := range records {
		if rec.Name == "www" && rec.TTL == 0 {
			t.Error("TTL should have default value, not 0")
		}
	}
}

func TestAddRecordMXWithPriority(t *testing.T) {
	manager, _ := createTestManager(t)

	createReq := models.CreateDomainRequest{Name: "example.com"}
	if err := manager.CreateDomain("example.com", createReq); err != nil {
		t.Fatalf("CreateDomain failed: %v", err)
	}

	recordReq := models.CreateRecordRequest{
		Name:     "@",
		Type:     models.RecordTypeMX,
		Value:    "mail.example.com.",
		Priority: 10,
		TTL:      3600,
	}
	err := manager.AddRecord("example.com", recordReq)
	if err != nil {
		t.Fatalf("AddRecord failed: %v", err)
	}

	records, err := manager.ListRecords("example.com")
	if err != nil {
		t.Fatalf("ListRecords failed: %v", err)
	}

	found := false
	for _, rec := range records {
		if rec.Type == models.RecordTypeMX && rec.Priority == 10 {
			found = true
			break
		}
	}
	if !found {
		t.Error("MX record with priority was not added correctly")
	}
}

func TestListRecords(t *testing.T) {
	manager, _ := createTestManager(t)

	createReq := models.CreateDomainRequest{Name: "example.com"}
	if err := manager.CreateDomain("example.com", createReq); err != nil {
		t.Fatalf("CreateDomain failed: %v", err)
	}

	// Add multiple records
	records := []models.CreateRecordRequest{
		{Name: "www", Type: models.RecordTypeA, Value: "192.168.1.1"},
		{Name: "api", Type: models.RecordTypeA, Value: "192.168.1.2"},
		{Name: "@", Type: models.RecordTypeMX, Value: "mail.example.com.", Priority: 10},
	}

	for _, rec := range records {
		if err := manager.AddRecord("example.com", rec); err != nil {
			t.Fatalf("AddRecord failed: %v", err)
		}
	}

	listedRecords, err := manager.ListRecords("example.com")
	if err != nil {
		t.Fatalf("ListRecords failed: %v", err)
	}

	// Should have at least the records we added (plus defaults)
	if len(listedRecords) < 3 {
		t.Errorf("expected at least 3 records, got %d", len(listedRecords))
	}
}

func TestListRecordsNotFound(t *testing.T) {
	manager, _ := createTestManager(t)

	_, err := manager.ListRecords("nonexistent.com")
	if err == nil {
		t.Error("ListRecords should return error for non-existent domain")
	}
}

func TestUpdateRecord(t *testing.T) {
	manager, _ := createTestManager(t)

	createReq := models.CreateDomainRequest{Name: "example.com"}
	if err := manager.CreateDomain("example.com", createReq); err != nil {
		t.Fatalf("CreateDomain failed: %v", err)
	}

	// Add a record first
	addReq := models.CreateRecordRequest{
		Name:  "www",
		Type:  models.RecordTypeA,
		Value: "192.168.1.1",
		TTL:   3600,
	}
	if err := manager.AddRecord("example.com", addReq); err != nil {
		t.Fatalf("AddRecord failed: %v", err)
	}

	// Update the record
	updateReq := models.UpdateRecordRequest{
		Value: "192.168.1.100",
		TTL:   7200,
	}
	err := manager.UpdateRecord("example.com", "www", models.RecordTypeA, updateReq)
	if err != nil {
		t.Fatalf("UpdateRecord failed: %v", err)
	}

	// Verify update
	records, err := manager.ListRecords("example.com")
	if err != nil {
		t.Fatalf("ListRecords failed: %v", err)
	}

	found := false
	for _, rec := range records {
		if rec.Name == "www" && rec.Value == "192.168.1.100" && rec.TTL == 7200 {
			found = true
			break
		}
	}
	if !found {
		t.Error("Record was not updated correctly")
	}
}

func TestUpdateRecordNotFound(t *testing.T) {
	manager, _ := createTestManager(t)

	createReq := models.CreateDomainRequest{Name: "example.com"}
	if err := manager.CreateDomain("example.com", createReq); err != nil {
		t.Fatalf("CreateDomain failed: %v", err)
	}

	updateReq := models.UpdateRecordRequest{Value: "192.168.1.1"}
	err := manager.UpdateRecord("example.com", "nonexistent", models.RecordTypeA, updateReq)
	if err == nil {
		t.Error("UpdateRecord should return error for non-existent record")
	}
}

func TestDeleteRecord(t *testing.T) {
	manager, _ := createTestManager(t)

	createReq := models.CreateDomainRequest{Name: "example.com"}
	if err := manager.CreateDomain("example.com", createReq); err != nil {
		t.Fatalf("CreateDomain failed: %v", err)
	}

	// Add a record
	addReq := models.CreateRecordRequest{
		Name:  "temp",
		Type:  models.RecordTypeA,
		Value: "192.168.1.1",
	}
	if err := manager.AddRecord("example.com", addReq); err != nil {
		t.Fatalf("AddRecord failed: %v", err)
	}

	// Delete the record
	err := manager.DeleteRecord("example.com", "temp", models.RecordTypeA)
	if err != nil {
		t.Fatalf("DeleteRecord failed: %v", err)
	}

	// Verify deletion
	records, err := manager.ListRecords("example.com")
	if err != nil {
		t.Fatalf("ListRecords failed: %v", err)
	}

	for _, rec := range records {
		if rec.Name == "temp" {
			t.Error("Record was not deleted")
			break
		}
	}
}

func TestDeleteRecordNotFound(t *testing.T) {
	manager, _ := createTestManager(t)

	createReq := models.CreateDomainRequest{Name: "example.com"}
	if err := manager.CreateDomain("example.com", createReq); err != nil {
		t.Fatalf("CreateDomain failed: %v", err)
	}

	err := manager.DeleteRecord("example.com", "nonexistent", models.RecordTypeA)
	if err == nil {
		t.Error("DeleteRecord should return error for non-existent record")
	}
}

func TestFormatRecordLine(t *testing.T) {
	manager, _ := createTestManager(t)

	tests := []struct {
		name     string
		record   models.DNSRecord
		expected string
	}{
		{
			name: "A record",
			record: models.DNSRecord{
				Name: "www",
				Type: models.RecordTypeA,
				Value: "192.168.1.1",
				TTL:  3600,
			},
			expected: "www\t3600\tIN\tA\t192.168.1.1",
		},
		{
			name: "MX record with priority",
			record: models.DNSRecord{
				Name:     "@",
				Type:     models.RecordTypeMX,
				Value:    "mail.example.com.",
				TTL:      3600,
				Priority: 10,
			},
			expected: "@\t3600\tIN\tMX\t10 mail.example.com.",
		},
		{
			name: "CNAME record",
			record: models.DNSRecord{
				Name:  "blog",
				Type:  models.RecordTypeCNAME,
				Value: "www.example.com.",
				TTL:   3600,
			},
			expected: "blog\t3600\tIN\tCNAME\twww.example.com.",
		},
		{
			name: "TXT record",
			record: models.DNSRecord{
				Name:  "@",
				Type:  models.RecordTypeTXT,
				Value: "\"v=spf1 include:_spf.google.com ~all\"",
				TTL:   3600,
			},
			expected: "@\t3600\tIN\tTXT\t\"v=spf1 include:_spf.google.com ~all\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.formatRecordLine(tt.record)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestMatchesRecordLine(t *testing.T) {
	manager, _ := createTestManager(t)

	tests := []struct {
		name       string
		line       string
		recordName string
		recordType models.DNSRecordType
		expected   bool
	}{
		{
			name:       "match A record",
			line:       "www\t3600\tIN\tA\t192.168.1.1",
			recordName: "www",
			recordType: models.RecordTypeA,
			expected:   true,
		},
		{
			name:       "match MX record",
			line:       "@\t3600\tIN\tMX\t10 mail.example.com.",
			recordName: "@",
			recordType: models.RecordTypeMX,
			expected:   true,
		},
		{
			name:       "no match different name",
			line:       "www\t3600\tIN\tA\t192.168.1.1",
			recordName: "api",
			recordType: models.RecordTypeA,
			expected:   false,
		},
		{
			name:       "no match different type",
			line:       "www\t3600\tIN\tA\t192.168.1.1",
			recordName: "www",
			recordType: models.RecordTypeCNAME,
			expected:   false,
		},
		{
			name:       "short line",
			line:       "invalid",
			recordName: "www",
			recordType: models.RecordTypeA,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := manager.matchesRecordLine(tt.line, tt.recordName, tt.recordType)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestParseRecordLine(t *testing.T) {
	manager, _ := createTestManager(t)

	tests := []struct {
		name     string
		line     string
		wantName string
		wantType models.DNSRecordType
		wantValue string
		wantTTL  int
	}{
		{
			name:      "A record with TTL",
			line:      "www 3600 IN A 192.168.1.1",
			wantName:  "www",
			wantType:  models.RecordTypeA,
			wantValue: "192.168.1.1",
			wantTTL:   3600,
		},
		{
			name:      "A record without explicit TTL",
			line:      "www IN A 192.168.1.1",
			wantName:  "www",
			wantType:  models.RecordTypeA,
			wantValue: "192.168.1.1",
			wantTTL:   3600, // Default TTL from config
		},
		{
			name:      "MX record with priority",
			line:      "@ IN MX 10 mail.example.com.",
			wantName:  "@",
			wantType:  models.RecordTypeMX,
			wantValue: "mail.example.com.",
			wantTTL:   3600, // Default TTL from config
		},
		{
			name:      "CNAME record",
			line:      "blog IN CNAME www.example.com.",
			wantName:  "blog",
			wantType:  models.RecordTypeCNAME,
			wantValue: "www.example.com.",
			wantTTL:   3600, // Default TTL from config
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := manager.parseRecordLine(tt.line)
			if record == nil {
				t.Fatal("parseRecordLine returned nil")
			}
			if record.Name != tt.wantName {
				t.Errorf("Name: expected %s, got %s", tt.wantName, record.Name)
			}
			if record.Type != tt.wantType {
				t.Errorf("Type: expected %s, got %s", tt.wantType, record.Type)
			}
			if record.Value != tt.wantValue {
				t.Errorf("Value: expected %s, got %s", tt.wantValue, record.Value)
			}
			if record.TTL != tt.wantTTL {
				t.Errorf("TTL: expected %d, got %d", tt.wantTTL, record.TTL)
			}
		})
	}
}

func TestGenerateZoneFile(t *testing.T) {
	manager, _ := createTestManager(t)

	req := models.CreateDomainRequest{
		Name: "example.com",
		Nameservers: []string{"ns1.example.com.", "ns2.example.com."},
		SOA: models.SOARecord{
			MName: "ns1.example.com.",
			RName: "admin.example.com.",
		},
	}

	content := manager.generateZoneFile("example.com", req)

	// Verify zone file content
	checks := []struct {
		name     string
		contains string
	}{
		{"SOA header", "SOA"},
		{"Primary nameserver", "ns1.example.com."},
		{"Admin email", "admin.example.com."},
		{"NS record", "IN\tNS"},
		{"Default A record", "IN\tA"},
		{"ORIGIN", "$ORIGIN example.com"},
		{"TTL", "$TTL"},
	}

	for _, check := range checks {
		if !strings.Contains(content, check.contains) {
			t.Errorf("Zone file should contain %q", check.contains)
		}
	}
}

func TestGenerateZoneFileWithDefaults(t *testing.T) {
	manager, _ := createTestManager(t)

	req := models.CreateDomainRequest{
		Name: "test.com",
	}

	content := manager.generateZoneFile("test.com", req)

	// Should have default nameserver
	if !strings.Contains(content, "ns1.test.com.") {
		t.Error("Should have default nameserver")
	}

	// Should have SOA values
	if !strings.Contains(content, "; Serial") {
		t.Error("Should have Serial comment")
	}
	if !strings.Contains(content, "; Refresh") {
		t.Error("Should have Refresh comment")
	}
}

func TestParseSOARecord(t *testing.T) {
	manager, _ := createTestManager(t)

	soaLines := []string{
		"@ IN SOA ns1.example.com. admin.example.com. (",
		"2024011501 ; Serial",
		"7200 ; Refresh",
		"3600 ; Retry",
		"1209600 ; Expire",
		"86400 ) ; Minimum TTL",
	}

	soa := manager.parseSOARecord(soaLines)

	if soa.MName != "ns1.example.com." {
		t.Errorf("MName: expected ns1.example.com., got %s", soa.MName)
	}
	if soa.RName != "admin.example.com." {
		t.Errorf("RName: expected admin.example.com., got %s", soa.RName)
	}
	if soa.Serial != 2024011501 {
		t.Errorf("Serial: expected 2024011501, got %d", soa.Serial)
	}
	if soa.Refresh != 7200 {
		t.Errorf("Refresh: expected 7200, got %d", soa.Refresh)
	}
	if soa.Retry != 3600 {
		t.Errorf("Retry: expected 3600, got %d", soa.Retry)
	}
	if soa.Expire != 1209600 {
		t.Errorf("Expire: expected 1209600, got %d", soa.Expire)
	}
	if soa.Minimum != 86400 {
		t.Errorf("Minimum: expected 86400, got %d", soa.Minimum)
	}
}

func TestParseZoneFile(t *testing.T) {
	manager, _ := createTestManager(t)

	zoneContent := `; Zone file for example.com
$ORIGIN example.com.
$TTL 3600

@ IN SOA ns1.example.com. admin.example.com. (
	2024011501 ; Serial
	7200 ; Refresh
	3600 ; Retry
	1209600 ; Expire
	86400 ; Minimum TTL
)

@ IN NS ns1.example.com.
@ IN NS ns2.example.com.

@ IN A 192.168.1.1
www IN A 192.168.1.2
mail IN A 192.168.1.3

@ IN MX 10 mail.example.com.

blog IN CNAME www.example.com.
`

	records, soa, nameservers := manager.parseZoneFile(zoneContent)

	// Check SOA
	if soa.MName != "ns1.example.com." {
		t.Errorf("SOA MName: expected ns1.example.com., got %s", soa.MName)
	}

	// Check nameservers
	if len(nameservers) != 2 {
		t.Errorf("expected 2 nameservers, got %d", len(nameservers))
	}

	// Check records
	if len(records) < 5 {
		t.Errorf("expected at least 5 records, got %d", len(records))
	}

	// Find specific records
	foundA := false
	foundMX := false
	foundCNAME := false
	for _, rec := range records {
		if rec.Type == models.RecordTypeA && rec.Name == "www" {
			foundA = true
		}
		if rec.Type == models.RecordTypeMX {
			foundMX = true
		}
		if rec.Type == models.RecordTypeCNAME && rec.Name == "blog" {
			foundCNAME = true
		}
	}

	if !foundA {
		t.Error("Should find A record")
	}
	if !foundMX {
		t.Error("Should find MX record")
	}
	if !foundCNAME {
		t.Error("Should find CNAME record")
	}
}

func TestGenerateRecordID(t *testing.T) {
	// Generate multiple IDs and ensure they're unique
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := generateRecordID()
		if ids[id] {
			t.Errorf("Duplicate ID generated: %s", id)
		}
		ids[id] = true
	}

	// Check ID format
	id := generateRecordID()
	if !strings.HasPrefix(id, "rec_") {
		t.Errorf("ID should start with 'rec_', got %s", id)
	}
}

func TestManagerConcurrency(t *testing.T) {
	manager, _ := createTestManager(t)

	// Create initial domain
	createReq := models.CreateDomainRequest{Name: "example.com"}
	if err := manager.CreateDomain("example.com", createReq); err != nil {
		t.Fatalf("CreateDomain failed: %v", err)
	}

	// Concurrent reads
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			manager.GetDomain("example.com")
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestReloadZoneNoRndc(t *testing.T) {
	manager, _ := createTestManager(t)
	manager.config.RndcPath = "/nonexistent/rndc"

	err := manager.ReloadZone("example.com")
	if err == nil {
		t.Error("ReloadZone should fail when rndc doesn't exist")
	}
}

func TestReloadAllNoRndc(t *testing.T) {
	manager, _ := createTestManager(t)
	manager.config.RndcPath = "/nonexistent/rndc"

	err := manager.ReloadAll()
	if err == nil {
		t.Error("ReloadAll should fail when rndc doesn't exist")
	}
}
