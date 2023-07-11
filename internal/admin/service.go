package admin

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/cicconee/weather-app/internal/app"
	"github.com/golang-jwt/jwt/v4"
)

type Service struct {
	Secret []byte
	DB     *sql.DB
}

func New(secret []byte, db *sql.DB) *Service {
	return &Service{
		Secret: secret,
		DB:     db,
	}
}

// Signup will create a admin and store it into the database. A admin will only
// signup successfully if the username is not in use.
func (s *Service) Signup(ctx context.Context, username string, password string) error {
	admin := AdminEntity{Username: username}

	// Check if username is in use.
	err := admin.SelectWhereUsername(ctx, s.DB)
	if err == nil {
		return &app.ServerResponseError{
			Err:        fmt.Errorf("username %q in use", admin.Username),
			Msg:        "Username is taken",
			StatusCode: http.StatusConflict,
		}
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("selecting admin (username=%s): %w", admin.Username, err)
	}

	// Hash and set PasswordHash. SetPasswordHash will also validate the password.
	if err := admin.SetPasswordHash(password); err != nil {
		return fmt.Errorf("Setting password hash: %w", err)
	}

	// Validate the username.
	if err := admin.ValidateUsername(); err != nil {
		return fmt.Errorf("Validating username: %w", err)
	}

	admin.Approved = false
	admin.CreatedAt = time.Now().UTC()

	// Insert admin.
	if err := admin.Insert(ctx, s.DB); err != nil {
		return fmt.Errorf("inserting admin (username=%s): %w", admin.Username, err)
	}

	return nil
}

// Login will get an Admin associated with the username. It then hashes
// the provided password and compares it to the password stored in the
// database. If the credentials are valid, and Admin has been approved,
// it will return an access token.
//
// All access tokens created for an admin will be for valid credentials
// and the admin has an approval status set to true. If the credentials
// are incorrect or the admin has not been approved, no access token will
// be returned.
//
// This is the only way to get a admin access token.
func (s *Service) Login(ctx context.Context, username string, password string) (string, error) {
	admin := AdminEntity{Username: username}
	if err := admin.SelectWhereUsername(ctx, s.DB); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", &app.ServerResponseError{
				Err:        fmt.Errorf("admin not found"),
				Msg:        "Invalid credentials",
				StatusCode: http.StatusUnauthorized,
			}
		}
		return "", fmt.Errorf("selecting admin (username=%s): %w", admin.Username, err)
	}

	if !admin.CheckPasswordHash(password) {
		return "", &app.ServerResponseError{
			Err:        fmt.Errorf("invalid password"),
			Msg:        "Invalid credentials",
			StatusCode: http.StatusUnauthorized,
		}
	}

	if !admin.IsApproved() {
		return "", &app.ServerResponseError{
			Err:        errors.New("admin not approved"),
			Msg:        "The administrative department has not granted you administrative rights",
			StatusCode: http.StatusUnauthorized,
		}
	}

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = fmt.Sprintf("%d", admin.ID)
	claims["exp"] = time.Now().Add(time.Hour).Unix()

	tokenStr, err := token.SignedString(s.Secret)
	if err != nil {
		return "", fmt.Errorf("signing token: %w", err)
	}

	return tokenStr, nil
}

// Validate will parse and validate a token. If the token belongs to an
// admin, the admin account will be returned. Any errors that occur while
// parsing or validating the token will be returned.
//
// A parsed and validated token will return an error if the token "sub"
// claim does not contain an admins ID that is stored in the database.
//
// Validate will not check if the admins account has been approved, this
// will be the callers responsibility.
func (s *Service) Validate(ctx context.Context, tokenStr string) (Account, error) {
	token, err := jwt.Parse(
		tokenStr,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("Invalid signing method")
			}
			return s.Secret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))
	if err != nil {
		return Account{}, &app.ServerResponseError{
			Err:        fmt.Errorf("parsing token: %w", err),
			Msg:        "Please login",
			StatusCode: http.StatusUnauthorized,
		}
	}

	// Validate the claims.
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		// This should never return since only tokens that will parse successfully
		// are created in the Login method. The Login method will always set the
		// token claims as type jwt.MapClaims.
		return Account{}, errors.New("Could not get token claims")
	}

	if !claims.VerifyExpiresAt(time.Now().Unix(), true) {
		return Account{}, &app.ServerResponseError{
			Err:        errors.New("Token is expired"),
			Msg:        "Please login",
			StatusCode: http.StatusUnauthorized,
		}
	}

	subClaim, ok := claims["sub"]
	if !ok {
		// This should never return since only tokens that will parse successfully
		// are created in the Login method. The Login method will always set the sub
		// claim value.
		return Account{}, errors.New("missing sub claim")
	}

	subStr, ok := subClaim.(string)
	if !ok {
		// This should never return since only tokens that will parse successfully
		// are created in the Login method. The Login method will always set the sub
		// claim as type string.
		return Account{}, errors.New("sub claim not type string")
	}

	sub, err := strconv.Atoi(subStr)
	if err != nil {
		// This should never return since only tokens that will parse successfully
		// are created in the Login method. The Login method will always set the sub
		// claim as a string that is created from a int.
		return Account{}, fmt.Errorf("parsing sub to int: %w", err)
	}

	// Make sure the admin still exists. This is neccessary because an admin with a
	// valid token may have been deleted between the time of receiving a token and
	// validating it i.e. some one deemed them not worthy of being an admin anymore.
	admin := AdminEntity{ID: sub}
	if err := admin.Select(ctx, s.DB); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Account{}, &app.ServerResponseError{
				Err:        fmt.Errorf("admin not found (id=%d)", admin.ID),
				Msg:        "Account not found",
				StatusCode: http.StatusUnauthorized,
			}
		}

		return Account{}, fmt.Errorf("selecting admin: %w", err)
	}

	return admin.Account(), nil
}
