package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/dhruv15803/echo-blog-app/helpers"
	"github.com/dhruv15803/echo-blog-app/storage"
	"github.com/go-chi/chi/v5"
)

type CreateBlogPayload struct {
	BlogTitle       string `json:"blog_title"`
	BlogDescription string `json:"blog_description"`
	BlogContent     string `json:"blog_content"`
	BlogThumbnail   string `json:"blog_thumbnail"`
	BlogTopicIds    []int  `json:"blog_topic_ids"`
}

func (h *Handler) CreateBlogHandler(w http.ResponseWriter, r *http.Request) {

	userId, ok := r.Context().Value(AuthUserId).(int)
	if !ok {
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

	var createBlogPayload CreateBlogPayload

	if err := json.NewDecoder(r.Body).Decode(&createBlogPayload); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	blogTitle := strings.ToTitle(strings.TrimSpace(createBlogPayload.BlogTitle))
	blogDescription := strings.TrimSpace(createBlogPayload.BlogDescription)
	blogContent := createBlogPayload.BlogContent // this is supposed to be stringified JSON
	blogThumbnail := createBlogPayload.BlogThumbnail
	blogTopicIds := createBlogPayload.BlogTopicIds

	isBlogContentValidStringifiedJson := false

	_, err = json.Marshal(blogContent)
	if err != nil {
		writeJSONError(w, "invalid blog content", http.StatusBadRequest)
		return
	}

	isBlogContentValidStringifiedJson = true

	if blogTitle == "" || len(blogTopicIds) == 0 || !isBlogContentValidStringifiedJson {
		writeJSONError(w, "blog title, blog topics, and valid blog content is required", http.StatusBadRequest)
		return
	}

	for _, topicId := range blogTopicIds {
		_, err := h.storage.GetTopicById(topicId)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeJSONError(w, fmt.Sprintf("topic with id %v not found for blog", topicId), http.StatusBadRequest)
				return
			} else {
				writeJSONError(w, "internal server error", http.StatusInternalServerError)
				return
			}
		}
	}

	// all topics in blogTopicIds[] are valid topics
	// there should be no duplicate topics in the slice(no duplicate value)

	if helpers.HasDuplicates(blogTopicIds) {
		writeJSONError(w, "blog cannot have multiple same topics", http.StatusBadRequest)
		return
	}

	newBlog, err := h.storage.CreateBlog(blogTitle, blogDescription, blogContent, blogThumbnail, user.Id, blogTopicIds)
	if err != nil {
		log.Printf("failed to create blog :- %v\n", err.Error())
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	type Response struct {
		Success bool                   `json:"success"`
		Message string                 `json:"message"`
		Blog    storage.BlogWithTopics `json:"blog"`
	}

	if err := writeJSON(w, Response{Success: true, Message: "blog created sucessfully", Blog: *newBlog}, http.StatusCreated); err != nil {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) DeleteBlogHandler(w http.ResponseWriter, r *http.Request) {

	userId, ok := r.Context().Value(AuthUserId).(int)
	if !ok {
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

	blogId, err := strconv.Atoi(chi.URLParam(r, "blogId"))
	if err != nil {
		writeJSONError(w, "invalid request param blogId", http.StatusBadRequest)
		return
	}

	blog, err := h.storage.GetBlogById(blogId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, "blog not found", http.StatusBadRequest)
			return
		} else {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	if user.Id != blog.BlogAuthorId {
		writeJSONError(w, "user not allowed to delete blog", http.StatusUnauthorized)
		return
	}

	// delete blog
	if err = h.storage.DeleteBlogById(blog.Id); err != nil {
		log.Printf("failed to delete blog :- %v\n", err.Error())
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	type Response struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}

	if err := writeJSON(w, Response{Success: true, Message: "blog deleted successfully"}, http.StatusOK); err != nil {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) LikeBlogHandler(w http.ResponseWriter, r *http.Request) {

	userId, ok := r.Context().Value(AuthUserId).(int)
	if !ok {
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

	blogId, err := strconv.Atoi(chi.URLParam(r, "blogId"))
	if err != nil {
		writeJSONError(w, "invalid request param blogId", http.StatusBadRequest)
		return
	}

	blog, err := h.storage.GetBlogById(blogId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, "blog not found", http.StatusBadRequest)
			return
		} else {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	// if a like by the user already exists on blog
	// delete like , else create a like

	blogLike, err := h.storage.GetBlogLikeByUser(user.Id, blog.Id)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	if blogLike == nil {

		// add like
		blogLike, err := h.storage.CreateBlogLike(user.Id, blog.Id)
		if err != nil {
			log.Printf("failed to create blog like :- %v\n", err.Error())
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		type Response struct {
			Success  bool             `json:"success"`
			Message  string           `json:"message"`
			BlogLike storage.BlogLike `json:"blog_like"`
		}

		if err := writeJSON(w, Response{Success: true, Message: "liked blog successfully", BlogLike: *blogLike}, http.StatusCreated); err != nil {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
		}
	} else {
		// delete like

		if err = h.storage.RemoveLike(blogLike.LikedById, blogLike.LikedBlogId); err != nil {
			log.Printf("failed to remove blog like :- %v\n", err.Error())
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		type Response struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
		}

		if err := writeJSON(w, Response{Success: true, Message: "removed like sucessfully"}, http.StatusOK); err != nil {
			writeJSONError(w, "interna server error", http.StatusInternalServerError)
		}
	}
}
