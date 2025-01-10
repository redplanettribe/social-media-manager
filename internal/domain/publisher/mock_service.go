// Code generated by mockery v2.43.2. DO NOT EDIT.

package publisher

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// MockService is an autogenerated mock type for the Service type
type MockService struct {
	mock.Mock
}

// AddSecret provides a mock function with given fields: ctx, projectID, platformID, key, secret
func (_m *MockService) AddSecret(ctx context.Context, projectID string, platformID string, key string, secret string) error {
	ret := _m.Called(ctx, projectID, platformID, key, secret)

	if len(ret) == 0 {
		panic("no return value specified for AddSecret")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, string) error); ok {
		r0 = rf(ctx, projectID, platformID, key, secret)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// AddUserPlatformSecrets provides a mock function with given fields: ctx, projectID, platformID, key, secret
func (_m *MockService) AddUserPlatformSecrets(ctx context.Context, projectID string, platformID string, key string, secret string) error {
	ret := _m.Called(ctx, projectID, platformID, key, secret)

	if len(ret) == 0 {
		panic("no return value specified for AddUserPlatformSecrets")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, string) error); ok {
		r0 = rf(ctx, projectID, platformID, key, secret)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetAvailableSocialNetworks provides a mock function with given fields: ctx
func (_m *MockService) GetAvailableSocialNetworks(ctx context.Context) ([]Platform, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for GetAvailableSocialNetworks")
	}

	var r0 []Platform
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) ([]Platform, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) []Platform); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]Platform)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PublishPostToAssignedSocialNetworks provides a mock function with given fields: ctx, projecID, postID
func (_m *MockService) PublishPostToAssignedSocialNetworks(ctx context.Context, projecID string, postID string) error {
	ret := _m.Called(ctx, projecID, postID)

	if len(ret) == 0 {
		panic("no return value specified for PublishPostToAssignedSocialNetworks")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, projecID, postID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// PublishPostToSocialNetwork provides a mock function with given fields: ctx, projectID, postID, platformID
func (_m *MockService) PublishPostToSocialNetwork(ctx context.Context, projectID string, postID string, platformID string) error {
	ret := _m.Called(ctx, projectID, postID, platformID)

	if len(ret) == 0 {
		panic("no return value specified for PublishPostToSocialNetwork")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) error); ok {
		r0 = rf(ctx, projectID, postID, platformID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewMockService creates a new instance of MockService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockService(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockService {
	mock := &MockService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
