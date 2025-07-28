package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/dhruv15803/echo-blog-app/storage"
	"github.com/go-chi/chi/v5"
)

type CreateTopicRequestBody struct {
	TopicTitle string `json:"topic_title"`
}

type UpdateTopicRequestBody struct {
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

func (h *Handler) DeleteTopicHandler(w http.ResponseWriter, r *http.Request) {
	topicId, err := strconv.Atoi(chi.URLParam(r, "topicId"))
	if err != nil {
		writeJSONError(w, "invalid request param topicId", http.StatusBadRequest)
		return
	}

	topic, err := h.storage.GetTopicById(topicId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, "topic not found", http.StatusBadRequest)
			return
		} else {
			log.Printf("failed to get topic by id :- %v\n", err.Error())
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	// delete the topic

	if err = h.storage.DeleteTopicById(topic.Id); err != nil {
		log.Printf("failed to delete topic :- %v\n", err.Error())
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	type Response struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}

	if err := writeJSON(w, Response{Success: true, Message: "deleted topic successfully"}, http.StatusOK); err != nil {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) UpdateTopicHandler(w http.ResponseWriter, r *http.Request) {
	var updateTopicPayload UpdateTopicRequestBody

	topicId, err := strconv.Atoi(chi.URLParam(r, "topicId"))
	if err != nil {
		writeJSONError(w, "invalid request param topicId", http.StatusBadRequest)
		return
	}

	topic, err := h.storage.GetTopicById(topicId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, "topic not found", http.StatusBadRequest)
			return
		} else {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	if err = json.NewDecoder(r.Body).Decode(&updateTopicPayload); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	newTopicTitle := strings.ToLower(strings.TrimSpace(updateTopicPayload.TopicTitle))
	if newTopicTitle == "" {
		writeJSONError(w, "topic title is required", http.StatusBadRequest)
		return
	}

	existingTopicsWithNewTitle, err := h.storage.GetTopicByTopicTitle(newTopicTitle)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("failed to get topic by title :- %v\n", err.Error())
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if len(existingTopicsWithNewTitle) != 0 {

		if len(existingTopicsWithNewTitle) > 1 {
			log.Println("more than 1 topic exists with new title")
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		existingTopicWithNewTitle := existingTopicsWithNewTitle[0]

		if existingTopicWithNewTitle.Id != topic.Id {
			writeJSONError(w, "topic with this title already exists", http.StatusBadRequest)
			return
		} else if existingTopicWithNewTitle.Id == topic.Id {
			// topic title is the same as new title
			type Response struct {
				Success bool   `json:"success"`
				Message string `json:"message"`
			}

			if err := writeJSON(w, Response{Success: true, Message: "topic title updated"}, http.StatusOK); err != nil {
				writeJSONError(w, "internal server error", http.StatusInternalServerError)
				return
			}

		} else {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}
	} else {

		updatedTopic, err := h.storage.UpdateTopicById(topic.Id, newTopicTitle)
		if err != nil {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		type Response struct {
			Success      bool          `json:"success"`
			Message      string        `json:"message"`
			UpdatedTopic storage.Topic `json:"updated_topic"`
		}

		if err := writeJSON(w, Response{Success: true, Message: "topic updated successfully", UpdatedTopic: *updatedTopic}, http.StatusOK); err != nil {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
		}
	}
}

func (h *Handler) GetTopicsHandler(w http.ResponseWriter, r *http.Request) {

	searchText := r.URL.Query().Get("search")
	topicTitleSearch := strings.ToLower(strings.TrimSpace(searchText))

	isSearchByTitle := false

	if topicTitleSearch != "" {
		isSearchByTitle = true
	}

	pageNum, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		writeJSONError(w, "invalid query param page", http.StatusBadRequest)
		return
	}

	limitNum, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		writeJSONError(w, "invalid query param limit", http.StatusBadRequest)
		return
	}

	skip := pageNum*limitNum - limitNum

	var topics []storage.Topic
	var noOfPages int

	if isSearchByTitle {

		topics, err = h.storage.GetTopicsBySearchTitleText(topicTitleSearch, skip, limitNum)
		if err != nil {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		totalTopics, err := h.storage.GetTopicsCountBySearchTitleText(topicTitleSearch)
		if err != nil {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		noOfPages = int(math.Ceil(float64(totalTopics) / float64(limitNum)))

	} else {

		topics, err = h.storage.GetTopics(skip, limitNum)
		if err != nil {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		totalTopicsCount, err := h.storage.GetTopicsCount()
		if err != nil {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		noOfPages = int(math.Ceil(float64(totalTopicsCount) / float64(limitNum)))
	}

	type Response struct {
		Success   bool            `json:"success"`
		Topics    []storage.Topic `json:"topics"`
		NoOfPages int             `json:"no_of_pages"`
	}

	if err := writeJSON(w, Response{Success: true, Topics: topics, NoOfPages: int(noOfPages)}, http.StatusOK); err != nil {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
