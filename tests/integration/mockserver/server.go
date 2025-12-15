// Package mockserver provides a mock freee API server for integration testing.
//
// The mock server simulates the freee API responses, allowing for end-to-end
// testing without connecting to the actual API. It supports:
//   - OAuth2 token exchange
//   - Deals CRUD operations
//   - Pagination
//   - Error responses
//
// Example:
//
//	server := mockserver.NewServer()
//	defer server.Close()
//
//	// Use server.URL as the base URL for the client
//	client := client.NewClient(
//	    client.WithBaseURL(server.URL),
//	    client.WithTokenSource(tokenSource),
//	)
package mockserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Server is a mock freee API server for testing.
type Server struct {
	// HTTPServer is the underlying httptest.Server.
	HTTPServer *httptest.Server

	// URL is the base URL of the mock server.
	URL string

	// mu protects the internal state.
	mu sync.RWMutex

	// tokens stores valid access tokens.
	tokens map[string]bool

	// deals stores mock deals data.
	deals map[int64]*Deal

	// journals stores mock journals data.
	journals map[int64]*Journal

	// walletTxns stores mock wallet transactions data.
	walletTxns map[int64]*WalletTxn

	// transfers stores mock transfers data.
	transfers map[int64]*Transfer

	// nextDealID is the next ID for creating deals.
	nextDealID int64

	// errorMode can be set to simulate error responses.
	errorMode ErrorMode

	// requestLog stores all received requests for verification.
	requestLog []*Request
}

// ErrorMode defines the type of error to simulate.
type ErrorMode int

const (
	// ErrorModeNone indicates no error simulation.
	ErrorModeNone ErrorMode = iota

	// ErrorModeUnauthorized simulates 401 Unauthorized.
	ErrorModeUnauthorized

	// ErrorModeBadRequest simulates 400 Bad Request.
	ErrorModeBadRequest

	// ErrorModeNotFound simulates 404 Not Found.
	ErrorModeNotFound

	// ErrorModeInternalServer simulates 500 Internal Server Error.
	ErrorModeInternalServer

	// ErrorModeTooManyRequests simulates 429 Too Many Requests.
	ErrorModeTooManyRequests

	// ErrorModeServiceUnavailable simulates 503 Service Unavailable.
	ErrorModeServiceUnavailable
)

// Request represents a logged HTTP request.
type Request struct {
	Method  string
	Path    string
	Headers http.Header
	Body    string
}

// Deal represents a mock deal.
type Deal struct {
	ID         int64          `json:"id"`
	CompanyID  int64          `json:"company_id"`
	IssueDate  string         `json:"issue_date"`
	DueDate    *string        `json:"due_date,omitempty"`
	Amount     int64          `json:"amount"`
	DueAmount  int64          `json:"due_amount"`
	Type       string         `json:"type"`
	PartnerID  *int64         `json:"partner_id,omitempty"`
	RefNumber  *string        `json:"ref_number,omitempty"`
	Status     string         `json:"status"`
	Details    []DealDetail   `json:"details"`
	Payments   []DealPayment  `json:"payments"`
	Receipts   []DealReceipt  `json:"receipts"`
	CreateTime string         `json:"created_at"`
	UpdateTime string         `json:"updated_at"`
}

// DealDetail represents a deal detail line.
type DealDetail struct {
	ID            int64   `json:"id"`
	AccountItemID int64   `json:"account_item_id"`
	TaxCode       int     `json:"tax_code"`
	Amount        int64   `json:"amount"`
	Vat           int64   `json:"vat"`
	Description   *string `json:"description,omitempty"`
	EntryType     string  `json:"entry_side"`
}

// DealPayment represents a deal payment.
type DealPayment struct {
	ID            int64  `json:"id"`
	Date          string `json:"date"`
	FromWalletID  int64  `json:"from_walletable_id"`
	FromWalletType string `json:"from_walletable_type"`
	Amount        int64  `json:"amount"`
}

// DealReceipt represents a deal receipt.
type DealReceipt struct {
	ID int64 `json:"id"`
}

// Journal represents a mock journal entry.
type Journal struct {
	ID        int64  `json:"id"`
	CompanyID int64  `json:"company_id"`
}

// WalletTxn represents a mock wallet transaction.
type WalletTxn struct {
	ID        int64  `json:"id"`
	CompanyID int64  `json:"company_id"`
}

// Transfer represents a mock transfer.
type Transfer struct {
	ID        int64  `json:"id"`
	CompanyID int64  `json:"company_id"`
}

// NewServer creates a new mock freee API server.
func NewServer() *Server {
	s := &Server{
		tokens:     make(map[string]bool),
		deals:      make(map[int64]*Deal),
		journals:   make(map[int64]*Journal),
		walletTxns: make(map[int64]*WalletTxn),
		transfers:  make(map[int64]*Transfer),
		nextDealID: 1,
		requestLog: make([]*Request, 0),
	}

	// Add a default valid token
	s.tokens["test-access-token"] = true

	// Create the HTTP server
	mux := http.NewServeMux()
	s.registerRoutes(mux)
	s.HTTPServer = httptest.NewServer(mux)
	s.URL = s.HTTPServer.URL

	return s
}

// Close shuts down the mock server.
func (s *Server) Close() {
	s.HTTPServer.Close()
}

// SetErrorMode sets the error mode for the server.
func (s *Server) SetErrorMode(mode ErrorMode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.errorMode = mode
}

// ClearErrorMode clears any error mode.
func (s *Server) ClearErrorMode() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.errorMode = ErrorModeNone
}

// AddToken adds a valid access token.
func (s *Server) AddToken(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[token] = true
}

// RemoveToken removes an access token.
func (s *Server) RemoveToken(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tokens, token)
}

// AddDeal adds a mock deal.
func (s *Server) AddDeal(deal *Deal) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if deal.ID == 0 {
		deal.ID = s.nextDealID
		s.nextDealID++
	}
	s.deals[deal.ID] = deal
}

// GetDeals returns all mock deals.
func (s *Server) GetDeals() map[int64]*Deal {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[int64]*Deal)
	for k, v := range s.deals {
		result[k] = v
	}
	return result
}

// ClearDeals clears all mock deals.
func (s *Server) ClearDeals() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deals = make(map[int64]*Deal)
	s.nextDealID = 1
}

// GetRequestLog returns the request log.
func (s *Server) GetRequestLog() []*Request {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Request, len(s.requestLog))
	copy(result, s.requestLog)
	return result
}

// ClearRequestLog clears the request log.
func (s *Server) ClearRequestLog() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.requestLog = make([]*Request, 0)
}

// registerRoutes registers all API routes.
func (s *Server) registerRoutes(mux *http.ServeMux) {
	// OAuth2 endpoints
	mux.HandleFunc("/public_api/token", s.handleToken)

	// API v1 endpoints
	mux.HandleFunc("/api/1/deals", s.handleDeals)
	mux.HandleFunc("/api/1/deals/", s.handleDeal)
	mux.HandleFunc("/api/1/journals", s.handleJournals)
	mux.HandleFunc("/api/1/wallet_txns", s.handleWalletTxns)
	mux.HandleFunc("/api/1/transfers", s.handleTransfers)
	mux.HandleFunc("/api/1/users/me", s.handleUsersMe)
	mux.HandleFunc("/api/1/companies", s.handleCompanies)
}

// logRequest logs an incoming request.
func (s *Server) logRequest(r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	req := &Request{
		Method:  r.Method,
		Path:    r.URL.Path,
		Headers: r.Header.Clone(),
	}
	s.requestLog = append(s.requestLog, req)
}

// checkAuth checks if the request is authenticated.
func (s *Server) checkAuth(w http.ResponseWriter, r *http.Request) bool {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		s.writeError(w, http.StatusUnauthorized, "unauthorized", "Authorization header required")
		return false
	}

	token := strings.TrimPrefix(auth, "Bearer ")
	s.mu.RLock()
	valid := s.tokens[token]
	s.mu.RUnlock()

	if !valid {
		s.writeError(w, http.StatusUnauthorized, "unauthorized", "Invalid access token")
		return false
	}

	return true
}

// checkErrorMode checks if an error mode is set and writes the error response.
func (s *Server) checkErrorMode(w http.ResponseWriter) bool {
	s.mu.RLock()
	mode := s.errorMode
	s.mu.RUnlock()

	switch mode {
	case ErrorModeUnauthorized:
		s.writeError(w, http.StatusUnauthorized, "unauthorized", "Simulated unauthorized error")
		return true
	case ErrorModeBadRequest:
		s.writeError(w, http.StatusBadRequest, "bad_request", "Simulated bad request error")
		return true
	case ErrorModeNotFound:
		s.writeError(w, http.StatusNotFound, "not_found", "Simulated not found error")
		return true
	case ErrorModeInternalServer:
		s.writeError(w, http.StatusInternalServerError, "internal_server_error", "Simulated internal server error")
		return true
	case ErrorModeTooManyRequests:
		s.writeError(w, http.StatusTooManyRequests, "too_many_requests", "Simulated rate limit error")
		return true
	case ErrorModeServiceUnavailable:
		s.writeError(w, http.StatusServiceUnavailable, "service_unavailable", "Simulated service unavailable error")
		return true
	}

	return false
}

// writeError writes an error response.
func (s *Server) writeError(w http.ResponseWriter, status int, errType, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := map[string]interface{}{
		"status_code": status,
		"errors": []map[string]interface{}{
			{
				"type":     errType,
				"messages": []string{message},
			},
		},
	}
	json.NewEncoder(w).Encode(resp)
}

// writeJSON writes a JSON response.
func (s *Server) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// handleToken handles OAuth2 token exchange.
func (s *Server) handleToken(w http.ResponseWriter, r *http.Request) {
	s.logRequest(r)

	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	if err := r.ParseForm(); err != nil {
		s.writeError(w, http.StatusBadRequest, "bad_request", "Invalid form data")
		return
	}

	grantType := r.FormValue("grant_type")
	switch grantType {
	case "authorization_code":
		code := r.FormValue("code")
		if code == "" {
			s.writeError(w, http.StatusBadRequest, "bad_request", "Missing authorization code")
			return
		}
		// Return a mock token
		s.writeJSON(w, http.StatusOK, map[string]interface{}{
			"access_token":  "mock-access-token-" + time.Now().Format("20060102150405"),
			"token_type":    "Bearer",
			"expires_in":    86400,
			"refresh_token": "mock-refresh-token",
			"scope":         "read write",
			"created_at":    time.Now().Unix(),
		})

	case "refresh_token":
		refreshToken := r.FormValue("refresh_token")
		if refreshToken == "" {
			s.writeError(w, http.StatusBadRequest, "bad_request", "Missing refresh token")
			return
		}
		// Return a new mock token
		s.writeJSON(w, http.StatusOK, map[string]interface{}{
			"access_token":  "mock-access-token-refreshed-" + time.Now().Format("20060102150405"),
			"token_type":    "Bearer",
			"expires_in":    86400,
			"refresh_token": "mock-refresh-token-new",
			"scope":         "read write",
			"created_at":    time.Now().Unix(),
		})

	default:
		s.writeError(w, http.StatusBadRequest, "bad_request", "Unsupported grant type")
	}
}

// handleDeals handles /api/1/deals endpoint.
func (s *Server) handleDeals(w http.ResponseWriter, r *http.Request) {
	s.logRequest(r)

	if s.checkErrorMode(w) {
		return
	}

	if !s.checkAuth(w, r) {
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleListDeals(w, r)
	case http.MethodPost:
		s.handleCreateDeal(w, r)
	default:
		s.writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
	}
}

// handleDeal handles /api/1/deals/{id} endpoint.
func (s *Server) handleDeal(w http.ResponseWriter, r *http.Request) {
	s.logRequest(r)

	if s.checkErrorMode(w) {
		return
	}

	if !s.checkAuth(w, r) {
		return
	}

	// Extract deal ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/1/deals/")
	dealID, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "bad_request", "Invalid deal ID")
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGetDeal(w, r, dealID)
	case http.MethodPut:
		s.handleUpdateDeal(w, r, dealID)
	case http.MethodDelete:
		s.handleDeleteDeal(w, r, dealID)
	default:
		s.writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
	}
}

// handleListDeals handles GET /api/1/deals.
func (s *Server) handleListDeals(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	query := r.URL.Query()
	companyID := query.Get("company_id")
	if companyID == "" {
		s.writeError(w, http.StatusBadRequest, "bad_request", "company_id is required")
		return
	}

	offset, _ := strconv.ParseInt(query.Get("offset"), 10, 64)
	limit, _ := strconv.ParseInt(query.Get("limit"), 10, 64)
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	dealType := query.Get("type")
	status := query.Get("status")

	s.mu.RLock()
	var deals []*Deal
	for _, deal := range s.deals {
		// Filter by type if specified
		if dealType != "" && deal.Type != dealType {
			continue
		}
		// Filter by status if specified
		if status != "" && deal.Status != status {
			continue
		}
		deals = append(deals, deal)
	}
	s.mu.RUnlock()

	// Apply pagination
	totalCount := int64(len(deals))
	if offset >= totalCount {
		deals = []*Deal{}
	} else {
		end := offset + limit
		if end > totalCount {
			end = totalCount
		}
		deals = deals[offset:end]
	}

	// Build response
	resp := map[string]interface{}{
		"deals": deals,
		"meta": map[string]interface{}{
			"total_count": totalCount,
		},
	}

	s.writeJSON(w, http.StatusOK, resp)
}

// handleGetDeal handles GET /api/1/deals/{id}.
func (s *Server) handleGetDeal(w http.ResponseWriter, r *http.Request, dealID int64) {
	s.mu.RLock()
	deal, ok := s.deals[dealID]
	s.mu.RUnlock()

	if !ok {
		s.writeError(w, http.StatusNotFound, "not_found", fmt.Sprintf("Deal %d not found", dealID))
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"deal": deal,
	})
}

// handleCreateDeal handles POST /api/1/deals.
func (s *Server) handleCreateDeal(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CompanyID int64        `json:"company_id"`
		IssueDate string       `json:"issue_date"`
		DueDate   *string      `json:"due_date,omitempty"`
		Type      string       `json:"type"`
		PartnerID *int64       `json:"partner_id,omitempty"`
		RefNumber *string      `json:"ref_number,omitempty"`
		Details   []DealDetail `json:"details"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "bad_request", "Invalid JSON")
		return
	}

	// Validate required fields
	if req.CompanyID == 0 {
		s.writeError(w, http.StatusBadRequest, "validation", "company_id is required")
		return
	}
	if req.IssueDate == "" {
		s.writeError(w, http.StatusBadRequest, "validation", "issue_date is required")
		return
	}
	if req.Type == "" {
		s.writeError(w, http.StatusBadRequest, "validation", "type is required")
		return
	}
	if len(req.Details) == 0 {
		s.writeError(w, http.StatusBadRequest, "validation", "details is required")
		return
	}

	// Calculate total amount
	var totalAmount int64
	for _, detail := range req.Details {
		totalAmount += detail.Amount
	}

	// Create the deal
	s.mu.Lock()
	deal := &Deal{
		ID:         s.nextDealID,
		CompanyID:  req.CompanyID,
		IssueDate:  req.IssueDate,
		DueDate:    req.DueDate,
		Amount:     totalAmount,
		DueAmount:  totalAmount,
		Type:       req.Type,
		PartnerID:  req.PartnerID,
		RefNumber:  req.RefNumber,
		Status:     "unsettled",
		Details:    req.Details,
		Payments:   []DealPayment{},
		Receipts:   []DealReceipt{},
		CreateTime: time.Now().Format(time.RFC3339),
		UpdateTime: time.Now().Format(time.RFC3339),
	}
	s.deals[deal.ID] = deal
	s.nextDealID++
	s.mu.Unlock()

	s.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"deal": deal,
	})
}

// handleUpdateDeal handles PUT /api/1/deals/{id}.
func (s *Server) handleUpdateDeal(w http.ResponseWriter, r *http.Request, dealID int64) {
	s.mu.Lock()
	deal, ok := s.deals[dealID]
	if !ok {
		s.mu.Unlock()
		s.writeError(w, http.StatusNotFound, "not_found", fmt.Sprintf("Deal %d not found", dealID))
		return
	}

	var req struct {
		CompanyID int64        `json:"company_id"`
		IssueDate string       `json:"issue_date"`
		DueDate   *string      `json:"due_date,omitempty"`
		Type      string       `json:"type"`
		PartnerID *int64       `json:"partner_id,omitempty"`
		RefNumber *string      `json:"ref_number,omitempty"`
		Details   []DealDetail `json:"details"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.mu.Unlock()
		s.writeError(w, http.StatusBadRequest, "bad_request", "Invalid JSON")
		return
	}

	// Update fields
	if req.IssueDate != "" {
		deal.IssueDate = req.IssueDate
	}
	if req.DueDate != nil {
		deal.DueDate = req.DueDate
	}
	if req.Type != "" {
		deal.Type = req.Type
	}
	if req.PartnerID != nil {
		deal.PartnerID = req.PartnerID
	}
	if req.RefNumber != nil {
		deal.RefNumber = req.RefNumber
	}
	if len(req.Details) > 0 {
		deal.Details = req.Details
		// Recalculate amount
		var totalAmount int64
		for _, detail := range req.Details {
			totalAmount += detail.Amount
		}
		deal.Amount = totalAmount
		deal.DueAmount = totalAmount
	}
	deal.UpdateTime = time.Now().Format(time.RFC3339)
	s.mu.Unlock()

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"deal": deal,
	})
}

// handleDeleteDeal handles DELETE /api/1/deals/{id}.
func (s *Server) handleDeleteDeal(w http.ResponseWriter, r *http.Request, dealID int64) {
	s.mu.Lock()
	_, ok := s.deals[dealID]
	if !ok {
		s.mu.Unlock()
		s.writeError(w, http.StatusNotFound, "not_found", fmt.Sprintf("Deal %d not found", dealID))
		return
	}
	delete(s.deals, dealID)
	s.mu.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

// handleJournals handles /api/1/journals endpoint.
func (s *Server) handleJournals(w http.ResponseWriter, r *http.Request) {
	s.logRequest(r)

	if s.checkErrorMode(w) {
		return
	}

	if !s.checkAuth(w, r) {
		return
	}

	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	// Get query parameters
	query := r.URL.Query()
	companyID := query.Get("company_id")
	if companyID == "" {
		s.writeError(w, http.StatusBadRequest, "bad_request", "company_id is required")
		return
	}

	offset, _ := strconv.ParseInt(query.Get("offset"), 10, 64)
	limit, _ := strconv.ParseInt(query.Get("limit"), 10, 64)
	if limit <= 0 {
		limit = 20
	}

	s.mu.RLock()
	var journals []*Journal
	for _, journal := range s.journals {
		journals = append(journals, journal)
	}
	s.mu.RUnlock()

	// Apply pagination
	totalCount := int64(len(journals))
	if offset >= totalCount {
		journals = []*Journal{}
	} else {
		end := offset + limit
		if end > totalCount {
			end = totalCount
		}
		journals = journals[offset:end]
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"journals": journals,
		"meta": map[string]interface{}{
			"total_count": totalCount,
		},
	})
}

// handleWalletTxns handles /api/1/wallet_txns endpoint.
func (s *Server) handleWalletTxns(w http.ResponseWriter, r *http.Request) {
	s.logRequest(r)

	if s.checkErrorMode(w) {
		return
	}

	if !s.checkAuth(w, r) {
		return
	}

	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	// Get query parameters
	query := r.URL.Query()
	companyID := query.Get("company_id")
	if companyID == "" {
		s.writeError(w, http.StatusBadRequest, "bad_request", "company_id is required")
		return
	}

	offset, _ := strconv.ParseInt(query.Get("offset"), 10, 64)
	limit, _ := strconv.ParseInt(query.Get("limit"), 10, 64)
	if limit <= 0 {
		limit = 20
	}

	s.mu.RLock()
	var txns []*WalletTxn
	for _, txn := range s.walletTxns {
		txns = append(txns, txn)
	}
	s.mu.RUnlock()

	// Apply pagination
	totalCount := int64(len(txns))
	if offset >= totalCount {
		txns = []*WalletTxn{}
	} else {
		end := offset + limit
		if end > totalCount {
			end = totalCount
		}
		txns = txns[offset:end]
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"wallet_txns": txns,
		"meta": map[string]interface{}{
			"total_count": totalCount,
		},
	})
}

// handleTransfers handles /api/1/transfers endpoint.
func (s *Server) handleTransfers(w http.ResponseWriter, r *http.Request) {
	s.logRequest(r)

	if s.checkErrorMode(w) {
		return
	}

	if !s.checkAuth(w, r) {
		return
	}

	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	// Get query parameters
	query := r.URL.Query()
	companyID := query.Get("company_id")
	if companyID == "" {
		s.writeError(w, http.StatusBadRequest, "bad_request", "company_id is required")
		return
	}

	offset, _ := strconv.ParseInt(query.Get("offset"), 10, 64)
	limit, _ := strconv.ParseInt(query.Get("limit"), 10, 64)
	if limit <= 0 {
		limit = 20
	}

	s.mu.RLock()
	var transfers []*Transfer
	for _, transfer := range s.transfers {
		transfers = append(transfers, transfer)
	}
	s.mu.RUnlock()

	// Apply pagination
	totalCount := int64(len(transfers))
	if offset >= totalCount {
		transfers = []*Transfer{}
	} else {
		end := offset + limit
		if end > totalCount {
			end = totalCount
		}
		transfers = transfers[offset:end]
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"transfers": transfers,
		"meta": map[string]interface{}{
			"total_count": totalCount,
		},
	})
}

// handleUsersMe handles /api/1/users/me endpoint.
func (s *Server) handleUsersMe(w http.ResponseWriter, r *http.Request) {
	s.logRequest(r)

	if s.checkErrorMode(w) {
		return
	}

	if !s.checkAuth(w, r) {
		return
	}

	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"user": map[string]interface{}{
			"id":           1,
			"email":        "test@example.com",
			"display_name": "Test User",
			"companies": []map[string]interface{}{
				{
					"id":           1,
					"display_name": "Test Company",
					"role":         "admin",
				},
			},
		},
	})
}

// handleCompanies handles /api/1/companies endpoint.
func (s *Server) handleCompanies(w http.ResponseWriter, r *http.Request) {
	s.logRequest(r)

	if s.checkErrorMode(w) {
		return
	}

	if !s.checkAuth(w, r) {
		return
	}

	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"companies": []map[string]interface{}{
			{
				"id":                1,
				"display_name":      "Test Company",
				"role":              "admin",
				"use_partner_code":  true,
				"use_account_items": true,
			},
		},
	})
}
