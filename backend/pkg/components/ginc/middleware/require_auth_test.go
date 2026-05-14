package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAuthByRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	os.Setenv("MANAGER_API_KEY", "manager_token_123")
	os.Setenv("STOREKEEPER_API_KEY", "storekeeper_token_123")

	defer func() {
		os.Unsetenv("MANAGER_API_KEY")
		os.Unsetenv("STOREKEEPER_API_KEY")
	}()

	tests := []struct {
		name              string
		path              string
		authHeader        string
		allowedRoles      []string
		expectedStatus    int
		expectedRole      string
		expectRoleInCtx   bool
		expectBodyKey     string
		expectBodyMessage string
	}{
		// --- Public Paths ---
		{
			name:            "Allow access to /ping without token",
			path:            "/ping",
			authHeader:      "",
			allowedRoles:    []string{"manager"},
			expectedStatus:  http.StatusOK,
			expectRoleInCtx: false,
		},
		{
			name:            "Allow access to /swagger without token",
			path:            "/swagger/index.html",
			authHeader:      "",
			allowedRoles:    []string{"manager"},
			expectedStatus:  http.StatusOK,
			expectRoleInCtx: false,
		},

		// --- Authentication Errors (401) ---
		// Missing header và missing Bearer prefix → cùng message vì middleware check chung 1 điều kiện
		{
			name:              "Error: missing Authorization header",
			path:              "/api/v1/items",
			authHeader:        "",
			allowedRoles:      []string{"manager"},
			expectedStatus:    http.StatusUnauthorized,
			expectRoleInCtx:   false,
			expectBodyKey:     "message",
			expectBodyMessage: "Missing or Invalid token",
		},
		{
			name:              "Error: missing Bearer prefix",
			path:              "/api/v1/items",
			authHeader:        "manager_token_123",
			allowedRoles:      []string{"manager"},
			expectedStatus:    http.StatusUnauthorized,
			expectRoleInCtx:   false,
			expectBodyKey:     "message",
			expectBodyMessage: "Missing or Invalid token",
		},
		// Bearer prefix có nhưng token rỗng → token = "" không match key nào → "Invalid token"
		{
			name:              "Error: Bearer prefix but empty token",
			path:              "/api/v1/items",
			authHeader:        "Bearer ",
			allowedRoles:      []string{"manager"},
			expectedStatus:    http.StatusUnauthorized,
			expectRoleInCtx:   false,
			expectBodyKey:     "message",
			expectBodyMessage: "Invalid token",
		},
		{
			name:              "Error: invalid token",
			path:              "/api/v1/items",
			authHeader:        "Bearer fake_hacker_token",
			allowedRoles:      []string{"manager"},
			expectedStatus:    http.StatusUnauthorized,
			expectRoleInCtx:   false,
			expectBodyKey:     "message",
			expectBodyMessage: "Invalid token",
		},

		// --- Authorization Errors (403) ---
		{
			name:              "403: Storekeeper calls Manager-only API",
			path:              "/api/v1/reports",
			authHeader:        "Bearer storekeeper_token_123",
			allowedRoles:      []string{"manager"},
			expectedStatus:    http.StatusForbidden,
			expectRoleInCtx:   false,
			expectBodyKey:     "message",
			expectBodyMessage: "Permission denied",
		},
		{
			name:              "403: Manager calls Storekeeper-only API",
			path:              "/api/v1/movements",
			authHeader:        "Bearer manager_token_123",
			allowedRoles:      []string{"storekeeper"},
			expectedStatus:    http.StatusForbidden,
			expectRoleInCtx:   false,
			expectBodyKey:     "message",
			expectBodyMessage: "Permission denied",
		},

		// --- Happy Paths (200 OK) ---
		{
			name:            "Success: Manager calls Manager-only API",
			path:            "/api/v1/reports",
			authHeader:      "Bearer manager_token_123",
			allowedRoles:    []string{"manager"},
			expectedStatus:  http.StatusOK,
			expectedRole:    "manager",
			expectRoleInCtx: true,
		},
		{
			name:            "Success: Storekeeper calls Storekeeper-only API",
			path:            "/api/v1/movements",
			authHeader:      "Bearer storekeeper_token_123",
			allowedRoles:    []string{"storekeeper"},
			expectedStatus:  http.StatusOK,
			expectedRole:    "storekeeper",
			expectRoleInCtx: true,
		},
		{
			name:            "Success: Shared API - Manager token",
			path:            "/api/v1/items",
			authHeader:      "Bearer manager_token_123",
			allowedRoles:    []string{"manager", "storekeeper"},
			expectedStatus:  http.StatusOK,
			expectedRole:    "manager",
			expectRoleInCtx: true,
		},
		// ✅ Fix #3: Thêm case storekeeper dùng shared API
		{
			name:            "Success: Shared API - Storekeeper token",
			path:            "/api/v1/items",
			authHeader:      "Bearer storekeeper_token_123",
			allowedRoles:    []string{"manager", "storekeeper"},
			expectedStatus:  http.StatusOK,
			expectedRole:    "storekeeper",
			expectRoleInCtx: true,
		},
		{
			name:            "Success: No allowedRoles restriction - any valid token passes",
			path:            "/api/v1/public-data",
			authHeader:      "Bearer storekeeper_token_123",
			allowedRoles:    []string{},
			expectedStatus:  http.StatusOK,
			expectedRole:    "storekeeper",
			expectRoleInCtx: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(AuthByRole(tt.allowedRoles...))

			r.GET(tt.path, func(c *gin.Context) {
				role, exists := c.Get("user_role")

				if tt.expectRoleInCtx {
					if !exists {
						t.Errorf("Expected context to have user_role but it was not found")
						return
					}
					if role != tt.expectedRole {
						t.Errorf("Expected role = %q, got %q", tt.expectedRole, role)
					}
				} else {
					if exists {
						t.Errorf("Expected user_role to NOT be set in context, but got %q", role)
					}
				}

				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected HTTP Status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectBodyKey != "" {
				var body map[string]any
				if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
					t.Errorf("Expected JSON response body but failed to parse: %v\nBody: %s", err, w.Body.String())
				} else {
					val, ok := body[tt.expectBodyKey]
					if !ok {
						t.Errorf("Expected response body to have key %q, got keys: %v", tt.expectBodyKey, body)
					}
					if tt.expectBodyMessage != "" {
						strVal, isStr := val.(string)
						if !isStr {
							t.Errorf("Expected body[%q] to be a string, got type %T: %v", tt.expectBodyKey, val, val)
						} else if strVal != tt.expectBodyMessage {
							t.Errorf("Expected body[%q] = %q, got %q", tt.expectBodyKey, tt.expectBodyMessage, strVal)
						}
					}
				}
			}
		})
	}
}

func TestAuthByRole_EnvNotSet(t *testing.T) {
	gin.SetMode(gin.TestMode)

	os.Unsetenv("MANAGER_API_KEY")
	os.Unsetenv("STOREKEEPER_API_KEY")

	r := gin.New()
	r.Use(AuthByRole("manager"))
	r.GET("/api/v1/items", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/items", nil)
	req.Header.Set("Authorization", "Bearer manager_token_123")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 when env vars are not set, got %d", w.Code)
	}
}
