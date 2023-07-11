package admin

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/cicconee/weather-app/internal/app"
	"golang.org/x/crypto/bcrypt"
)

type Account struct {
	ID       int
	Approved bool
}

func (a *Account) IsApproved() bool {
	return a.Approved
}

type AdminEntity struct {
	ID           int
	Username     string
	PasswordHash string
	Approved     bool
	CreatedAt    time.Time
}

func (a *AdminEntity) ValidateUsername() error {
	if a.Username == "" {
		return &app.ServerResponseError{
			Err:        errors.New("Empty username"),
			Msg:        "Must provide a username",
			StatusCode: http.StatusUnprocessableEntity,
		}
	}

	return nil
}

func (a *AdminEntity) SetPasswordHash(password string) error {
	if password == "" {
		return &app.ServerResponseError{
			Err:        errors.New("Empty password"),
			Msg:        "Must provide a password",
			StatusCode: http.StatusUnprocessableEntity,
		}
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return err
	}

	a.PasswordHash = string(passwordHash)

	return nil
}

func (a *AdminEntity) CheckPasswordHash(p string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(a.PasswordHash), []byte(p))
	return err == nil
}

func (a *AdminEntity) IsApproved() bool {
	return a.Approved
}

func (a *AdminEntity) Account() Account {
	return Account{
		ID:       a.ID,
		Approved: a.Approved,
	}
}

func (s *AdminEntity) Scan(scanner func(...any) error) error {
	return scanner(
		&s.ID,
		&s.Username,
		&s.PasswordHash,
		&s.Approved,
		&s.CreatedAt,
	)
}

func (s *AdminEntity) Select(ctx context.Context, db *sql.DB) error {
	query := `SELECT id, username, password_hash, approved, created_at
			  FROM admins WHERE id = $1`

	return s.Scan(db.QueryRowContext(ctx, query, s.ID).Scan)
}

func (s *AdminEntity) SelectWhereUsername(ctx context.Context, db *sql.DB) error {
	query := `SELECT id, username, password_hash, approved, created_at
			  FROM admins WHERE username = $1`

	return s.Scan(db.QueryRowContext(ctx, query, s.Username).Scan)
}

func (s *AdminEntity) Insert(ctx context.Context, db *sql.DB) error {
	query := `INSERT INTO admins(username, password_hash, approved, created_at) 
			  VALUES($1, $2, $3, $4)`

	_, err := db.ExecContext(ctx, query,
		s.Username,
		s.PasswordHash,
		s.Approved,
		s.CreatedAt)

	return err
}
