package accounts

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"uuid"

	"github.com/stilnar92/modern-project-template/core/internal/platform/observability"
	"github.com/stilnar92/modern-project-template/core/internal/platform/problem"
)

// CookieName is the name of the session cookie.
const CookieName = "session"

type accountKey struct{}

// AccountID returns the account of the session that RequireSession found.
func AccountID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(accountKey{}).(uuid.UUID)
	return id, ok
}

// SetCookie sends the session cookie: HttpOnly, SameSite=Lax, Path=/.
func (s *Service) SetCookie(w http.ResponseWriter, sess Session) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    sess.Token,
		Path:     "/",
		Expires:  sess.ExpiresAt,
		HttpOnly: true,
		Secure:   s.opts.SecureCookie,
		SameSite: http.SameSiteLaxMode,
	})
}

// ClearCookie removes the session cookie in the browser.
func (s *Service) ClearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   s.opts.SecureCookie,
		SameSite: http.SameSiteLaxMode,
	})
}

// RequireSession lets a request pass only with a live session. It puts the
// account into the context. Without a session the answer is 401.
func (s *Service) RequireSession(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var token string
			if c, err := r.Cookie(CookieName); err == nil {
				token = c.Value
			}
			id, err := s.SessionAccount(r.Context(), token)
			if errors.Is(err, ErrNoSession) {
				problem.Write(w, http.StatusUnauthorized, "Log in again.")
				return
			}
			if err != nil {
				observability.ReportError(r.Context(), logger, "session lookup failed", err)
				problem.Write(w, http.StatusInternalServerError, "")
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), accountKey{}, id)))
		})
	}
}
