package handlers

import (
	"net/http"
	"strconv"

	"SmartFeed/internal/service"
	mw "SmartFeed/internal/transport/http/middleware"
)

type FeedHandler struct {
	feed *service.FeedService
}

func NewFeedHandler(feed *service.FeedService) *FeedHandler {
	return &FeedHandler{feed: feed}
}

// List godoc
// @Summary List feed posts
// @Tags feed
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} map[string]interface{}
// @Failure 401,500 {object} map[string]interface{}
// @Router /api/v1/feed [get]
func (h *FeedHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := mw.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	limit := parseInt64(r.URL.Query().Get("limit"), 20)
	offset := parseInt64(r.URL.Query().Get("offset"), 0)

	posts, err := h.feed.GetFeed(r.Context(), userID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": posts})
}

func parseInt64(raw string, def int64) int64 {
	if raw == "" {
		return def
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return def
	}
	return value
}
