package helpers

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"unicode/utf8"
)

func IsPasswordStrong(password string) bool {

	isPasswordStrong := false

	const MIN_PASSWORD_LENGTH = 6
	const SPECIAL_CHARS = "!@#$%^&*()-_=+[]{}|;:',.<>?/~`"
	const UPPERCASE_CHARS = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const LOWERCASE_CHARS = "abcdefghijklmnopqrstuvwxyz"
	const NUMERICAL_CHARS = "0123456789"
	hasSpecialChar := false
	hasUppercaseChar := false
	hasLowercaseChar := false
	hasNumericalChar := false

	if utf8.RuneCountInString(password) < MIN_PASSWORD_LENGTH {
		return false
	}

	for _, passwordChar := range password {

		if hasSpecialChar && hasUppercaseChar && hasLowercaseChar && hasNumericalChar {
			isPasswordStrong = true
			break
		}

		if !hasSpecialChar && strings.Contains(SPECIAL_CHARS, string(passwordChar)) {
			hasSpecialChar = true
		}

		if !hasUppercaseChar && strings.Contains(UPPERCASE_CHARS, string(passwordChar)) {
			hasUppercaseChar = true
		}

		if !hasLowercaseChar && strings.Contains(LOWERCASE_CHARS, string(passwordChar)) {
			hasLowercaseChar = true
		}

		if !hasNumericalChar && strings.Contains(NUMERICAL_CHARS, string(passwordChar)) {
			hasNumericalChar = true
		}

	}

	if !isPasswordStrong {
		isPasswordStrong = hasSpecialChar && hasUppercaseChar && hasLowercaseChar && hasNumericalChar
	}

	return isPasswordStrong
}

func IsEmailValid(email string) bool {

	if email == "" {
		return false
	}

	if !strings.Contains(email, "@") {
		return false
	}

	emailPartsArr := strings.Split(email, "@")
	firstPart, secondPart := emailPartsArr[0], emailPartsArr[1]

	if firstPart == "" || secondPart == "" {
		return false
	}

	if strings.Contains(secondPart, ".") && len(strings.Split(secondPart, ".")) == 2 {
		return true
	} else {
		return false
	}
}

func GenerateCryptographicToken(byteSize int) (string, error) {

	bytes := make([]byte, byteSize)

	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	plainTextToken := hex.EncodeToString(bytes)
	return plainTextToken, nil
}
