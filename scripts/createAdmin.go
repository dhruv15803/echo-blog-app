package scripts

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/dhruv15803/echo-blog-app/helpers"
	"github.com/dhruv15803/echo-blog-app/storage"
	"golang.org/x/crypto/bcrypt"
)

type Scripts struct {
	storage *storage.Storage
}

func NewScripts(storage *storage.Storage) *Scripts {
	return &Scripts{
		storage: storage,
	}
}

func (s *Scripts) CreateAdminUser(email string, plainTextPassword string) (*storage.User, error) {

	adminEmail := strings.ToLower(strings.TrimSpace(email))
	adminPlainTextPassword := strings.TrimSpace(plainTextPassword)

	if adminEmail == "" || adminPlainTextPassword == "" {
		return nil, errors.New("email and passoword required")
	}

	if !helpers.IsEmailValid(adminEmail) {
		return nil, errors.New("invalid email")
	}

	if !helpers.IsPasswordStrong(adminPlainTextPassword) {
		return nil, errors.New("weak password")
	}

	// check if verified user with 'adminEmail' already exists
	existingUser, err := s.storage.GetVerifiedUserByEmail(adminEmail)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("admin user not created , email taken")
	}

	adminHashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPlainTextPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	adminUser, err := s.storage.CreateAdminUser(adminEmail, string(adminHashedPassword))
	if err != nil {
		return nil, err
	}

	return adminUser, nil
}
