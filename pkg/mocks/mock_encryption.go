// Code generated by MockGen. DO NOT EDIT.
// Source: cellar/pkg/cryptography (interfaces: Encryption)
//
// Generated by this command:
//
//	mockgen -destination=../mocks/mock_encryption.go -package=mocks . Encryption
//

// Package mocks is a generated GoMock package.
package mocks

import (
	models "cellar/pkg/models"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockEncryption is a mock of Encryption interface.
type MockEncryption struct {
	ctrl     *gomock.Controller
	recorder *MockEncryptionMockRecorder
	isgomock struct{}
}

// MockEncryptionMockRecorder is the mock recorder for MockEncryption.
type MockEncryptionMockRecorder struct {
	mock *MockEncryption
}

// NewMockEncryption creates a new mock instance.
func NewMockEncryption(ctrl *gomock.Controller) *MockEncryption {
	mock := &MockEncryption{ctrl: ctrl}
	mock.recorder = &MockEncryptionMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockEncryption) EXPECT() *MockEncryptionMockRecorder {
	return m.recorder
}

// Decrypt mocks base method.
func (m *MockEncryption) Decrypt(content string) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Decrypt", content)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Decrypt indicates an expected call of Decrypt.
func (mr *MockEncryptionMockRecorder) Decrypt(content any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Decrypt", reflect.TypeOf((*MockEncryption)(nil).Decrypt), content)
}

// Encrypt mocks base method.
func (m *MockEncryption) Encrypt(content []byte) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Encrypt", content)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Encrypt indicates an expected call of Encrypt.
func (mr *MockEncryptionMockRecorder) Encrypt(content any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Encrypt", reflect.TypeOf((*MockEncryption)(nil).Encrypt), content)
}

// Health mocks base method.
func (m *MockEncryption) Health() models.Health {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Health")
	ret0, _ := ret[0].(models.Health)
	return ret0
}

// Health indicates an expected call of Health.
func (mr *MockEncryptionMockRecorder) Health() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Health", reflect.TypeOf((*MockEncryption)(nil).Health))
}
