package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/dhruv15803/echo-blog-app/helpers"
	"github.com/dhruv15803/echo-blog-app/storage"
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
