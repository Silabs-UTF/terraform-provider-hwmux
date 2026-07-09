package hwmux

import (
	"net/http"
	"testing"
)

func TestIsDeviceAlreadyActiveStatusError(t *testing.T) {
	tests := []struct {
		name     string
		httpRes  *http.Response
		body     string
		expected bool
	}{
		{
			name:     "matches exact known message",
			httpRes:  &http.Response{StatusCode: http.StatusBadRequest},
			body:     "Device XXXXX already has status ACTIVE",
			expected: true,
		},
		{
			name:     "matches case-insensitive message",
			httpRes:  &http.Response{StatusCode: http.StatusBadRequest},
			body:     "device y already has status active",
			expected: true,
		},
		{
			name:     "does not match different 400 error",
			httpRes:  &http.Response{StatusCode: http.StatusBadRequest},
			body:     "Device is in use",
			expected: false,
		},
		{
			name:     "does not match non-400 error",
			httpRes:  &http.Response{StatusCode: http.StatusConflict},
			body:     "Device XXXXX already has status ACTIVE",
			expected: false,
		},
		{
			name:     "does not match nil response",
			httpRes:  nil,
			body:     "Device XXXXX already has status ACTIVE",
			expected: false,
		},
		{
			name:     "does not match empty body",
			httpRes:  &http.Response{StatusCode: http.StatusBadRequest},
			body:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isDeviceAlreadyActiveStatusError(tt.httpRes, tt.body)
			if got != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestIsLabelInUseDeleteError(t *testing.T) {
	tests := []struct {
		name     string
		httpRes  *http.Response
		body     string
		expected bool
	}{
		{
			name:     "matches exact known message",
			httpRes:  &http.Response{StatusCode: http.StatusBadRequest},
			body:     "Label is in use and cannot be deleted",
			expected: true,
		},
		{
			name:     "matches case-insensitive message",
			httpRes:  &http.Response{StatusCode: http.StatusBadRequest},
			body:     "label is in use and cannot be deleted",
			expected: true,
		},
		{
			name:     "does not match different 400 error",
			httpRes:  &http.Response{StatusCode: http.StatusBadRequest},
			body:     "Label name already exists",
			expected: false,
		},
		{
			name:     "does not match non-400 error",
			httpRes:  &http.Response{StatusCode: http.StatusConflict},
			body:     "Label is in use and cannot be deleted",
			expected: false,
		},
		{
			name:     "does not match nil response",
			httpRes:  nil,
			body:     "Label is in use and cannot be deleted",
			expected: false,
		},
		{
			name:     "does not match empty body",
			httpRes:  &http.Response{StatusCode: http.StatusBadRequest},
			body:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isLabelInUseDeleteError(tt.httpRes, tt.body)
			if got != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestIsReservationBlockedDeleteError(t *testing.T) {
	tests := []struct {
		name     string
		httpRes  *http.Response
		body     string
		expected bool
	}{
		{
			name:     "matches active reservations wording",
			httpRes:  &http.Response{StatusCode: http.StatusBadRequest},
			body:     "Device has active reservations and cannot be deleted",
			expected: true,
		},
		{
			name:     "matches queued reservations wording",
			httpRes:  &http.Response{StatusCode: http.StatusBadRequest},
			body:     "Device group has queued reservations",
			expected: true,
		},
		{
			name:     "matches case-insensitive wording",
			httpRes:  &http.Response{StatusCode: http.StatusBadRequest},
			body:     "DEVICE HAS ACTIVE RESERVATIONS AND CANNOT BE DELETED",
			expected: true,
		},
		{
			name:     "does not match unrelated 400",
			httpRes:  &http.Response{StatusCode: http.StatusBadRequest},
			body:     "Invalid device id",
			expected: false,
		},
		{
			name:     "does not match non-400",
			httpRes:  &http.Response{StatusCode: http.StatusConflict},
			body:     "active reservations",
			expected: false,
		},
		{
			name:     "does not match nil response",
			httpRes:  nil,
			body:     "active reservations",
			expected: false,
		},
		{
			name:     "does not match empty body",
			httpRes:  &http.Response{StatusCode: http.StatusBadRequest},
			body:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isReservationBlockedDeleteError(tt.httpRes, tt.body)
			if got != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}
