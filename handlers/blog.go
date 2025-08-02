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

type CreateBlogCommentPayload struct {
	CommentContent  string `json:"comment_content"`
	ParentCommentId *int   `json:"parent_comment_id"`
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

func (h *Handler) CreateBlogCommentHandler(w http.ResponseWriter, r *http.Request) {

	var createBlogCommentPayload CreateBlogCommentPayload

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

	if err := json.NewDecoder(r.Body).Decode(&createBlogCommentPayload); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	blogCommentContent := strings.TrimSpace(createBlogCommentPayload.CommentContent)
	isChildComment := createBlogCommentPayload.ParentCommentId != nil

	if blogCommentContent == "" {
		writeJSONError(w, "comment content is required", http.StatusBadRequest)
		return
	}

	var blogComment *storage.BlogComment

	if isChildComment {

		parentCommentId := *createBlogCommentPayload.ParentCommentId
		// if we are creating  a child comment for this blog
		// then the parent comment should exist and also should be blog's comment

		parentComment, err := h.storage.GetBlogCommentById(parentCommentId)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeJSONError(w, "parent comment not found", http.StatusBadRequest)
				return
			} else {
				writeJSONError(w, "internal server error", http.StatusInternalServerError)
				return
			}
		}

		if parentComment.BlogId != blog.Id {
			writeJSONError(w, "parent comment is not a comment of this blog", http.StatusBadRequest)
			return
		}

		blogComment, err = h.storage.CreateChildBlogComment(blogCommentContent, blog.Id, user.Id, parentComment.Id)
		if err != nil {
			log.Printf("failed to create child blog comment :- %v\n", err.Error())
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		log.Println(blogComment)
	} else {

		blogComment, err = h.storage.CreateBlogComment(blogCommentContent, blog.Id, user.Id)
		if err != nil {
			log.Printf("failed to create blog comment :- %v\n", err.Error())
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	type Response struct {
		Success     bool                `json:"success"`
		Message     string              `json:"message"`
		BlogComment storage.BlogComment `json:"blog_comment"`
	}

	if err := writeJSON(w, Response{Success: true, Message: "created blog comment", BlogComment: *blogComment}, http.StatusOK); err != nil {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) LikeBlogCommentHandler(w http.ResponseWriter, r *http.Request) {
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

	blogCommentId, err := strconv.Atoi(chi.URLParam(r, "blogCommentId"))
	if err != nil {
		writeJSONError(w, "invalid request param blogCommentId", http.StatusBadRequest)
		return
	}

	blogComment, err := h.storage.GetBlogCommentById(blogCommentId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(w, "blog comment not found", http.StatusBadRequest)
			return
		} else {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	// check if this user has liked this blog comment
	blogCommentLike, err := h.storage.GetBlogCommentLike(user.Id, blogComment.Id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("failed to get blog comment like :- %v\n", err.Error())
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	var responseMsg string
	if blogCommentLike == nil {

		// create comment like
		_, err := h.storage.CreateBlogCommentLike(user.Id, blogComment.Id)
		if err != nil {
			log.Printf("failed to create blog comment like :- %v\n", err.Error())
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		responseMsg = "create blog comment like"

	} else {
		// remove blog comment like
		if err := h.storage.RemoveBlogCommentLike(blogCommentLike.LikedById, blogCommentLike.LikedBlogCommentId); err != nil {
			log.Printf("failed to remove blog comment like :- %v\n", err.Error())
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		responseMsg = "remove blog comment like"
	}

	type Response struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}

	if err := writeJSON(w, Response{Success: true, Message: responseMsg}, http.StatusOK); err != nil {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) BookmarkBlogHandler(w http.ResponseWriter, r *http.Request) {

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

	// check if bookmark by user of this blog already exists
	blogBookmark, err := h.storage.GetBlogBookmark(user.Id, blog.Id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	var responseMsg string

	if blogBookmark == nil {

		// create a bookmark
		_, err := h.storage.CreateBlogBookmark(user.Id, blog.Id)
		if err != nil {
			log.Printf("failed to create blog bookmark :- %v\n", err.Error())
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		responseMsg = "created blog bookmark"

	} else {
		// remove a bookmark
		if err := h.storage.RemoveBlogBookmark(blogBookmark.BookmarkedById, blogBookmark.BookmarkedBlogId); err != nil {
			log.Printf("failed to remove blog bookmark :- %v\n", err.Error())
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		responseMsg = "remove blog bookmark"
	}

	type Response struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}

	if err := writeJSON(w, Response{Success: true, Message: responseMsg}, http.StatusOK); err != nil {
		writeJSONError(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
