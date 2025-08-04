package handlers

import (
	"database/sql"
	"errors"
	"github.com/dhruv15803/echo-blog-app/storage"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"strconv"
)

func (h *Handler) FollowUserHandler(w http.ResponseWriter, r *http.Request) {

	authUserId, ok := r.Context().Value(AuthUserId).(int)
	if !ok {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	authUser, err := h.storage.GetUserById(authUserId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, "user not found", http.StatusBadRequest)
			return
		} else {
			log.Printf("failed to get user by id :- %v\n", err.Error())
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	userId, err := strconv.Atoi(chi.URLParam(r, "userId"))
	if err != nil {
		writeJSONError(w, "invalid request param userId", http.StatusBadRequest)
		return
	}

	user, err := h.storage.GetUserById(userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, "user not found", http.StatusBadRequest)
			return
		} else {
			log.Printf("failed to get user by id :- %v\n", err.Error())
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	if authUser.Id == user.Id {
		writeJSONError(w, "cannot follow this user", http.StatusBadRequest)
		return
	}

	// check if already following
	follow, err := h.storage.GetFollow(authUser.Id, user.Id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("failed to get follow:- %v\n", err.Error())
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if follow == nil {
		// create follow
		newFollow, err := h.storage.CreateFollow(authUser.Id, user.Id)
		if err != nil {
			log.Printf("failed to create follow:- %v\n", err.Error())
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		type Response struct {
			Success bool           `json:"success"`
			Message string         `json:"message"`
			Follow  storage.Follow `json:"follow"`
		}

		if err := writeJSON(w, Response{Success: true, Message: "followed user", Follow: *newFollow}, http.StatusCreated); err != nil {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
		}
	} else {

		// remove follow
		if err := h.storage.RemoveFollow(follow.FollowerId, follow.FollowingId); err != nil {
			log.Printf("failed to remove follow:- %v\n", err.Error())
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}
		type Response struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
		}

		if err := writeJSON(w, Response{Success: true, Message: "removed follow"}, http.StatusOK); err != nil {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
		}
	}
}
