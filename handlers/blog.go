package handlers

import "net/http"

type CreateBlogPayload struct {
	BlogTitle       string   `json:"blog_title"`
	BlogDescription string   `json:"blog_description"`
	BlogContent     string   `json:"blog_content"`
	BlogThumbnail   string   `json:"blog_thumbnail"`
	BlogTopicIds    []string `json:"blog_topic_ids"`
}

func (h *Handler) CreateBlogHandler(w http.ResponseWriter, r *http.Request) {

}
