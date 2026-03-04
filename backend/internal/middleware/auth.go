package middleware

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type userIDKey struct{}

// ---- JWKS (Auth0 RS256) ----

type jwkKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type jwksResponse struct {
	Keys []jwkKey `json:"keys"`
}

// Auth0KeyFunc fetches and caches Auth0 public keys.
type Auth0KeyFunc struct {
	mu        sync.RWMutex
	keys      map[string]*rsa.PublicKey
	jwksURL   string
	audience  string
	lastFetch time.Time
}

func NewAuth0KeyFunc(domain, audience string) *Auth0KeyFunc {
	return &Auth0KeyFunc{
		jwksURL:  fmt.Sprintf("https://%s/.well-known/jwks.json", domain),
		audience: audience,
		keys:     make(map[string]*rsa.PublicKey),
	}
}

func (a *Auth0KeyFunc) refresh() error {
	resp, err := http.Get(a.jwksURL) //nolint:gosec
	if err != nil {
		return fmt.Errorf("fetch jwks: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read jwks: %w", err)
	}
	var set jwksResponse
	if err := json.Unmarshal(body, &set); err != nil {
		return fmt.Errorf("parse jwks: %w", err)
	}
	newKeys := make(map[string]*rsa.PublicKey, len(set.Keys))
	for _, k := range set.Keys {
		if k.Kty != "RSA" || k.Use != "sig" {
			continue
		}
		pub, err := parseRSAPublicKey(k.N, k.E)
		if err != nil {
			continue
		}
		newKeys[k.Kid] = pub
	}
	a.mu.Lock()
	a.keys = newKeys
	a.lastFetch = time.Now()
	a.mu.Unlock()
	return nil
}

func parseRSAPublicKey(nB64, eB64 string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nB64)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(eB64)
	if err != nil {
		return nil, err
	}
	eInt := 0
	for _, b := range eBytes {
		eInt = eInt<<8 + int(b)
	}
	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: eInt,
	}, nil
}

func (a *Auth0KeyFunc) getKey(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
	kid, ok := token.Header["kid"].(string)
	if !ok {
		return nil, fmt.Errorf("missing kid in token header")
	}

	a.mu.RLock()
	key := a.keys[kid]
	stale := time.Since(a.lastFetch) > 24*time.Hour
	a.mu.RUnlock()

	if key == nil || stale {
		if err := a.refresh(); err != nil {
			return nil, err
		}
		a.mu.RLock()
		key = a.keys[kid]
		a.mu.RUnlock()
	}
	if key == nil {
		return nil, fmt.Errorf("unknown kid: %s", kid)
	}
	return key, nil
}

// ---- Middleware ----

// UpsertUserFunc is called after successful Auth0 authentication to ensure
// the user exists in our database.
type UpsertUserFunc func(ctx context.Context, id, email string) error

func NewAuthMiddleware(jwtSecret string, auth0KF *Auth0KeyFunc, upsert UpsertUserFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, `{"type":"about:blank","title":"Unauthorized","status":401,"detail":"missing token"}`, http.StatusUnauthorized)
				return
			}
			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			var (
				token *jwt.Token
				err   error
			)

			if auth0KF != nil {
				token, err = jwt.Parse(tokenStr, auth0KF.getKey, jwt.WithValidMethods([]string{"RS256"}))
			} else {
				token, err = jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
					if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, jwt.ErrSignatureInvalid
					}
					return []byte(jwtSecret), nil
				})
			}

			if err != nil || !token.Valid {
				http.Error(w, `{"type":"about:blank","title":"Unauthorized","status":401,"detail":"invalid token"}`, http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, `{"type":"about:blank","title":"Unauthorized","status":401,"detail":"invalid claims"}`, http.StatusUnauthorized)
				return
			}

			userID, ok := claims["sub"].(string)
			if !ok || userID == "" {
				http.Error(w, `{"type":"about:blank","title":"Unauthorized","status":401,"detail":"missing sub"}`, http.StatusUnauthorized)
				return
			}

			// Validate audience if configured
			if auth0KF != nil && auth0KF.audience != "" {
				if !claimContainsAudience(claims, auth0KF.audience) {
					http.Error(w, `{"type":"about:blank","title":"Unauthorized","status":401,"detail":"invalid audience"}`, http.StatusUnauthorized)
					return
				}
			}

			// Auto-upsert Auth0 user on first login
			if auth0KF != nil && upsert != nil {
				email, _ := claims["email"].(string)
				if err := upsert(r.Context(), userID, email); err != nil {
					http.Error(w, `{"type":"about:blank","title":"Internal Server Error","status":500,"detail":"user upsert failed"}`, http.StatusInternalServerError)
					return
				}
			}

			ctx := context.WithValue(r.Context(), userIDKey{}, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func claimContainsAudience(claims jwt.MapClaims, audience string) bool {
	switch v := claims["aud"].(type) {
	case string:
		return v == audience
	case []interface{}:
		for _, a := range v {
			if s, ok := a.(string); ok && s == audience {
				return true
			}
		}
	}
	return false
}

func GetUserID(ctx context.Context) string {
	if id, ok := ctx.Value(userIDKey{}).(string); ok {
		return id
	}
	return ""
}
