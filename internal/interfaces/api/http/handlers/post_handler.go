package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/pedrodcsjostrom/opencm/internal/application/commands"
	"github.com/pedrodcsjostrom/opencm/internal/application/interfaces"
)

type PostHandler struct {
	commandBus interfaces.CommandBus
}

func NewPostHandler(commandBus interfaces.CommandBus) *PostHandler {
	return &PostHandler{commandBus: commandBus}
}

type createPostRequest struct {
	TeamID      string  `json:"team_id"`
	Title       string  `json:"title"`
	Content     string  `json:"content"`
	ScheduledAt *string `json:"scheduled_at,omitempty"`
}

func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	var req createPostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	teamID, err := uuid.Parse(req.TeamID)
	if err != nil {
		http.Error(w, "Invalid team ID", http.StatusBadRequest)
		return
	}

	var scheduledAt *time.Time
	if req.ScheduledAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ScheduledAt)
		if err != nil {
			http.Error(w, "Invalid scheduled_at format", http.StatusBadRequest)
			return
		}
		scheduledAt = &t
	}

	cmd := &commands.CreatePostCommand{
		TeamID:      teamID,
		Title:       req.Title,
		Content:     req.Content,
		ScheduledAt: scheduledAt,
	}

	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
