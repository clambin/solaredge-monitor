// Code generated by mockery v2.32.4. DO NOT EDIT.

package mocks

import (
	repository "github.com/clambin/solaredge-monitor/internal/repository"
	mock "github.com/stretchr/testify/mock"
)

// Repository is an autogenerated mock type for the Repository type
type Repository struct {
	mock.Mock
}

type Repository_Expecter struct {
	mock *mock.Mock
}

func (_m *Repository) EXPECT() *Repository_Expecter {
	return &Repository_Expecter{mock: &_m.Mock}
}

// Store provides a mock function with given fields: _a0
func (_m *Repository) Store(_a0 repository.Measurement) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(repository.Measurement) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Repository_Store_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Store'
type Repository_Store_Call struct {
	*mock.Call
}

// Store is a helper method to define mock.On call
//   - _a0 repository.Measurement
func (_e *Repository_Expecter) Store(_a0 interface{}) *Repository_Store_Call {
	return &Repository_Store_Call{Call: _e.mock.On("Store", _a0)}
}

func (_c *Repository_Store_Call) Run(run func(_a0 repository.Measurement)) *Repository_Store_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(repository.Measurement))
	})
	return _c
}

func (_c *Repository_Store_Call) Return(_a0 error) *Repository_Store_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Repository_Store_Call) RunAndReturn(run func(repository.Measurement) error) *Repository_Store_Call {
	_c.Call.Return(run)
	return _c
}

// NewRepository creates a new instance of Repository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *Repository {
	mock := &Repository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}