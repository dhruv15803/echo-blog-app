package handlers

import "github.com/dhruv15803/echo-blog-app/storage"

type Handler struct {
	storage *storage.Storage
}

func NewHandler(storage *storage.Storage) *Handler {
	return &Handler{
		storage: storage,
	}
}
