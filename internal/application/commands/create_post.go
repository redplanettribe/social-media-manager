package commands

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/pedrodcsjostrom/opencm/internal/domain/post"
)

type CreatePostCommand struct {
	TeamID      uuid.UUID
	Title       string
	Content     string
	ScheduledAt *time.Time
}

type CreatePostHandler struct {
	postService post.Service
}

func NewCreatePostHandler(postService post.Service) *CreatePostHandler {
	return &CreatePostHandler{postService: postService}
}

func (h *CreatePostHandler) Handle(ctx context.Context, cmd *CreatePostCommand) error {
	newPost, err := post.NewPost(cmd.TeamID, cmd.Title, cmd.Content, cmd.ScheduledAt)
	if err != nil {
		return err
	}
	return h.postService.CreatePost(newPost)
}
