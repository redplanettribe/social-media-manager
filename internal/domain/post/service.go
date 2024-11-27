package post

type Service interface {
	CreatePost(post *Post) error
	// Additional methods as needed
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreatePost(post *Post) error {
	// Business rules can be applied here
	return s.repo.Save(post)
}
