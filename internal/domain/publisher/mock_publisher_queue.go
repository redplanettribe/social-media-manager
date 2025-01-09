// Code generated by mockery v2.43.2. DO NOT EDIT.

package publisher

import (
	context "context"

	post "github.com/pedrodcsjostrom/opencm/internal/domain/post"
	mock "github.com/stretchr/testify/mock"
)

// MockPublisherQueue is an autogenerated mock type for the PublisherQueue type
type MockPublisherQueue struct {
	mock.Mock
}

// CountRunning provides a mock function with given fields:
func (_m *MockPublisherQueue) CountRunning() int {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for CountRunning")
	}

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// Enqueue provides a mock function with given fields: ctx, p
func (_m *MockPublisherQueue) Enqueue(ctx context.Context, p *post.PublishPost) {
	_m.Called(ctx, p)
}

// Start provides a mock function with given fields: ctx
func (_m *MockPublisherQueue) Start(ctx context.Context) {
	_m.Called(ctx)
}

// Stop provides a mock function with given fields:
func (_m *MockPublisherQueue) Stop() {
	_m.Called()
}

// NewMockPublisherQueue creates a new instance of MockPublisherQueue. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockPublisherQueue(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockPublisherQueue {
	mock := &MockPublisherQueue{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
