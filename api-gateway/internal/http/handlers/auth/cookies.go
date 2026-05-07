package auth

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tmcrypted/berusmenu/api-gateway/internal/config"
	authv1 "github.com/tmcrypted/berusmenu/proto/gen/go/auth/v1"
)

type cookieManager struct {
	cfg config.Config
}

func newCookieManager(cfg config.Config) cookieManager {
	return cookieManager{cfg: cfg}
}

func (m cookieManager) setAuthCookies(c *gin.Context, tokens *authv1.AuthTokens) error {
	if tokens == nil {
		return nil
	}

	accessMaxAge := secondsUntil(tokens.GetAccessExpiresAt().AsTime())
	if accessMaxAge <= 0 {
		accessMaxAge = int((15 * time.Minute).Seconds())
	}

	csrfToken, err := randomCSRFToken()
	if err != nil {
		return err
	}

	m.setCookie(c, m.cfg.AccessCookieName, tokens.GetAccessToken(), accessMaxAge, "/", true)
	m.setCookie(c, m.cfg.RefreshCookieName, tokens.GetRefreshToken(), int(m.cfg.RefreshCookieTTL.Seconds()), "/v1/auth", true)
	m.setCookie(c, m.cfg.CSRFCookieName, csrfToken, int(m.cfg.RefreshCookieTTL.Seconds()), "/", false)
	return nil
}

func (m cookieManager) clearAuthCookies(c *gin.Context) {
	m.setCookie(c, m.cfg.AccessCookieName, "", -1, "/", true)
	m.setCookie(c, m.cfg.RefreshCookieName, "", -1, "/v1/auth", true)
	m.setCookie(c, m.cfg.CSRFCookieName, "", -1, "/", false)
}

func (m cookieManager) setCookie(c *gin.Context, name string, value string, maxAge int, path string, httpOnly bool) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		Domain:   m.cfg.CookieDomain,
		MaxAge:   maxAge,
		Secure:   m.cfg.CookieSecure,
		HttpOnly: httpOnly,
		SameSite: http.SameSiteLaxMode,
	})
}

func secondsUntil(t time.Time) int {
	if t.IsZero() {
		return 0
	}
	return int(time.Until(t).Seconds())
}

func randomCSRFToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
