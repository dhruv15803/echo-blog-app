package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dhruv15803/echo-blog-app/mailer"
	"github.com/dhruv15803/echo-blog-app/storage"
	"golang.org/x/crypto/bcrypt"
)

type RegisterUserPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	var registerUserPayload RegisterUserPayload

	if err := json.NewDecoder(r.Body).Decode(&registerUserPayload); err != nil {
		log.Printf("failed to decode request body into go payload struct :- %v\n", err.Error())
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	userEmail := strings.ToLower(strings.TrimSpace(registerUserPayload.Email))
	userPassword := strings.TrimSpace(registerUserPayload.Password)

	if userEmail == "" || userPassword == "" {
		writeJSONError(w, "email and password required", http.StatusBadRequest)
		return
	}

	// check if a verified user already exists by the email

	_, err := h.storage.GetVerifiedUserByEmail(userEmail)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("failed to get verified user by email :- %v\n", err.Error())
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// if here then no verified user by the email

	// hash password using bcrypt
	hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(userPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("failed to hash plain text password using bcrypt :- %v\n", err.Error())
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// send the invitation token to the email with an activation link
	plainTextToken, err := GenerateSecureToken(32)
	if err != nil {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// the plain text token will be sent to the respective email
	// the hashed version will be stored in db
	hashedTokenByteArray := sha256.Sum256([]byte(plainTextToken))
	hashedToken := hex.EncodeToString(hashedTokenByteArray[:])
	userInvitationExpirationTime := time.Now().Add(time.Minute * 30)

	// create user and invitation //
	user, err := h.storage.CreateUserAndInvitation(userEmail, string(hashedPasswordBytes), hashedToken, userInvitationExpirationTime)
	if err != nil {
		log.Printf("failed to create and invite user :- %v\n", err.Error())
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// send mail (with retry mechanism if fails)

	maxRetryCount := 3
	currentCount := 0
	hasInvitationMailSent := false

	for currentCount < maxRetryCount {

		if err := mailer.SendGoInvitationMail(os.Getenv("GOMAIL_FROM_EMAIL"), user.Email, "user activation - echo blog", "./templates/inviteEmail.html", plainTextToken); err != nil {

			log.Printf("failed to send invitation mail , current count - %d , error :- %v\n", currentCount, err.Error())

			currentCount = currentCount + 1
			continue
		}

		hasInvitationMailSent = true
		break
	}

	if !hasInvitationMailSent {
		log.Println("failed to send invitation email")
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	type Response struct {
		Success bool         `json:"success"`
		Message string       `json:"message"`
		User    storage.User `json:"user"`
	}

	if err := writeJSON(w, Response{Success: true, Message: "user registered", User: user}, http.StatusCreated); err != nil {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
	}
}

func GenerateSecureToken(size int) (string, error) {

	bytes := make([]byte, size)

	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}
