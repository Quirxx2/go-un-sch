// Code generated by mockery v2.15.0. DO NOT EDIT.

package golangunitedschoolcerts

import mock "github.com/stretchr/testify/mock"

// MockTemplater is an autogenerated mock type for the Templater type
type MockTemplater struct {
	mock.Mock
}

type MockTemplater_Expecter struct {
	mock *mock.Mock
}

func (_m *MockTemplater) EXPECT() *MockTemplater_Expecter {
	return &MockTemplater_Expecter{mock: &_m.Mock}
}

// GenerateCertificate provides a mock function with given fields: template, certificate, link
func (_m *MockTemplater) GenerateCertificate(template string, certificate *Certificate, link string) (*[]byte, error) {
	ret := _m.Called(template, certificate, link)

	var r0 *[]byte
	if rf, ok := ret.Get(0).(func(string, *Certificate, string) *[]byte); ok {
		r0 = rf(template, certificate, link)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*[]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, *Certificate, string) error); ok {
		r1 = rf(template, certificate, link)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockTemplater_GenerateCertificate_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GenerateCertificate'
type MockTemplater_GenerateCertificate_Call struct {
	*mock.Call
}

// GenerateCertificate is a helper method to define mock.On call
//   - template string
//   - certificate *Certificate
//   - link string
func (_e *MockTemplater_Expecter) GenerateCertificate(template interface{}, certificate interface{}, link interface{}) *MockTemplater_GenerateCertificate_Call {
	return &MockTemplater_GenerateCertificate_Call{Call: _e.mock.On("GenerateCertificate", template, certificate, link)}
}

func (_c *MockTemplater_GenerateCertificate_Call) Run(run func(template string, certificate *Certificate, link string)) *MockTemplater_GenerateCertificate_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(*Certificate), args[2].(string))
	})
	return _c
}

func (_c *MockTemplater_GenerateCertificate_Call) Return(_a0 *[]byte, _a1 error) *MockTemplater_GenerateCertificate_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

type mockConstructorTestingTNewMockTemplater interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockTemplater creates a new instance of MockTemplater. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockTemplater(t mockConstructorTestingTNewMockTemplater) *MockTemplater {
	mock := &MockTemplater{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
