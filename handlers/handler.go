package handlers

import (
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/dhruv15803/echo-blog-app/storage"
)

type Handler struct {
	storage *storage.Storage
	cld     *cloudinary.Cloudinary
}

func NewHandler(storage *storage.Storage, cld *cloudinary.Cloudinary) *Handler {
	return &Handler{
		storage: storage,
		cld:     cld,
	}
}
