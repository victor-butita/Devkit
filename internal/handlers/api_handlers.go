package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/victor-butita/devkit/internal/services" // Use your module path
)

type APIHandlers struct {
	mockStore    *services.MockStore
	geminiSvc    *services.GeminiService
}

func NewAPIHandlers(ms *services.MockStore, gs *services.GeminiService) *APIHandlers {
	return &APIHandlers{
		mockStore:    ms,
		geminiSvc:    gs,
	}
}

// --- Helper for responses ---
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// --- Mockify Handlers ---
func (h *APIHandlers) HandleCreateMock(w http.ResponseWriter, r *http.Request) {
	var reqBody map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	bodyBytes, _ := json.Marshal(reqBody)

	id, err := h.mockStore.CreateMock(string(bodyBytes))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create mock")
		return
	}
	
	// Construct the full URL to return to the user
	scheme := "http"
	if r.TLS != nil { scheme = "https" }
	mockURL := scheme + "://" + r.Host + "/mock/" + id

	writeJSON(w, http.StatusCreated, map[string]string{"url": mockURL, "id": id})
}

func (h *APIHandlers) HandleGetMock(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	mockData, found := h.mockStore.GetMock(id)
	if !found {
		writeError(w, http.StatusNotFound, "Mock not found")
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(mockData))
}

// --- RegexCraft Handler ---
func (h *APIHandlers) HandleGenerateRegex(w http.ResponseWriter, r *http.Request) {
	var req struct { Description string `json:"description"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Description == "" {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	result, err := h.geminiSvc.GenerateRegex(req.Description)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	parts := strings.SplitN(result, "|||", 2)
	if len(parts) != 2 {
		writeJSON(w, http.StatusOK, map[string]string{"regex": result, "explanation": "AI did not provide an explanation."})
		return
	}
	
	writeJSON(w, http.StatusOK, map[string]string{"regex": strings.TrimSpace(parts[0]), "explanation": strings.TrimSpace(parts[1])})
}

// --- ConfigSwitch Handler ---
func (h *APIHandlers) HandleConvertConfig(w http.ResponseWriter, r *http.Request) {
	var req struct { Input string `json:"input"`; From string `json:"from"`; To string `json:"to"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Input == "" {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	output, err := services.ConvertConfig(req.Input, req.From, req.To)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"output": output})
}

// --- QueryGen Handler ---
func (h *APIHandlers) HandleGenerateSQL(w http.ResponseWriter, r *http.Request) {
	var req struct { Schema string `json:"schema"`; Description string `json:"description"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Description == "" {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	result, err := h.geminiSvc.GenerateSQL(req.Schema, req.Description)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"query": strings.TrimSpace(result)})
}

// --- JSON Beautifier Handler ---
func (h *APIHandlers) HandleFormatJSON(w http.ResponseWriter, r *http.Request) {
	var raw json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON input")
		return
	}

	var indented bytes.Buffer
	if err := json.Indent(&indented, raw, "", "  "); err != nil {
		writeError(w, http.StatusBadRequest, "Failed to format JSON")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"formatted_json": indented.String()})
}