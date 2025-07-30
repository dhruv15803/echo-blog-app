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

	"github.com/dhruv15803/echo-blog-app/helpers"
	"github.com/dhruv15803/echo-blog-app/mailer"
	"github.com/dhruv15803/echo-blog-app/storage"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type RegisterUserPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginUserPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ForgotPasswordPayload struct {
	Email string `json:"email"`
}

type ResetPasswordPayload struct {
	Password string `json:"password"`
}

var (
	JWT_SECRET = []byte(os.Getenv("JWT_SECRET"))
)

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

	if !helpers.IsEmailValid(userEmail) {
		writeJSONError(w, "incorrect email format", http.StatusBadRequest)
		return
	}

	if !helpers.IsPasswordStrong(userPassword) {
		writeJSONError(w, "weak password", http.StatusBadRequest)
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
	plainTextToken, err := helpers.GenerateCryptographicToken(32)
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

func (h *Handler) LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	var loginUserPayload LoginUserPayload

	if err := json.NewDecoder(r.Body).Decode(&loginUserPayload); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	userEmail := strings.ToLower(strings.TrimSpace(loginUserPayload.Email))
	userPassword := strings.TrimSpace(loginUserPayload.Password)

	if userEmail == "" || userPassword == "" {
		writeJSONError(w, "email and password required", http.StatusBadRequest)
		return
	}

	user, err := h.storage.GetVerifiedUserByEmail(userEmail)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, "invalid email or password", http.StatusBadRequest)
			return
		} else {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	plainTextUserPassword := userPassword
	hashedPassword := user.Password

	if err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainTextUserPassword)); err != nil {
		writeJSONError(w, "invalid email or password", http.StatusBadRequest)
		return
	}

	claims := jwt.MapClaims{
		"sub": user.Id,
		"exp": time.Now().Add(time.Hour * 24 * 2).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(JWT_SECRET)
	if err != nil {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	var sameSiteConfig http.SameSite

	if os.Getenv("GO_ENV") == "development" {
		sameSiteConfig = http.SameSiteLaxMode
	} else {
		sameSiteConfig = http.SameSiteNoneMode
	}

	cookie := http.Cookie{
		Name:     "auth_token",
		Value:    tokenStr,
		HttpOnly: true,
		Secure:   os.Getenv("GO_ENV") == "production",
		Path:     "/",
		SameSite: sameSiteConfig,
	}

	http.SetCookie(w, &cookie)

	type Response struct {
		Success bool         `json:"success"`
		Message string       `json:"message"`
		User    storage.User `json:"user"`
	}

	if err := writeJSON(w, Response{Success: true, Message: "user logged in", User: *user}, http.StatusOK); err != nil {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) ActivateUserHandler(w http.ResponseWriter, r *http.Request) {

	// token from request parameter
	plainTextToken := chi.URLParam(r, "token")

	// hash plain text token with sha256

	hashedTokenByteArr := sha256.Sum256([]byte(plainTextToken))
	hashedToken := hex.EncodeToString(hashedTokenByteArr[:])

	activatedUser, err := h.storage.ActivateUserHandler(hashedToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, "no valid user invite found", http.StatusBadRequest)
			return
		} else {
			log.Printf("failed to activate user by the provided plain text token :- %v\n", err.Error())
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	// persist a jwt auth token in a cookie on the client browser on this endpoint
	claims := jwt.MapClaims{
		"sub": activatedUser.Id,
		"exp": time.Now().Add(time.Hour * 24 * 2).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(JWT_SECRET)
	if err != nil {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	var sameSiteConfig http.SameSite

	if os.Getenv("GO_ENV") == "development" {
		sameSiteConfig = http.SameSiteLaxMode
	} else {
		sameSiteConfig = http.SameSiteNoneMode
	}

	cookie := http.Cookie{
		Name:     "auth_token",
		Value:    tokenStr,
		HttpOnly: true,
		Secure:   os.Getenv("GO_ENV") == "production",
		Path:     "/",
		SameSite: sameSiteConfig,
	}

	http.SetCookie(w, &cookie)

	type Response struct {
		Success bool         `json:"success"`
		Message string       `json:"message"`
		User    storage.User `json:"user"`
	}

	if err := writeJSON(w, Response{Success: true, Message: "activated and logged in", User: *activatedUser}, http.StatusOK); err != nil {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetAuthUser(w http.ResponseWriter, r *http.Request) {

	userId, ok := r.Context().Value(AuthUserId).(int)
	if !ok {
		log.Println("failed to assert auth user id from context to type int")
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	user, err := h.storage.GetUserById(userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, "user not found", http.StatusBadRequest)
			return
		} else {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	type Response struct {
		Success bool         `json:"success"`
		User    storage.User `json:"user"`
	}

	if err := writeJSON(w, Response{Success: true, User: *user}, http.StatusOK); err != nil {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) {

	var forgotPasswordPayload ForgotPasswordPayload

	if err := json.NewDecoder(r.Body).Decode(&forgotPasswordPayload); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	userEmail := strings.ToLower(strings.TrimSpace(forgotPasswordPayload.Email))

	if userEmail == "" {
		writeJSONError(w, "email is required", http.StatusBadRequest)
		return
	}

	user, err := h.storage.GetVerifiedUserByEmail(userEmail)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, "user not found", http.StatusBadRequest)
			return
		} else {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	plainTextToken, err := helpers.GenerateCryptographicToken(32)
	if err != nil {
		log.Printf("failed to generate token :- %v\n", err.Error())
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	hashedTokenByteArr := sha256.Sum256([]byte(plainTextToken))
	hashedTokenStr := hex.EncodeToString(hashedTokenByteArr[:])
	expirationTime := time.Now().Add(time.Minute * 15) // 15 minutes

	_, err = h.storage.CreatePasswordReset(hashedTokenStr, user.Id, expirationTime)
	if err != nil {
		log.Printf("failed to create password reset :- %v\n", err.Error())
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	maxRetries := 3
	currentCount := 0
	isPasswordResetMailSent := false

	for currentCount < maxRetries {

		if err := mailer.SendGoPasswordResetMail(os.Getenv("GOMAIL_FROM_EMAIL"), user.Email, "Echo BLog Password Reset", "./templates/forgotPassword.html", plainTextToken); err != nil {
			log.Printf("failed to send password reset mail , retry count - %v", currentCount+1)
			currentCount++
		}

		// mail sent
		isPasswordResetMailSent = true
		if isPasswordResetMailSent {
			break
		}
	}

	if !isPasswordResetMailSent {
		log.Printf("failed to sent password reset mail after %v retries", maxRetries)
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	type Response struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}

	if err := writeJSON(w, Response{Success: true, Message: "password reset email sent successfully"}, http.StatusOK); err != nil {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {

	var resetPasswordPayload ResetPasswordPayload
	plainTextToken := chi.URLParam(r, "token")

	if err := json.NewDecoder(r.Body).Decode(&resetPasswordPayload); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	newPassword := strings.TrimSpace(resetPasswordPayload.Password)

	// hash plain token with sha256 and reset the corresponding user's password with this hashed token

	plainTextTokenBytes := []byte(plainTextToken)
	hashedTokenByteArr := sha256.Sum256(plainTextTokenBytes)
	hashedTokenStr := hex.EncodeToString(hashedTokenByteArr[:])

	// reset password
	hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	_, err = h.storage.ResetPassword(string(hashedNewPassword), hashedTokenStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, "reset not available or already used", http.StatusBadRequest)
			return
		} else {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	type Response struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}

	if err := writeJSON(w, Response{Success: true, Message: "password reset successfull"}, http.StatusOK); err != nil {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
