package utils

import (
	"fmt"
	"testing"
	"time"
)

var errorTests = []struct {
	name        string
	value       string
	expectedErr string
}{
	{
		name:        "not valid unit",
		value:       "1n",
		expectedErr: fmt.Sprintf(ERROR_NotValidUnit, "n"),
	},
	{
		name:        "valid unit but not valid value",
		value:       "as",
		expectedErr: fmt.Sprintf(ERROR_NotValidNumberValue, "a"),
	},
	{
		name:        "valid unit and value equals 0",
		value:       "0s",
		expectedErr: fmt.Sprintf(ERROR_NotValidQuantity, "0"),
	},
	{
		name:        "empty value",
		value:       "",
		expectedErr: ERROR_NotValidTime,
	},
}

var timeTests = []struct {
	name     string
	value    string
	expected time.Duration
}{
	{
		name:     "parsed seconds",
		value:    "20s",
		expected: time.Duration(20) * time.Second,
	},
	{
		name:     "parsed minutes",
		value:    "25m",
		expected: time.Duration(25) * time.Minute,
	},
	{
		name:     "parsed hours",
		value:    "25h",
		expected: time.Duration(25) * time.Hour,
	},
	{
		name:     "parsed days",
		value:    "7d",
		expected: time.Duration(7) * 24 * time.Hour,
	},
}

func TestValidateStringTime(t *testing.T) {
	for _, test := range errorTests {
		t.Run(test.name, func(t *testing.T) {
			_, err := MakeTimeFromString(test.value)
			if err.Error() != test.expectedErr {
				t.Errorf("not valid error %q, expected %q", err, test.expectedErr)
			}
		})
	}

	for _, test := range timeTests {
		t.Run(test.name, func(t *testing.T) {
			value, err := MakeTimeFromString(test.value)
			if err != nil {
				t.Errorf("not expected error: %v", err)
			}
			if value != test.expected {
				t.Errorf("not valid time %q, expected %q", value, test.expected)
			}
		})
	}
}
