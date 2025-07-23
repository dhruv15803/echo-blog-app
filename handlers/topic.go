package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/dhruv15803/echo-blog-app/storage"
)

type CreateTopicRequestBody struct {
	TopicTitle string `json:"topic_title"`
}

func (h *Handler) CreateTopicHandler(w http.ResponseWriter, r *http.Request) {
	// admin handler
	var createTopicPayload CreateTopicRequestBody

	if err := json.NewDecoder(r.Body).Decode(&createTopicPayload); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	topicTitle := strings.ToLower(strings.TrimSpace(createTopicPayload.TopicTitle))
	// check if  topic with topictitle already exists
	existingTopics, err := h.storage.GetTopicByTopicTitle(topicTitle)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if len(existingTopics) != 0 {
		writeJSONError(w, "topic already exists", http.StatusBadRequest)
		return
	}

	topic, err := h.storage.CreateTopic(topicTitle)
	if err != nil {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	type Response struct {
		Success bool          `json:"success"`
		Topic   storage.Topic `json:"topic"`
	}

	if err := writeJSON(w, Response{Success: true, Topic: *topic}, http.StatusCreated); err != nil {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
