// Package accounts owns the accounts and the server sessions (README.md).
package accounts

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"
	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/stilnar92/modern-project-template/core/internal/accounts/store"
	"github.com/stilnar92/modern-project-template/core/internal/platform/db"
)

// ErrNoSession means that the token has no live session: it is empty,
// unknown, expired or ended.
var ErrNoSession = errors.New("no live session")

// Options are the session settings of the process.
type Options struct {
	// TTL is the lifetime of a new session.
	TTL time.Duration
	// SecureCookie sets the Secure flag. Turn it off only for local HTTP.
	SecureCookie bool
}

// Service is the entry point of the module.
type Service struct {
	pool *pgxpool.Pool
	opts Options
}

// New returns the service. The pool must use the app role (ADR 0009).
func New(pool *pgxpool.Pool, opts Options) *Service {
	return &Service{pool: pool, opts: opts}
}

// Session is a new session. Token goes only into the cookie.
type Session struct {
	Token     string
	ExpiresAt time.Time
}

// CreateAccount creates an empty account. The product adds its fields and
// the sign-up flow later.
func (s *Service) CreateAccount(ctx context.Context) (uuid.UUID, error) {
	id := uuid.NewV7()
	err := db.WithAccount(ctx, s.pool, id.String(), func(tx pgx.Tx) error {
		return store.New(tx).CreateAccount(ctx, id)
	})
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("create account: %w", err)
	}
	return id, nil
}

// StartSession creates a session for the account. Call it only after the
// login flow checked the user.
func (s *Service) StartSession(ctx context.Context, accountID uuid.UUID) (Session, error) {
	sess := Session{Token: rand.Text(), ExpiresAt: time.Now().Add(s.opts.TTL)}
	err := db.WithAccount(ctx, s.pool, accountID.String(), func(tx pgx.Tx) error {
		return store.New(tx).CreateSession(ctx, store.CreateSessionParams{
			TokenHash: tokenHash(sess.Token),
			AccountID: accountID,
			ExpiresAt: sess.ExpiresAt,
		})
	})
	if err != nil {
		return Session{}, fmt.Errorf("start session: %w", err)
	}
	return sess, nil
}

// SessionAccount returns the account of a live session, or ErrNoSession.
func (s *Service) SessionAccount(ctx context.Context, token string) (uuid.UUID, error) {
	if token == "" {
		return uuid.UUID{}, ErrNoSession
	}
	id, err := store.New(s.pool).SessionAccount(ctx, tokenHash(token))
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.UUID{}, ErrNoSession
	}
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("session lookup: %w", err)
	}
	return id, nil
}

// EndSession ends one session of the account (logout).
func (s *Service) EndSession(ctx context.Context, accountID uuid.UUID, token string) error {
	return db.WithAccount(ctx, s.pool, accountID.String(), func(tx pgx.Tx) error {
		return store.New(tx).DeleteSession(ctx, tokenHash(token))
	})
}

// EndAllSessions ends each session of the account, for example after a
// change of the login address.
func (s *Service) EndAllSessions(ctx context.Context, accountID uuid.UUID) error {
	return db.WithAccount(ctx, s.pool, accountID.String(), func(tx pgx.Tx) error {
		return store.New(tx).DeleteAccountSessions(ctx, accountID)
	})
}

// tokenHash is what the database keeps instead of the token.
func tokenHash(token string) []byte {
	h := sha256.Sum256([]byte(token))
	return h[:]
}
