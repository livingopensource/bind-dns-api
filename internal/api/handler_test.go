package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/livingopensource/bind-dns-api/internal/bind"
	"github.com/livingopensource/bind-dns-api/internal/config"
	"github.com/livingopensource/bind-dns-api/internal/models"
)

func setupTestRouter(t *testing.T) (*gin.Engine, *bind.Manager, string) {
	gin.SetMode(gin.TestMode)

	tmpDir := t.TempDir()

	cfg := &config.BINDConfig{
		ZoneDirectory:  tmpDir,
		DefaultTTL:     3600,
		DefaultRefresh: 7200,
		DefaultRetry:   3600,
		DefaultExpire:  1209600,
		DefaultMinimum: 86400,
		RndcPath:       "/nonexistent/rndc",
	}

	manager := bind.NewManager(cfg)
	handler := NewHandler(manager)

	router := gin.New()
	handler.RegisterRoutes(router)

	return router, manager, tmpDir
}

func TestHealthCheck(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response models.HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Status != "healthy" {
		t.Errorf("expected status 'healthy', got %s", response.Status)
	}
	if response.Version != Version {
		t.Errorf("expected version %s, got %s", Version, response.Version)
	}
	if response.Timestamp == "" {
		t.Error("timestamp should not be empty")
	}
}

func TestListDomainsEmpty(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/domains", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response models.APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Error("expected success to be true")
	}

	data, ok := response.Data.([]interface{})
	if !ok || len(data) != 0 {
		t.Error("expected empty data array")
	}
}

func TestCreateDomain(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	reqBody := models.CreateDomainRequest{
		Name: "example.com",
		Nameservers: []string{"ns1.example.com.", "ns2.example.com."},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/domains", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var response models.APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Errorf("expected success to be true, got error: %s", response.Error)
	}
}

func TestCreateDomainInvalidJSON(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/domains", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var response models.APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Success {
		t.Error("expected success to be false")
	}
}

func TestCreateDomainEmptyName(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	reqBody := models.CreateDomainRequest{Name: ""}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/domains", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestCreateDomainDuplicate(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	reqBody := models.CreateDomainRequest{Name: "example.com"}
	body, _ := json.Marshal(reqBody)

	// Create first domain
	req1, _ := http.NewRequest(http.MethodPost, "/api/v1/domains", bytes.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	// Try to create again
	req2, _ := http.NewRequest(http.MethodPost, "/api/v1/domains", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", w2.Code)
	}
}

func TestGetDomain(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	// Create domain first
	reqBody := models.CreateDomainRequest{Name: "example.com"}
	body, _ := json.Marshal(reqBody)

	req1, _ := http.NewRequest(http.MethodPost, "/api/v1/domains", bytes.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	// Get domain
	req2, _ := http.NewRequest(http.MethodGet, "/api/v1/domains/example.com", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w2.Code, w2.Body.String())
	}

	var response models.APIResponse
	if err := json.Unmarshal(w2.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Errorf("expected success to be true, got error: %s", response.Error)
	}
}

func TestGetDomainNotFound(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/domains/nonexistent.com", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestUpdateDomain(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	// Create domain first
	createBody, _ := json.Marshal(models.CreateDomainRequest{Name: "example.com"})
	req1, _ := http.NewRequest(http.MethodPost, "/api/v1/domains", bytes.NewReader(createBody))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	// Update domain
	updateBody, _ := json.Marshal(models.CreateDomainRequest{
		Name: "example.com",
		Nameservers: []string{"ns1.updated.com."},
	})
	req2, _ := http.NewRequest(http.MethodPut, "/api/v1/domains/example.com", bytes.NewReader(updateBody))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w2.Code, w2.Body.String())
	}

	var response models.APIResponse
	if err := json.Unmarshal(w2.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Errorf("expected success to be true, got error: %s", response.Error)
	}
}

func TestUpdateDomainNotFound(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	updateBody, _ := json.Marshal(models.CreateDomainRequest{Name: "nonexistent.com"})
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/domains/nonexistent.com", bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestDeleteDomain(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	// Create domain first
	createBody, _ := json.Marshal(models.CreateDomainRequest{Name: "example.com"})
	req1, _ := http.NewRequest(http.MethodPost, "/api/v1/domains", bytes.NewReader(createBody))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	// Delete domain
	req2, _ := http.NewRequest(http.MethodDelete, "/api/v1/domains/example.com", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w2.Code, w2.Body.String())
	}

	var response models.APIResponse
	if err := json.Unmarshal(w2.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Errorf("expected success to be true, got error: %s", response.Error)
	}
}

func TestDeleteDomainNotFound(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/domains/nonexistent.com", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestListRecords(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	// Create domain first
	createBody, _ := json.Marshal(models.CreateDomainRequest{Name: "example.com"})
	req1, _ := http.NewRequest(http.MethodPost, "/api/v1/domains", bytes.NewReader(createBody))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	// List records
	req2, _ := http.NewRequest(http.MethodGet, "/api/v1/domains/example.com/records", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w2.Code, w2.Body.String())
	}

	var response models.APIResponse
	if err := json.Unmarshal(w2.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Errorf("expected success to be true, got error: %s", response.Error)
	}
}

func TestListRecordsNotFound(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/domains/nonexistent.com/records", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestAddRecord(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	// Create domain first
	createBody, _ := json.Marshal(models.CreateDomainRequest{Name: "example.com"})
	req1, _ := http.NewRequest(http.MethodPost, "/api/v1/domains", bytes.NewReader(createBody))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	// Add record
	recordBody, _ := json.Marshal(models.CreateRecordRequest{
		Name:  "api",
		Type:  models.RecordTypeA,
		Value: "192.168.1.100",
		TTL:   3600,
	})
	req2, _ := http.NewRequest(http.MethodPost, "/api/v1/domains/example.com/records", bytes.NewReader(recordBody))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", w2.Code, w2.Body.String())
	}

	var response models.APIResponse
	if err := json.Unmarshal(w2.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Errorf("expected success to be true, got error: %s", response.Error)
	}
}

func TestAddRecordInvalidJSON(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	// Create domain first
	createBody, _ := json.Marshal(models.CreateDomainRequest{Name: "example.com"})
	req1, _ := http.NewRequest(http.MethodPost, "/api/v1/domains", bytes.NewReader(createBody))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	// Add record with invalid JSON
	req2, _ := http.NewRequest(http.MethodPost, "/api/v1/domains/example.com/records", bytes.NewReader([]byte("invalid")))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w2.Code)
	}
}

func TestAddRecordNotFound(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	recordBody, _ := json.Marshal(models.CreateRecordRequest{
		Name:  "api",
		Type:  models.RecordTypeA,
		Value: "192.168.1.100",
	})
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/domains/nonexistent.com/records", bytes.NewReader(recordBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestUpdateRecord(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	// Create domain and add record
	createBody, _ := json.Marshal(models.CreateDomainRequest{Name: "example.com"})
	req1, _ := http.NewRequest(http.MethodPost, "/api/v1/domains", bytes.NewReader(createBody))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	recordBody, _ := json.Marshal(models.CreateRecordRequest{
		Name:  "www",
		Type:  models.RecordTypeA,
		Value: "192.168.1.1",
		TTL:   3600,
	})
	req2, _ := http.NewRequest(http.MethodPost, "/api/v1/domains/example.com/records", bytes.NewReader(recordBody))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	// Update record
	updateBody, _ := json.Marshal(models.UpdateRecordRequest{
		Value: "192.168.1.100",
		TTL:   7200,
	})
	req3, _ := http.NewRequest(http.MethodPut, "/api/v1/domains/example.com/records/www/A", bytes.NewReader(updateBody))
	req3.Header.Set("Content-Type", "application/json")
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)

	if w3.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w3.Code, w3.Body.String())
	}

	var response models.APIResponse
	if err := json.Unmarshal(w3.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Errorf("expected success to be true, got error: %s", response.Error)
	}
}

func TestUpdateRecordNotFound(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	// Create domain first
	createBody, _ := json.Marshal(models.CreateDomainRequest{Name: "example.com"})
	req1, _ := http.NewRequest(http.MethodPost, "/api/v1/domains", bytes.NewReader(createBody))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	// Update non-existent record
	updateBody, _ := json.Marshal(models.UpdateRecordRequest{Value: "192.168.1.1"})
	req2, _ := http.NewRequest(http.MethodPut, "/api/v1/domains/example.com/records/nonexistent/A", bytes.NewReader(updateBody))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w2.Code)
	}
}

func TestDeleteRecord(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	// Create domain and add record
	createBody, _ := json.Marshal(models.CreateDomainRequest{Name: "example.com"})
	req1, _ := http.NewRequest(http.MethodPost, "/api/v1/domains", bytes.NewReader(createBody))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	recordBody, _ := json.Marshal(models.CreateRecordRequest{
		Name:  "temp",
		Type:  models.RecordTypeA,
		Value: "192.168.1.1",
	})
	req2, _ := http.NewRequest(http.MethodPost, "/api/v1/domains/example.com/records", bytes.NewReader(recordBody))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	// Delete record
	req3, _ := http.NewRequest(http.MethodDelete, "/api/v1/domains/example.com/records/temp/A", nil)
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)

	if w3.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w3.Code, w3.Body.String())
	}

	var response models.APIResponse
	if err := json.Unmarshal(w3.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Errorf("expected success to be true, got error: %s", response.Error)
	}
}

func TestDeleteRecordNotFound(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	// Create domain first
	createBody, _ := json.Marshal(models.CreateDomainRequest{Name: "example.com"})
	req1, _ := http.NewRequest(http.MethodPost, "/api/v1/domains", bytes.NewReader(createBody))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	// Delete non-existent record
	req2, _ := http.NewRequest(http.MethodDelete, "/api/v1/domains/example.com/records/nonexistent/A", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w2.Code)
	}
}

func TestReloadZone(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	// Create domain first
	createBody, _ := json.Marshal(models.CreateDomainRequest{Name: "example.com"})
	req1, _ := http.NewRequest(http.MethodPost, "/api/v1/domains", bytes.NewReader(createBody))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	// Try to reload (will fail because rndc doesn't exist)
	req2, _ := http.NewRequest(http.MethodPost, "/api/v1/domains/example.com/reload", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	// Should return 500 because rndc doesn't exist
	if w2.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w2.Code)
	}
}

func TestReloadAll(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	// Try to reload all (will fail because rndc doesn't exist)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/reload", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 500 because rndc doesn't exist
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestAddMXRecord(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	// Create domain first
	createBody, _ := json.Marshal(models.CreateDomainRequest{Name: "example.com"})
	req1, _ := http.NewRequest(http.MethodPost, "/api/v1/domains", bytes.NewReader(createBody))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	// Add MX record with priority
	recordBody, _ := json.Marshal(models.CreateRecordRequest{
		Name:     "@",
		Type:     models.RecordTypeMX,
		Value:    "mail.example.com.",
		Priority: 10,
		TTL:      3600,
	})
	req2, _ := http.NewRequest(http.MethodPost, "/api/v1/domains/example.com/records", bytes.NewReader(recordBody))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", w2.Code, w2.Body.String())
	}
}

func TestAddTXTRecord(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	// Create domain first
	createBody, _ := json.Marshal(models.CreateDomainRequest{Name: "example.com"})
	req1, _ := http.NewRequest(http.MethodPost, "/api/v1/domains", bytes.NewReader(createBody))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	// Add TXT record
	recordBody, _ := json.Marshal(models.CreateRecordRequest{
		Name:  "@",
		Type:  models.RecordTypeTXT,
		Value: "\"v=spf1 include:_spf.google.com ~all\"",
		TTL:   3600,
	})
	req2, _ := http.NewRequest(http.MethodPost, "/api/v1/domains/example.com/records", bytes.NewReader(recordBody))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", w2.Code, w2.Body.String())
	}
}

func TestAddCNAMERecord(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	// Create domain first
	createBody, _ := json.Marshal(models.CreateDomainRequest{Name: "example.com"})
	req1, _ := http.NewRequest(http.MethodPost, "/api/v1/domains", bytes.NewReader(createBody))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	// Add CNAME record
	recordBody, _ := json.Marshal(models.CreateRecordRequest{
		Name:  "blog",
		Type:  models.RecordTypeCNAME,
		Value: "www.example.com.",
		TTL:   3600,
	})
	req2, _ := http.NewRequest(http.MethodPost, "/api/v1/domains/example.com/records", bytes.NewReader(recordBody))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", w2.Code, w2.Body.String())
	}
}

func TestFullWorkflow(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	// 1. Create domain
	createBody, _ := json.Marshal(models.CreateDomainRequest{
		Name: "test.com",
		Nameservers: []string{"ns1.test.com.", "ns2.test.com."},
	})
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/domains", bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("failed to create domain: %s", w.Body.String())
	}

	// 2. Add A record
	recordBody, _ := json.Marshal(models.CreateRecordRequest{
		Name:  "api",
		Type:  models.RecordTypeA,
		Value: "192.168.1.100",
		TTL:   3600,
	})
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/domains/test.com/records", bytes.NewReader(recordBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("failed to add record: %s", w.Body.String())
	}

	// 3. List records
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/domains/test.com/records", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("failed to list records: %s", w.Body.String())
	}

	// 4. Get domain
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/domains/test.com", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("failed to get domain: %s", w.Body.String())
	}

	// 5. List domains
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/domains", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("failed to list domains: %s", w.Body.String())
	}

	// 6. Delete domain
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/domains/test.com", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("failed to delete domain: %s", w.Body.String())
	}
}

func TestAPIResponseFormat(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	// Test success response format
	createBody, _ := json.Marshal(models.CreateDomainRequest{Name: "example.com"})
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/domains", bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response models.APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Success != true {
		t.Error("Success field should be true for successful operations")
	}
	if response.Error != "" {
		t.Error("Error field should be empty for successful operations")
	}
}

func TestRouterRegistration(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	// Test that core routes are registered and accessible
	coreRoutes := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/health"},
		{"GET", "/api/v1/domains"},
		{"POST", "/api/v1/domains"},
		{"POST", "/api/v1/reload"},
	}

	for _, route := range coreRoutes {
		req, _ := http.NewRequest(route.method, route.path, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// We just check that the route exists (not 404)
		// Actual status codes vary based on the operation
		if w.Code == http.StatusNotFound {
			t.Errorf("route %s %s returned 404", route.method, route.path)
		}
	}
}

func TestZoneFileCreation(t *testing.T) {
	router, _, tmpDir := setupTestRouter(t)

	// Create domain
	createBody, _ := json.Marshal(models.CreateDomainRequest{
		Name: "example.com",
		Nameservers: []string{"ns1.example.com."},
	})
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/domains", bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify zone file exists
	zoneFile := filepath.Join(tmpDir, "example.com.zone")
	if _, err := os.Stat(zoneFile); os.IsNotExist(err) {
		t.Fatal("zone file was not created")
	}

	// Verify zone file content
	content, err := os.ReadFile(zoneFile)
	if err != nil {
		t.Fatalf("failed to read zone file: %v", err)
	}

	contentStr := string(content)
	if !contains(contentStr, "example.com") {
		t.Error("zone file should contain domain name")
	}
	if !contains(contentStr, "SOA") {
		t.Error("zone file should contain SOA record")
	}
	if !contains(contentStr, "NS") {
		t.Error("zone file should contain NS record")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
