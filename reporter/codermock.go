// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/coder/xray/reporter (interfaces: CoderClient)
//
// Generated by this command:
//
//	mockgen -destination ./codermock.go -package reporter github.com/coder/xray/reporter CoderClient
//

// Package reporter is a generated GoMock package.
package reporter

import (
	context "context"
	reflect "reflect"

	codersdk "github.com/coder/coder/v2/codersdk"
	agentsdk "github.com/coder/coder/v2/codersdk/agentsdk"
	gomock "go.uber.org/mock/gomock"
)

// MockCoderClient is a mock of CoderClient interface.
type MockCoderClient struct {
	ctrl     *gomock.Controller
	recorder *MockCoderClientMockRecorder
}

// MockCoderClientMockRecorder is the mock recorder for MockCoderClient.
type MockCoderClientMockRecorder struct {
	mock *MockCoderClient
}

// NewMockCoderClient creates a new mock instance.
func NewMockCoderClient(ctrl *gomock.Controller) *MockCoderClient {
	mock := &MockCoderClient{ctrl: ctrl}
	mock.recorder = &MockCoderClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCoderClient) EXPECT() *MockCoderClientMockRecorder {
	return m.recorder
}

// AgentManifest mocks base method.
func (m *MockCoderClient) AgentManifest(arg0 context.Context, arg1 string) (agentsdk.Manifest, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AgentManifest", arg0, arg1)
	ret0, _ := ret[0].(agentsdk.Manifest)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AgentManifest indicates an expected call of AgentManifest.
func (mr *MockCoderClientMockRecorder) AgentManifest(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AgentManifest", reflect.TypeOf((*MockCoderClient)(nil).AgentManifest), arg0, arg1)
}

// PostJFrogXrayScan mocks base method.
func (m *MockCoderClient) PostJFrogXrayScan(arg0 context.Context, arg1 codersdk.JFrogXrayScan) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PostJFrogXrayScan", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// PostJFrogXrayScan indicates an expected call of PostJFrogXrayScan.
func (mr *MockCoderClientMockRecorder) PostJFrogXrayScan(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PostJFrogXrayScan", reflect.TypeOf((*MockCoderClient)(nil).PostJFrogXrayScan), arg0, arg1)
}
