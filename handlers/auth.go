package handlers

import "net/http"

type RegisterUserPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) RegisterUserHandler(w http.ResponseWriter, r *http.Request) {

}
