package mocks

import "go.uber.org/mock/gomock"

// NewRollingCollector is a MockCollector for a test that rolls with a real registry
// and does not care about the roll cards: each card asks and takes nothing.
func NewRollingCollector(ctrl *gomock.Controller) *MockCollector {
	collector := NewMockCollector(ctrl)
	collector.EXPECT().SelectModifiers(gomock.Any()).AnyTimes()
	collector.EXPECT().SelectReroll(gomock.Any()).AnyTimes()
	return collector
}
