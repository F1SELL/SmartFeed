package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"SmartFeed/internal/service"
	mw "SmartFeed/internal/transport/http/middleware"
)

type UserHandler struct {
	users *service.UserService
}

func NewUserHandler(users *service.UserService) *UserHandler {
	return &UserHandler{users: users}
}

// Me godoc
// @Summary Get current user profile
// @Tags users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401,404 {object} map[string]interface{}
// @Router /api/v1/users/me [get]
func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := mw.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.users.GetMe(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// Follow godoc
// @Summary Follow user
// @Tags users
// @Security BearerAuth
// @Produce json
// @Param id path int true "Followee ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400,401 {object} map[string]interface{}
// @Router /api/v1/users/follow/{id} [post]
func (h *UserHandler) Follow(w http.ResponseWriter, r *http.Request) {
	followerID, ok := mw.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idParam := chi.URLParam(r, "id")
	followeeID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	if err := h.users.Follow(r.Context(), followerID, followeeID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
