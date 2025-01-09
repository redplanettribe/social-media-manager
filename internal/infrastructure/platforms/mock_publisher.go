// Code generated by mockery v2.43.2. DO NOT EDIT.

package platforms

import (
	context "context"

	post "github.com/pedrodcsjostrom/opencm/internal/domain/post"
	mock "github.com/stretchr/testify/mock"
)

// MockPublisher is an autogenerated mock type for the Publisher type
type MockPublisher struct {
	mock.Mock
}

// AddSecret provides a mock function with given fields: key, secret
func (_m *MockPublisher) AddSecret(key string, secret string) (string, error) {
	ret := _m.Called(key, secret)

	if len(ret) == 0 {
		panic("no return value specified for AddSecret")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string) (string, error)); ok {
		return rf(key, secret)
	}
	if rf, ok := ret.Get(0).(func(string, string) string); ok {
		r0 = rf(key, secret)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(key, secret)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Publish provides a mock function with given fields: ctx, _a1
func (_m *MockPublisher) Publish(ctx context.Context, _a1 *post.QPost) error {
	ret := _m.Called(ctx, _a1)

	if len(ret) == 0 {
		panic("no return value specified for Publish")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *post.QPost) error); ok {
		r0 = rf(ctx, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ValidateSecrets provides a mock function with given fields: Secrets
func (_m *MockPublisher) ValidateSecrets(Secrets string) error {
	ret := _m.Called(Secrets)

	if len(ret) == 0 {
		panic("no return value specified for ValidateSecrets")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(Secrets)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewMockPublisher creates a new instance of MockPublisher. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockPublisher(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockPublisher {
	mock := &MockPublisher{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
