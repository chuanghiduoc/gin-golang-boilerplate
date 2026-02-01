package middleware

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSecurity_DefaultHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Security())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "SAMEORIGIN", w.Header().Get("X-Frame-Options"))
	assert.Contains(t, w.Header().Get("Strict-Transport-Security"), "max-age=")
	assert.Equal(t, "default-src 'self'", w.Header().Get("Content-Security-Policy"))
	assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
	assert.Equal(t, "none", w.Header().Get("X-Permitted-Cross-Domain-Policies"))
	assert.Equal(t, "noopen", w.Header().Get("X-Download-Options"))
}

func TestSecurity_CustomConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	customConfig := SecurityConfig{
		XSSProtection:         "0",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		HSTSMaxAge:            86400,
		HSTSIncludeSubdomains: false,
		ContentSecurityPolicy: "default-src 'self'; script-src 'self'",
		ReferrerPolicy:        "no-referrer",
		PermissionsPolicy:     "geolocation=()",
	}

	router := gin.New()
	router.Use(SecurityWithConfig(customConfig))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "0", w.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "default-src 'self'; script-src 'self'", w.Header().Get("Content-Security-Policy"))
	assert.Equal(t, "no-referrer", w.Header().Get("Referrer-Policy"))
}

func TestCORS_AllowOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CORS())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_PreflightRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CORS())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
}

func TestGetClientIP_XForwardedFor(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		ip := getClientIP(c)
		c.String(http.StatusOK, ip)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "192.168.1.1", w.Body.String())
}

func TestGetClientIP_XRealIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		ip := getClientIP(c)
		c.String(http.StatusOK, ip)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Real-IP", "10.0.0.1")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "10.0.0.1", w.Body.String())
}

func TestI18n_AcceptLanguageHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(I18n())
	router.GET("/test", func(c *gin.Context) {
		lang, _ := c.Get(string(LangKey))
		c.String(http.StatusOK, lang.(string))
	})

	tests := []struct {
		name           string
		acceptLanguage string
		expectedLang   string
	}{
		{
			name:           "Vietnamese",
			acceptLanguage: "vi-VN,vi;q=0.9",
			expectedLang:   "vi",
		},
		{
			name:           "English",
			acceptLanguage: "en-US,en;q=0.9",
			expectedLang:   "en",
		},
		{
			name:           "Default to English",
			acceptLanguage: "fr-FR,fr;q=0.9",
			expectedLang:   "en",
		},
		{
			name:           "No header defaults to English",
			acceptLanguage: "",
			expectedLang:   "en",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", nil)
			if tt.acceptLanguage != "" {
				req.Header.Set("Accept-Language", tt.acceptLanguage)
			}
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, tt.expectedLang, w.Body.String())
		})
	}
}

func TestI18n_QueryParameter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(I18n())
	router.GET("/test", func(c *gin.Context) {
		lang, _ := c.Get(string(LangKey))
		c.String(http.StatusOK, lang.(string))
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test?lang=vi", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "vi", w.Body.String())
}

func TestSanitizeQuery(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:     "empty query",
			input:    "",
			expected: map[string]string{},
		},
		{
			name:  "no sensitive params",
			input: "page=1&limit=10&search=test",
			expected: map[string]string{
				"page":   "1",
				"limit":  "10",
				"search": "test",
			},
		},
		{
			name:  "token param",
			input: "token=secret123&page=1",
			expected: map[string]string{
				"token": "[REDACTED]",
				"page":  "1",
			},
		},
		{
			name:  "access_token param",
			input: "access_token=abc123",
			expected: map[string]string{
				"access_token": "[REDACTED]",
			},
		},
		{
			name:  "password param",
			input: "user=admin&password=secret",
			expected: map[string]string{
				"user":     "admin",
				"password": "[REDACTED]",
			},
		},
		{
			name:  "api_key param",
			input: "api_key=key123&format=json",
			expected: map[string]string{
				"api_key": "[REDACTED]",
				"format":  "json",
			},
		},
		{
			name:  "mixed case sensitivity",
			input: "TOKEN=secret&Password=pass&API_KEY=key",
			expected: map[string]string{
				"TOKEN":    "[REDACTED]",
				"Password": "[REDACTED]",
				"API_KEY":  "[REDACTED]",
			},
		},
		{
			name:  "multiple sensitive params",
			input: "token=t1&secret=s1&key=k1&session=sess1",
			expected: map[string]string{
				"token":   "[REDACTED]",
				"secret":  "[REDACTED]",
				"key":     "[REDACTED]",
				"session": "[REDACTED]",
			},
		},
		{
			name:  "refresh_token param",
			input: "refresh_token=ref123",
			expected: map[string]string{
				"refresh_token": "[REDACTED]",
			},
		},
		{
			name:  "authorization param",
			input: "authorization=Bearer+token123",
			expected: map[string]string{
				"authorization": "[REDACTED]",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeQuery(tt.input)

			if tt.input == "" {
				assert.Empty(t, result)
				return
			}

			// Parse the result using url.ParseQuery to handle URL encoding
			parsed, err := url.ParseQuery(result)
			assert.NoError(t, err)

			for key, expectedValue := range tt.expected {
				// Case-insensitive key lookup
				var found bool
				for k, values := range parsed {
					if equalFoldKey(k, key) {
						if len(values) > 0 {
							assert.Equal(t, expectedValue, values[0], "Key %s should have value %s", key, expectedValue)
						}
						found = true
						break
					}
				}
				assert.True(t, found, "Key %s should exist in result", key)
			}
		})
	}
}

func equalFoldKey(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}
