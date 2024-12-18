// Code generated by mockery v2.50.0. DO NOT EDIT.

package mocks

import (
	io "io"

	repository "github.com/clambin/solaredge-monitor/internal/repository"
	mock "github.com/stretchr/testify/mock"
)

// Plotter is an autogenerated mock type for the Plotter type
type Plotter struct {
	mock.Mock
}

type Plotter_Expecter struct {
	mock *mock.Mock
}

func (_m *Plotter) EXPECT() *Plotter_Expecter {
	return &Plotter_Expecter{mock: &_m.Mock}
}

// Plot provides a mock function with given fields: _a0, _a1, _a2
func (_m *Plotter) Plot(_a0 io.Writer, _a1 repository.Measurements, _a2 bool) (int64, error) {
	ret := _m.Called(_a0, _a1, _a2)

	if len(ret) == 0 {
		panic("no return value specified for Plot")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(io.Writer, repository.Measurements, bool) (int64, error)); ok {
		return rf(_a0, _a1, _a2)
	}
	if rf, ok := ret.Get(0).(func(io.Writer, repository.Measurements, bool) int64); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(io.Writer, repository.Measurements, bool) error); ok {
		r1 = rf(_a0, _a1, _a2)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Plotter_Plot_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Plot'
type Plotter_Plot_Call struct {
	*mock.Call
}

// Plot is a helper method to define mock.On call
//   - _a0 io.Writer
//   - _a1 repository.Measurements
//   - _a2 bool
func (_e *Plotter_Expecter) Plot(_a0 interface{}, _a1 interface{}, _a2 interface{}) *Plotter_Plot_Call {
	return &Plotter_Plot_Call{Call: _e.mock.On("Plot", _a0, _a1, _a2)}
}

func (_c *Plotter_Plot_Call) Run(run func(_a0 io.Writer, _a1 repository.Measurements, _a2 bool)) *Plotter_Plot_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(io.Writer), args[1].(repository.Measurements), args[2].(bool))
	})
	return _c
}

func (_c *Plotter_Plot_Call) Return(_a0 int64, _a1 error) *Plotter_Plot_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Plotter_Plot_Call) RunAndReturn(run func(io.Writer, repository.Measurements, bool) (int64, error)) *Plotter_Plot_Call {
	_c.Call.Return(run)
	return _c
}

// NewPlotter creates a new instance of Plotter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewPlotter(t interface {
	mock.TestingT
	Cleanup(func())
}) *Plotter {
	mock := &Plotter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
