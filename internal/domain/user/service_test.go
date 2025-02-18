package user

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/redplanettribe/social-media-manager/internal/infrastructure/session"
)

func TestGetUser(t *testing.T) {
	mockRepo := new(MockRepository)
	passwordHasher := NewMockPasswordHasher(t)
	sessionManager := session.NewMockManager(t)
	service := NewService(mockRepo, sessionManager, passwordHasher)

	ctx := context.Background()
	id := uuid.New().String()
	userResponse := &UserResponse{ID: id, Username: "testuser", Email: "test@example.com"}

	mockRepo.On("FindByID", ctx, id).Return(userResponse, nil)

	result, err := service.GetUser(ctx)
	assert.NoError(t, err)
	assert.Equal(t, userResponse, result)
	mockRepo.AssertExpectations(t)
}

func TestGetUser_InvalidUUID(t *testing.T) {
	mockRepo := new(MockRepository)
	passwordHasher := NewMockPasswordHasher(t)
	sessionManager := session.NewMockManager(t)
	service := NewService(mockRepo, sessionManager, passwordHasher)

	ctx := context.Background()

	result, err := service.GetUser(ctx)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestCreateUser(t *testing.T) {
	tests := []struct {
		name             string
		mockFindUser     func(mockRepo *MockRepository, ctx context.Context, username, email string)
		mockHashPassword func(passwordHasher *MockPasswordHasher, password string)
		mockSaveUser     func(mockRepo *MockRepository, ctx context.Context)
		username         string
		password         string
		email            string
		expectedErr      error
	}{
		{
			name: "UserAlreadyExists",
			mockFindUser: func(mockRepo *MockRepository, ctx context.Context, username, email string) {
				mockRepo.On("FindByUsernameOrEmail", ctx, username, email).Return(&UserResponse{}, nil)
			},
			mockHashPassword: nil,
			mockSaveUser:     nil,
			username:         "testuser",
			password:         "password",
			email:            "test@example.com",
			expectedErr:      ErrExistingUser,
		},
		{
			name: "HashPasswordError",
			mockFindUser: func(mockRepo *MockRepository, ctx context.Context, username, email string) {
				mockRepo.On("FindByUsernameOrEmail", ctx, username, email).Return(nil, nil)
			},
			mockHashPassword: func(passwordHasher *MockPasswordHasher, password string) {
				passwordHasher.On("Hash", password).Return("", "", assert.AnError)
			},
			mockSaveUser: nil,
			username:     "testuser",
			password:     "password",
			email:        "test@example.com",
			expectedErr:  assert.AnError,
		},
		{
			name: "SaveError",
			mockFindUser: func(mockRepo *MockRepository, ctx context.Context, username, email string) {
				mockRepo.On("FindByUsernameOrEmail", ctx, username, email).Return(nil, nil)
			},
			mockHashPassword: func(passwordHasher *MockPasswordHasher, password string) {
				passwordHasher.On("Hash", password).Return("hashedPassword", "salt", nil)
			},
			mockSaveUser: func(mockRepo *MockRepository, ctx context.Context) {
				mockRepo.On("Save", ctx, mock.Anything).Return(assert.AnError)
			},
			username:    "testuser",
			password:    "password",
			email:       "test@example.com",
			expectedErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := new(MockRepository)
			passwordHasher := NewMockPasswordHasher(t)
			sessionManager := session.NewMockManager(t)
			service := NewService(mockRepo, sessionManager, passwordHasher)
			ctx := context.Background()

			// Apply the mocks
			if tt.mockFindUser != nil {
				tt.mockFindUser(mockRepo, ctx, tt.username, tt.email)
			}
			if tt.mockHashPassword != nil {
				tt.mockHashPassword(passwordHasher, tt.password)
			}
			if tt.mockSaveUser != nil {
				tt.mockSaveUser(mockRepo, ctx)
			}

			// Call the function under test
			_, err := service.CreateUser(ctx, tt.username, tt.password, tt.email)

			// Assertions
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify expectations
			mockRepo.AssertExpectations(t)
			passwordHasher.AssertExpectations(t)
		})
	}
}

func TestLogin(t *testing.T) {
	tests := []struct {
		name              string
		mockFindUser      func(mockRepo *MockRepository, ctx context.Context, email string)
		mockValidatePass  func(passwordHasher *MockPasswordHasher, password, hashedPassword, salt string)
		mockCreateSession func(sessionManager *session.MockManager, ctx context.Context, userID string)
		email             string
		password          string
		expectedSession   *session.Session
		expectedErr       error
	}{
		{
			name: "UserNotFound",
			mockFindUser: func(mockRepo *MockRepository, ctx context.Context, email string) {
				mockRepo.On("FindByEmail", ctx, email).Return(nil, nil)
			},
			mockValidatePass:  nil,
			mockCreateSession: nil,
			email:             "test@example.com",
			password:          "password",
			expectedSession:   &session.Session{},
			expectedErr:       ErrUserNotFound,
		},
		{
			name: "InvalidPassword",
			mockFindUser: func(mockRepo *MockRepository, ctx context.Context, email string) {
				mockRepo.On("FindByEmail", ctx, email).Return(&FullUserResponse{HashedPasword: "hashedPassword", Salt: "salt"}, nil)
			},
			mockValidatePass: func(passwordHasher *MockPasswordHasher, password, hashedPassword, salt string) {
				passwordHasher.On("Validate", password, hashedPassword, salt).Return(false)
			},
			mockCreateSession: nil,
			email:             "test@example.com",
			password:          "wrongpassword",
			expectedSession:   &session.Session{},
			expectedErr:       ErrInvalidPassword,
		},
		{
			name: "CreateSessionError",
			mockFindUser: func(mockRepo *MockRepository, ctx context.Context, email string) {
				mockRepo.On("FindByEmail", ctx, email).Return(&FullUserResponse{ID: "userID", HashedPasword: "hashedPassword", Salt: "salt"}, nil)
			},
			mockValidatePass: func(passwordHasher *MockPasswordHasher, password, hashedPassword, salt string) {
				passwordHasher.On("Validate", password, hashedPassword, salt).Return(true)
			},
			mockCreateSession: func(sessionManager *session.MockManager, ctx context.Context, userID string) {
				sessionManager.On("CreateSession", ctx, userID).Return(&session.Session{}, assert.AnError)
			},
			email:           "test@example.com",
			password:        "password",
			expectedSession: &session.Session{},
			expectedErr:     assert.AnError,
		},
		{
			name: "SuccessfulLogin",
			mockFindUser: func(mockRepo *MockRepository, ctx context.Context, email string) {
				mockRepo.On("FindByEmail", ctx, email).Return(&FullUserResponse{ID: "userID", HashedPasword: "hashedPassword", Salt: "salt"}, nil)
			},
			mockValidatePass: func(passwordHasher *MockPasswordHasher, password, hashedPassword, salt string) {
				passwordHasher.On("Validate", password, hashedPassword, salt).Return(true)
			},
			mockCreateSession: func(sessionManager *session.MockManager, ctx context.Context, userID string) {
				sessionManager.On("CreateSession", ctx, userID).Return(&session.Session{ID: "sessionID"}, nil)
			},
			email:           "test@example.com",
			password:        "password",
			expectedSession: &session.Session{ID: "sessionID"},
			expectedErr:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := new(MockRepository)
			passwordHasher := NewMockPasswordHasher(t)
			sessionManager := session.NewMockManager(t)
			service := NewService(mockRepo, sessionManager, passwordHasher)
			ctx := context.Background()

			// Apply the mocks
			if tt.mockFindUser != nil {
				tt.mockFindUser(mockRepo, ctx, tt.email)
			}
			if tt.mockValidatePass != nil {
				tt.mockValidatePass(passwordHasher, tt.password, "hashedPassword", "salt")
			}
			if tt.mockCreateSession != nil {
				tt.mockCreateSession(sessionManager, ctx, "userID")
			}

			// Call the function under test
			result, err := service.Login(ctx, tt.email, tt.password)

			// Assertions
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedSession, result)
			}

			// Verify expectations
			mockRepo.AssertExpectations(t)
			passwordHasher.AssertExpectations(t)
			sessionManager.AssertExpectations(t)
		})
	}
}
