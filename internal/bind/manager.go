package bind

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/livingopensource/bind-dns-api/internal/config"
	"github.com/livingopensource/bind-dns-api/internal/models"
)

// Manager handles BIND DNS operations
type Manager struct {
	config *config.BINDConfig
	mu     sync.RWMutex
}

// NewManager creates a new BIND manager
func NewManager(cfg *config.BINDConfig) *Manager {
	return &Manager{
		config: cfg,
	}
}

// ZoneExists checks if a zone file exists
func (m *Manager) ZoneExists(domainName string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	zoneFile := m.getZoneFilePath(domainName)
	_, err := os.Stat(zoneFile)
	return err == nil
}

// ListDomains returns all managed domains
func (m *Manager) ListDomains() ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	domains := make([]string, 0)

	entries, err := os.ReadDir(m.config.ZoneDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to read zone directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".zone")
		domains = append(domains, name)
	}

	return domains, nil
}

// GetDomain retrieves a domain with all its records
func (m *Manager) GetDomain(domainName string) (*models.Domain, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	zoneFile := m.getZoneFilePath(domainName)
	content, err := os.ReadFile(zoneFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read zone file: %w", err)
	}

	domain := &models.Domain{
		Name: domainName,
		Type: "master",
		File: zoneFile,
	}

	records, soa, nameservers := m.parseZoneFile(string(content))
	domain.Records = records
	domain.SOA = soa
	domain.Nameservers = nameservers

	return domain, nil
}

// CreateDomain creates a new zone file for a domain
func (m *Manager) CreateDomain(domainName string, req models.CreateDomainRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	zoneFile := m.getZoneFilePath(domainName)
	// Check if exists directly (avoid deadlock from calling ZoneExists while holding lock)
	if _, err := os.Stat(zoneFile); err == nil {
		return fmt.Errorf("domain %s already exists", domainName)
	}

	content := m.generateZoneFile(domainName, req)

	if err := os.WriteFile(zoneFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create zone file: %w", err)
	}

	return nil
}

// UpdateDomain updates an existing domain
func (m *Manager) UpdateDomain(domainName string, req models.CreateDomainRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	zoneFile := m.getZoneFilePath(domainName)
	// Check if exists directly (avoid deadlock from calling ZoneExists while holding lock)
	if _, err := os.Stat(zoneFile); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("domain %s does not exist", domainName)
		}
		return fmt.Errorf("failed to check domain existence: %w", err)
	}

	content := m.generateZoneFile(domainName, req)

	if err := os.WriteFile(zoneFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to update zone file: %w", err)
	}

	return nil
}

// DeleteDomain removes a domain's zone file
func (m *Manager) DeleteDomain(domainName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	zoneFile := m.getZoneFilePath(domainName)
	if err := os.Remove(zoneFile); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("domain %s does not exist", domainName)
		}
		return fmt.Errorf("failed to delete zone file: %w", err)
	}

	return nil
}

// ListRecords returns all DNS records for a domain
func (m *Manager) ListRecords(domainName string) ([]models.DNSRecord, error) {
	domain, err := m.GetDomain(domainName)
	if err != nil {
		return nil, err
	}
	return domain.Records, nil
}

// AddRecord adds a new DNS record to a domain
func (m *Manager) AddRecord(domainName string, req models.CreateRecordRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	zoneFile := m.getZoneFilePath(domainName)
	content, err := os.ReadFile(zoneFile)
	if err != nil {
		return fmt.Errorf("failed to read zone file: %w", err)
	}

	record := models.DNSRecord{
		ID:        generateRecordID(),
		Name:      req.Name,
		Type:      req.Type,
		Value:     req.Value,
		TTL:       req.TTL,
		Priority:  req.Priority,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if record.TTL == 0 {
		record.TTL = m.config.DefaultTTL
	}

	recordLine := m.formatRecordLine(record)
	updatedContent := string(content) + "\n" + recordLine

	if err := os.WriteFile(zoneFile, []byte(updatedContent), 0644); err != nil {
		return fmt.Errorf("failed to write zone file: %w", err)
	}

	return nil
}

// UpdateRecord updates an existing DNS record
func (m *Manager) UpdateRecord(domainName, recordName string, recordType models.DNSRecordType, req models.UpdateRecordRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	zoneFile := m.getZoneFilePath(domainName)
	content, err := os.ReadFile(zoneFile)
	if err != nil {
		return fmt.Errorf("failed to read zone file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	found := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, ";") || trimmed == "" {
			newLines = append(newLines, line)
			continue
		}

		if m.matchesRecordLine(trimmed, recordName, recordType) {
			found = true
			ttl := req.TTL
			if ttl == 0 {
				ttl = m.config.DefaultTTL
			}
			record := models.DNSRecord{
				Name:      recordName,
				Type:      recordType,
				Value:     req.Value,
				TTL:       ttl,
				Priority:  req.Priority,
				UpdatedAt: time.Now(),
			}
			newLines = append(newLines, m.formatRecordLine(record))
		} else {
			newLines = append(newLines, line)
		}
	}

	if !found {
		return fmt.Errorf("record %s of type %s not found", recordName, recordType)
	}

	if err := os.WriteFile(zoneFile, []byte(strings.Join(newLines, "\n")), 0644); err != nil {
		return fmt.Errorf("failed to write zone file: %w", err)
	}

	return nil
}

// DeleteRecord removes a DNS record from a domain
func (m *Manager) DeleteRecord(domainName, recordName string, recordType models.DNSRecordType) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	zoneFile := m.getZoneFilePath(domainName)
	content, err := os.ReadFile(zoneFile)
	if err != nil {
		return fmt.Errorf("failed to read zone file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	found := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, ";") || trimmed == "" {
			newLines = append(newLines, line)
			continue
		}

		if m.matchesRecordLine(trimmed, recordName, recordType) {
			found = true
			continue
		}
		newLines = append(newLines, line)
	}

	if !found {
		return fmt.Errorf("record %s of type %s not found", recordName, recordType)
	}

	if err := os.WriteFile(zoneFile, []byte(strings.Join(newLines, "\n")), 0644); err != nil {
		return fmt.Errorf("failed to write zone file: %w", err)
	}

	return nil
}

// ReloadZone reloads a specific zone using rndc
func (m *Manager) ReloadZone(domainName string) error {
	if _, err := exec.LookPath(m.config.RndcPath); err != nil {
		return fmt.Errorf("rndc not found at %s", m.config.RndcPath)
	}

	var args []string
	if m.config.RndcConfPath != "" {
		args = append(args, "-c", m.config.RndcConfPath)
	}
	args = append(args, "reload", domainName)

	cmd := exec.Command(m.config.RndcPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("rndc reload failed: %w, output: %s", err, string(output))
	}

	return nil
}

// ReloadAll reloads all zones
func (m *Manager) ReloadAll() error {
	if _, err := exec.LookPath(m.config.RndcPath); err != nil {
		return fmt.Errorf("rndc not found at %s", m.config.RndcPath)
	}

	var args []string
	if m.config.RndcConfPath != "" {
		args = append(args, "-c", m.config.RndcConfPath)
	}
	args = append(args, "reload")

	cmd := exec.Command(m.config.RndcPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("rndc reload failed: %w, output: %s", err, string(output))
	}

	return nil
}

// getZoneFilePath returns the path to a zone file
func (m *Manager) getZoneFilePath(domainName string) string {
	return filepath.Join(m.config.ZoneDirectory, domainName+".zone")
}

// generateZoneFile generates a zone file content
func (m *Manager) generateZoneFile(domainName string, req models.CreateDomainRequest) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("; Zone file for %s\n", domainName))
	sb.WriteString(fmt.Sprintf("; Generated by BIND DNS API - %s\n", time.Now().Format(time.RFC3339)))
	sb.WriteString("\n")

	// SOA Record
	soa := req.SOA
	if soa.MName == "" {
		soa.MName = fmt.Sprintf("ns1.%s.", domainName)
	}
	if soa.RName == "" {
		soa.RName = fmt.Sprintf("admin.%s.", domainName)
	}
	if soa.Serial == 0 {
		soa.Serial = time.Now().Unix()
	}
	if soa.Refresh == 0 {
		soa.Refresh = m.config.DefaultRefresh
	}
	if soa.Retry == 0 {
		soa.Retry = m.config.DefaultRetry
	}
	if soa.Expire == 0 {
		soa.Expire = m.config.DefaultExpire
	}
	if soa.Minimum == 0 {
		soa.Minimum = m.config.DefaultMinimum
	}

	sb.WriteString(fmt.Sprintf("$ORIGIN %s.\n", domainName))
	sb.WriteString(fmt.Sprintf("$TTL %d\n\n", m.config.DefaultTTL))
	sb.WriteString(fmt.Sprintf("@\tIN\tSOA\t%s\t%s\t(\n", soa.MName, soa.RName))
	sb.WriteString(fmt.Sprintf("\t\t\t%d\t; Serial\n", soa.Serial))
	sb.WriteString(fmt.Sprintf("\t\t\t%d\t; Refresh\n", soa.Refresh))
	sb.WriteString(fmt.Sprintf("\t\t\t%d\t; Retry\n", soa.Retry))
	sb.WriteString(fmt.Sprintf("\t\t\t%d\t; Expire\n", soa.Expire))
	sb.WriteString(fmt.Sprintf("\t\t\t%d\t; Minimum TTL\n", soa.Minimum))
	sb.WriteString("\t\t\t)\n\n")

	// Nameservers
	nameservers := req.Nameservers
	if len(nameservers) == 0 {
		nameservers = []string{fmt.Sprintf("ns1.%s.", domainName)}
	}

	for _, ns := range nameservers {
		sb.WriteString(fmt.Sprintf("@\tIN\tNS\t%s\n", ns))
	}

	sb.WriteString("\n")

	// Default A record for the domain
	sb.WriteString(fmt.Sprintf("@\tIN\tA\t127.0.0.1\n"))
	sb.WriteString(fmt.Sprintf("www\tIN\tA\t127.0.0.1\n"))

	return sb.String()
}

// parseZoneFile parses a zone file and extracts records
func (m *Manager) parseZoneFile(content string) ([]models.DNSRecord, models.SOARecord, []string) {
	var records []models.DNSRecord
	var soa models.SOARecord
	var nameservers []string

	scanner := bufio.NewScanner(strings.NewReader(content))
	var inSOA bool
	var soaLines []string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines for record parsing
		if line == "" || strings.HasPrefix(line, ";") {
			continue
		}

		// Handle SOA record
		if strings.Contains(line, "SOA") {
			inSOA = true
			soaLines = append(soaLines, line)
			continue
		}

		if inSOA {
			soaLines = append(soaLines, line)
			if strings.Contains(line, ")") {
				inSOA = false
				soa = m.parseSOARecord(soaLines)
			}
			continue
		}

		// Parse regular records
		record := m.parseRecordLine(line)
		if record != nil {
			if record.Type == models.RecordTypeNS {
				nameservers = append(nameservers, record.Value)
			}
			records = append(records, *record)
		}
	}

	return records, soa, nameservers
}

// parseSOARecord parses SOA record lines
func (m *Manager) parseSOARecord(lines []string) models.SOARecord {
	soa := models.SOARecord{}

	content := strings.Join(lines, " ")

	// Extract MName and RName
	parts := strings.Fields(content)
	for i, part := range parts {
		if part == "SOA" && i+2 < len(parts) {
			soa.MName = strings.Trim(parts[i+1], "()")
			soa.RName = strings.Trim(parts[i+2], "()")
			break
		}
	}

	// Extract numeric values
	for _, part := range parts {
		num, err := strconv.Atoi(strings.Trim(part, "();"))
		if err != nil {
			continue
		}
		if soa.Serial == 0 {
			soa.Serial = int64(num)
		} else if soa.Refresh == 0 {
			soa.Refresh = num
		} else if soa.Retry == 0 {
			soa.Retry = num
		} else if soa.Expire == 0 {
			soa.Expire = num
		} else if soa.Minimum == 0 {
			soa.Minimum = num
		}
	}

	return soa
}

// parseRecordLine parses a single record line
// Handles formats: NAME [TTL] [CLASS] TYPE VALUE or [TTL] [CLASS] NAME TYPE VALUE
func (m *Manager) parseRecordLine(line string) *models.DNSRecord {
	parts := strings.Fields(line)
	if len(parts) < 3 {
		return nil
	}

	record := &models.DNSRecord{
		ID:        generateRecordID(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		TTL:       m.config.DefaultTTL, // Default TTL from config
	}

	idx := 0

	// Check for TTL (numeric value) - can be first or second field
	hasTTL := false
	if ttl, err := strconv.Atoi(parts[idx]); err == nil {
		// Format: TTL CLASS NAME TYPE VALUE
		record.TTL = ttl
		idx++
		hasTTL = true
	}

	// Check for class (IN)
	if idx < len(parts) && strings.ToUpper(parts[idx]) == "IN" {
		idx++
	}

	// Need at least NAME and TYPE
	if idx >= len(parts)-1 {
		return nil
	}

	// Record name
	record.Name = parts[idx]
	idx++

	// If we didn't find TTL at the start, check for it after name
	// Format: NAME TTL CLASS TYPE VALUE
	if !hasTTL && idx < len(parts) {
		if ttl, err := strconv.Atoi(parts[idx]); err == nil {
			record.TTL = ttl
			idx++
		}
	}

	// Skip class (IN) if present after name/TTL
	if idx < len(parts) && strings.ToUpper(parts[idx]) == "IN" {
		idx++
	}

	if idx >= len(parts) {
		return nil
	}

	recordType := strings.ToUpper(parts[idx])
	// Validate it's a known record type before assigning
	switch recordType {
	case "A", "AAAA", "CNAME", "MX", "TXT", "NS", "SOA", "PTR", "SRV":
		record.Type = models.DNSRecordType(recordType)
		idx++
	default:
		// If we don't recognize the type, this might not be a valid record line
		return nil
	}

	// Record value (may include priority for MX records)
	if idx < len(parts) {
		// Check if next part is a priority number (for MX records)
		if record.Type == models.RecordTypeMX && idx < len(parts) {
			if priority, err := strconv.Atoi(parts[idx]); err == nil {
				record.Priority = priority
				idx++
			}
		}
		record.Value = strings.Join(parts[idx:], " ")
	}

	return record
}

// formatRecordLine formats a record into a zone file line
func (m *Manager) formatRecordLine(record models.DNSRecord) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s\t%d\tIN\t%s\t", record.Name, record.TTL, record.Type))

	if record.Type == models.RecordTypeMX && record.Priority > 0 {
		sb.WriteString(fmt.Sprintf("%d ", record.Priority))
	}

	sb.WriteString(record.Value)

	return sb.String()
}

// matchesRecordLine checks if a line matches the given record name and type
func (m *Manager) matchesRecordLine(line, name string, recordType models.DNSRecordType) bool {
	parts := strings.Fields(line)
	if len(parts) < 3 {
		return false
	}

	idx := 0
	// Skip TTL if present at start
	if _, err := strconv.Atoi(parts[idx]); err == nil {
		idx++
	}
	// Skip class (IN) if present before name
	if idx < len(parts) && strings.ToUpper(parts[idx]) == "IN" {
		idx++
	}

	// Check name
	if idx >= len(parts) || parts[idx] != name {
		return false
	}
	idx++

	// Skip TTL if present after name (format: NAME TTL CLASS TYPE VALUE)
	if idx < len(parts) {
		if _, err := strconv.Atoi(parts[idx]); err == nil {
			idx++
		}
	}

	// Skip class (IN) if present after name/TTL
	if idx < len(parts) && strings.ToUpper(parts[idx]) == "IN" {
		idx++
	}

	// Check type
	if idx >= len(parts) || strings.ToUpper(parts[idx]) != string(recordType) {
		return false
	}

	return true
}

// generateRecordID generates a unique record ID
func generateRecordID() string {
	return fmt.Sprintf("rec_%d", time.Now().UnixNano())
}
