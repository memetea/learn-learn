// api/util.go
package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type PaginationMeta struct {
	TotalRecords int64 `json:"total_records"`
	CurrentPage  int   `json:"current_page"`
	PageSize     int   `json:"page_size"`
}

// Response[T]
type Response[T any] struct {
	Status string          `json:"status"`
	Data   T               `json:"data,omitempty"`
	Meta   *PaginationMeta `json:"meta,omitempty"`
} // @name Response

type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// Success writes a successful JSON response with the given data and status code
func Success[T any](w http.ResponseWriter, data T, meta *PaginationMeta, code int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	response := Response[T]{
		Status: "success",
		Data:   data,
		Meta:   meta,
	}
	return json.NewEncoder(w).Encode(response)
}

// Error writes an error JSON response with the given message and status code
func Error(w http.ResponseWriter, message string, code int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	response := ErrorResponse{
		Status:  "error",
		Message: message,
	}
	return json.NewEncoder(w).Encode(response)
}

type APIEndpoint struct {
	Path                  string
	Method                string
	Handler               http.HandlerFunc
	Permission            string // 对应的权限标识
	PermissionDescription string
	// Menu                  *MenuItem
}

type APIEndpointProvider interface {
	GetApiEndpoints() []APIEndpoint
}

// DecodeJSONBody decodes a JSON request body into the specified type.
func DecodeJSONBody[T any](w http.ResponseWriter, r *http.Request) (*T, bool) {
	var body T
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		Error(w, "Invalid request payload", http.StatusBadRequest)
		return nil, false
	}
	return &body, true
}

// ParseUintParam parses a uint parameter from the URL.
func ParseUintParam(r *http.Request, param string) (uint, bool) {
	vars := mux.Vars(r)
	paramStr := vars[param]
	value, err := strconv.ParseUint(paramStr, 10, 32)
	if err != nil {
		return 0, false
	}
	return uint(value), true
}

func GetPaginationParams(r *http.Request) (page, pageSize int) {
	page = parseQueryParamInt(r, "page", 1)
	pageSize = parseQueryParamInt(r, "page_size", 10)
	return
}

func parseQueryParamInt(r *http.Request, key string, defaultValue int) int {
	valueStr := r.URL.Query().Get(key)
	value, err := strconv.Atoi(valueStr)
	if err != nil || value <= 0 {
		return defaultValue
	}
	return value
}
