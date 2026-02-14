package testjson_test

import (
	"testing"

	"github.com/AntiD2ta/gosilent/internal/testjson"
	"github.com/stretchr/testify/require"
)

func TestEvent_IsPackageLevel(t *testing.T) {
	tests := []struct {
		name     string
		event    testjson.TestEvent
		expected bool
	}{
		{
			name:     "PackageLevelPass",
			event:    testjson.TestEvent{Action: testjson.ActionPass, Test: ""},
			expected: true,
		},
		{
			name:     "PackageLevelFail",
			event:    testjson.TestEvent{Action: testjson.ActionFail, Test: ""},
			expected: true,
		},
		{
			name:     "PackageLevelSkip",
			event:    testjson.TestEvent{Action: testjson.ActionSkip, Test: ""},
			expected: true,
		},
		{
			name:     "TestLevelPass",
			event:    testjson.TestEvent{Action: testjson.ActionPass, Test: "TestFoo"},
			expected: false,
		},
		{
			name:     "TestLevelOutput",
			event:    testjson.TestEvent{Action: testjson.ActionOutput, Test: "TestFoo"},
			expected: false,
		},
		{
			name:     "PackageLevelOutput",
			event:    testjson.TestEvent{Action: testjson.ActionOutput, Test: ""},
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.expected, test.event.IsPackageLevel())
		})
	}
}
