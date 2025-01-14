package linkedin

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/pedrodcsjostrom/opencm/internal/domain/media"
	"github.com/pedrodcsjostrom/opencm/internal/domain/post"
)

type Poster interface {
	Post(ctx context.Context, post *post.PublishPost, media []*media.Media) error
}

type PosterFactory interface {
	NewPoster(p *post.PublishPost, userSecrets UserSecrets, platformSecrets PlatformSecrets) (Poster, error)
}

type posterFactory struct {
}

func NewPosterFactory() PosterFactory {
	return &posterFactory{}
}

func (pf *posterFactory) NewPoster(p *post.PublishPost, userSecrets UserSecrets, platformSecrets PlatformSecrets) (Poster, error) {
	if p.Type == post.PostTypeUndefined {
		return nil, errors.New("post type is undefined")
	}
	switch p.Type {
	case post.PostTypeText:
		return NewTextPoster(userSecrets, platformSecrets), nil
	default:
		return nil, errors.New("invalid post type")
	}
}

// LinkedInPost represents the JSON structure required by LinkedIn's API for creating a post.
type LinkedInPost struct {
    Author                    string       `json:"author"`
    Commentary                string       `json:"commentary"`
    Visibility                string       `json:"visibility"`
    Distribution              Distribution `json:"distribution"`
    LifecycleState            string       `json:"lifecycleState"`
    IsReshareDisabledByAuthor bool         `json:"isReshareDisabledByAuthor"`
}

// Distribution represents the distribution settings in the LinkedIn post.
type Distribution struct {
    FeedDistribution               string   `json:"feedDistribution"`
    TargetEntities                 []string `json:"targetEntities"`
    ThirdPartyDistributionChannels []string `json:"thirdPartyDistributionChannels"`
}

func setHeaders(req *http.Request, accessToken string) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-Restli-Protocol-Version", "2.0.0")
	req.Header.Set("LinkedIn-Version", "202411")
}