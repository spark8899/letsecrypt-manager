package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-jwt-secret-key"

func init() {
	gin.SetMode(gin.TestMode)
}

// ---- GenerateToken ----

func TestGenerateToken_ValidToken(t *testing.T) {
	token, err := GenerateToken("alice", testSecret, 24)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
}

func TestGenerateToken_ParseableClaims(t *testing.T) {
	token, _ := GenerateToken("bob", testSecret, 24)

	claims := &Claims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(testSecret), nil
	})
	if err != nil {
		t.Fatalf("ParseWithClaims() error = %v", err)
	}
	if !parsed.Valid {
		t.Error("expected valid token")
	}
	if claims.Username != "bob" {
		t.Errorf("Username = %s, want bob", claims.Username)
	}
}

func TestGenerateToken_ExpiryIsCorrect(t *testing.T) {
	before := time.Now().Truncate(time.Second)
	token, _ := GenerateToken("user", testSecret, 2)

	claims := &Claims{}
	jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(testSecret), nil
	})

	expiry := claims.ExpiresAt.Time
	// JWT NumericDate has second-level precision; allow ±2s tolerance
	expectedMin := before.Add(2*time.Hour - 2*time.Second)
	expectedMax := before.Add(2*time.Hour + 2*time.Second)

	if expiry.Before(expectedMin) || expiry.After(expectedMax) {
		t.Errorf("expiry %v outside expected range [%v, %v]", expiry, expectedMin, expectedMax)
	}
}

func TestGenerateToken_DifferentSecretsFail(t *testing.T) {
	token, _ := GenerateToken("user", testSecret, 24)

	claims := &Claims{}
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte("wrong-secret"), nil
	})
	if err == nil {
		t.Error("expected error parsing token with wrong secret")
	}
}

// ---- AuthMiddleware ----

func setupRouter(secret string) *gin.Engine {
	r := gin.New()
	protected := r.Group("/", AuthMiddleware(secret))
	protected.GET("/protected", func(c *gin.Context) {
		username := c.GetString("username")
		c.JSON(http.StatusOK, gin.H{"username": username})
	})
	return r
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	r := setupRouter(testSecret)
	token, _ := GenerateToken("alice", testSecret, 24)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	r := setupRouter(testSecret)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestAuthMiddleware_InvalidBearerFormat(t *testing.T) {
	r := setupRouter(testSecret)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Token sometoken")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	r := setupRouter(testSecret)

	// Create a token that's already expired
	claims := Claims{
		Username: "expired_user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte(testSecret))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestAuthMiddleware_WrongSecret(t *testing.T) {
	r := setupRouter(testSecret)
	token, _ := GenerateToken("user", "different-secret", 24)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestAuthMiddleware_TamperedToken(t *testing.T) {
	r := setupRouter(testSecret)
	token, _ := GenerateToken("user", testSecret, 24)
	tampered := token + "tampered"

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tampered)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestAuthMiddleware_BearerCaseInsensitive(t *testing.T) {
	r := setupRouter(testSecret)
	token, _ := GenerateToken("user", testSecret, 24)

	for _, prefix := range []string{"Bearer", "bearer", "BEARER"} {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", prefix+" "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("prefix %q: status = %d, want 200", prefix, w.Code)
		}
	}
}

func TestAuthMiddleware_SetsUsernameInContext(t *testing.T) {
	r := setupRouter(testSecret)
	token, _ := GenerateToken("testuser", testSecret, 24)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	body := w.Body.String()
	if body == "" {
		t.Fatal("empty response body")
	}
	// body should contain the username
	if !containsString(body, "testuser") {
		t.Errorf("response body %q should contain 'testuser'", body)
	}
}

func containsString(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && stringContains(s, sub))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
