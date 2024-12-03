package session

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestManager_ValidateSession(t *testing.T) {
	mockRepo := new(MockRepository)
	manager := NewManager(mockRepo)

	ctx := context.Background()
	sessionID := "test-session-id"
	validSession := &Session{
		ExpiresAt: time.Now().Add(time.Hour * 24),
	}
	invalidSession := &Session{}

	tests := []struct {
		name            string
		sessionID       string
		setUpMock       func(ctx context.Context, mockRepo *MockRepository)
		expectedError   error
		expectedSession *Session
	}{
		{
			name:      "valid session",
			sessionID: sessionID,
			setUpMock: func(ctx context.Context, mockRepo *MockRepository) {
				mockRepo.On("GetSessionByID", ctx, sessionID).Return(validSession, nil).Once()
			},
			expectedError:   nil,
			expectedSession: validSession,
		},
		{
			name:      "invalid session",
			sessionID: sessionID,
			setUpMock: func(ctx context.Context, mockRepo *MockRepository) {
				mockRepo.On("GetSessionByID", ctx, sessionID).Return(invalidSession, nil).Once()
			},
			expectedError:   ErrInvalidSession,
			expectedSession: &Session{},
		},
		{
			name:      "session not found",
			sessionID: sessionID,
			setUpMock: func(ctx context.Context, mockRepo *MockRepository) {
				mockRepo.On("GetSessionByID", ctx, sessionID).Return(nil, assert.AnError).Once()
			},
			expectedError:   assert.AnError,
			expectedSession: &Session{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setUpMock(ctx, mockRepo)
			session, err := manager.ValidateSession(ctx, tt.sessionID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedSession, session)
		})
	}

	mockRepo.AssertExpectations(t)
}
func TestManager_CreateSession(t *testing.T) {
	mockRepo := NewMockRepository(t)
	manager := NewManager(mockRepo)

	ctx := context.Background()
	userID := "test-user-id"

	tests := []struct {
		name            string
		userID          string
		setUpMock       func(ctx context.Context, mockRepo *MockRepository)
		expectedError   error
		expectedSession *Session
	}{
		{
			name:   "successful session creation",
			userID: userID,
			setUpMock: func(ctx context.Context, mockRepo *MockRepository) {
				mockRepo.On("DeleteSessionsForUser", ctx, userID).Return(nil).Once()
				mockRepo.On("CreateSession", ctx, mock.Anything).Return("new-session-id", nil).Once()
			},
			expectedError: nil,
		},
		{
			name:   "error deleting sessions",
			userID: userID,
			setUpMock: func(ctx context.Context, mockRepo *MockRepository) {
				mockRepo.On("DeleteSessionsForUser", ctx, userID).Return(assert.AnError).Once()
			},
			expectedError:   assert.AnError,
			expectedSession: &Session{},
		},
		{
			name:   "error creating session",
			userID: userID,
			setUpMock: func(ctx context.Context, mockRepo *MockRepository) {
				mockRepo.On("DeleteSessionsForUser", ctx, userID).Return(nil).Once()
				mockRepo.On("CreateSession", ctx, mock.Anything).Return("", assert.AnError).Once()
			},
			expectedError:   assert.AnError,
			expectedSession: &Session{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setUpMock(ctx, mockRepo)
			session, err := manager.CreateSession(ctx, tt.userID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}
			if tt.expectedSession != nil {
				assert.Equal(t, tt.expectedSession, session)
			}
		})
	}

	mockRepo.AssertExpectations(t)
}
