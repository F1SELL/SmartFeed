package handlers

import (
	"encoding/json"
	"net/http"

	"SmartFeed/internal/service"
	mw "SmartFeed/internal/transport/http/middleware"
)

type PostHandler struct {
	posts *service.PostService
}

func NewPostHandler(posts *service.PostService) *PostHandler {
	return &PostHandler{posts: posts}
}

type createPostRequest struct {
	Content string `json:"content"`
}

// Create godoc
// @Summary Create a post
// @Tags posts
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body createPostRequest true "Post data"
// @Success 201 {object} map[string]interface{}
// @Failure 400,401 {object} map[string]interface{}
// @Router /api/v1/posts [post]
func (h *PostHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := mw.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req createPostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	post, err := h.posts.Create(r.Context(), userID, req.Content)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, post)
}
