// Code generated by MockGen. DO NOT EDIT.
// Source: interface.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	uuid "github.com/google/uuid"
	model "github.com/vlad-marlo/godo/internal/model"
)

// MockInterface is a mock of Interface interface.
type MockInterface struct {
	ctrl     *gomock.Controller
	recorder *MockInterfaceMockRecorder
}

// MockInterfaceMockRecorder is the mock recorder for MockInterface.
type MockInterfaceMockRecorder struct {
	mock *MockInterface
}

// NewMockInterface creates a new mock instance.
func NewMockInterface(ctrl *gomock.Controller) *MockInterface {
	mock := &MockInterface{ctrl: ctrl}
	mock.recorder = &MockInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockInterface) EXPECT() *MockInterfaceMockRecorder {
	return m.recorder
}

// CreateGroup mocks base method.
func (m *MockInterface) CreateGroup(ctx context.Context, user uuid.UUID, name, description string) (*model.CreateGroupResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateGroup", ctx, user, name, description)
	ret0, _ := ret[0].(*model.CreateGroupResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateGroup indicates an expected call of CreateGroup.
func (mr *MockInterfaceMockRecorder) CreateGroup(ctx, user, name, description interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateGroup", reflect.TypeOf((*MockInterface)(nil).CreateGroup), ctx, user, name, description)
}

// CreateInvite mocks base method.
func (m *MockInterface) CreateInvite(ctx context.Context, user, group uuid.UUID, role *model.Role, limit int) (*model.CreateInviteResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateInvite", ctx, user, group, role, limit)
	ret0, _ := ret[0].(*model.CreateInviteResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateInvite indicates an expected call of CreateInvite.
func (mr *MockInterfaceMockRecorder) CreateInvite(ctx, user, group, role, limit interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateInvite", reflect.TypeOf((*MockInterface)(nil).CreateInvite), ctx, user, group, role, limit)
}

// CreateTask mocks base method.
func (m *MockInterface) CreateTask(ctx context.Context, user uuid.UUID, task model.TaskCreateRequest) (*model.Task, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateTask", ctx, user, task)
	ret0, _ := ret[0].(*model.Task)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateTask indicates an expected call of CreateTask.
func (mr *MockInterfaceMockRecorder) CreateTask(ctx, user, task interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateTask", reflect.TypeOf((*MockInterface)(nil).CreateTask), ctx, user, task)
}

// CreateToken mocks base method.
func (m *MockInterface) CreateToken(ctx context.Context, username, password, token string) (*model.CreateTokenResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateToken", ctx, username, password, token)
	ret0, _ := ret[0].(*model.CreateTokenResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateToken indicates an expected call of CreateToken.
func (mr *MockInterfaceMockRecorder) CreateToken(ctx, username, password, token interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateToken", reflect.TypeOf((*MockInterface)(nil).CreateToken), ctx, username, password, token)
}

// GetMe mocks base method.
func (m *MockInterface) GetMe(ctx context.Context, user uuid.UUID) (*model.GetMeResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMe", ctx, user)
	ret0, _ := ret[0].(*model.GetMeResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMe indicates an expected call of GetMe.
func (mr *MockInterfaceMockRecorder) GetMe(ctx, user interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMe", reflect.TypeOf((*MockInterface)(nil).GetMe), ctx, user)
}

// GetTask mocks base method.
func (m *MockInterface) GetTask(ctx context.Context, user, task uuid.UUID) (*model.Task, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTask", ctx, user, task)
	ret0, _ := ret[0].(*model.Task)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetTask indicates an expected call of GetTask.
func (mr *MockInterfaceMockRecorder) GetTask(ctx, user, task interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTask", reflect.TypeOf((*MockInterface)(nil).GetTask), ctx, user, task)
}

// GetUserFromToken mocks base method.
func (m *MockInterface) GetUserFromToken(ctx context.Context, t string) (uuid.UUID, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserFromToken", ctx, t)
	ret0, _ := ret[0].(uuid.UUID)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserFromToken indicates an expected call of GetUserFromToken.
func (mr *MockInterfaceMockRecorder) GetUserFromToken(ctx, t interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserFromToken", reflect.TypeOf((*MockInterface)(nil).GetUserFromToken), ctx, t)
}

// GetUserTasks mocks base method.
func (m *MockInterface) GetUserTasks(ctx context.Context, user uuid.UUID) (*model.GetTasksResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserTasks", ctx, user)
	ret0, _ := ret[0].(*model.GetTasksResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserTasks indicates an expected call of GetUserTasks.
func (mr *MockInterfaceMockRecorder) GetUserTasks(ctx, user interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserTasks", reflect.TypeOf((*MockInterface)(nil).GetUserTasks), ctx, user)
}

// Ping mocks base method.
func (m *MockInterface) Ping(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Ping", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Ping indicates an expected call of Ping.
func (mr *MockInterfaceMockRecorder) Ping(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ping", reflect.TypeOf((*MockInterface)(nil).Ping), ctx)
}

// RegisterUser mocks base method.
func (m *MockInterface) RegisterUser(ctx context.Context, email, password string) (*model.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RegisterUser", ctx, email, password)
	ret0, _ := ret[0].(*model.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RegisterUser indicates an expected call of RegisterUser.
func (mr *MockInterfaceMockRecorder) RegisterUser(ctx, email, password interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterUser", reflect.TypeOf((*MockInterface)(nil).RegisterUser), ctx, email, password)
}

// UseInvite mocks base method.
func (m *MockInterface) UseInvite(ctx context.Context, user, group, invite uuid.UUID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UseInvite", ctx, user, group, invite)
	ret0, _ := ret[0].(error)
	return ret0
}

// UseInvite indicates an expected call of UseInvite.
func (mr *MockInterfaceMockRecorder) UseInvite(ctx, user, group, invite interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UseInvite", reflect.TypeOf((*MockInterface)(nil).UseInvite), ctx, user, group, invite)
}
