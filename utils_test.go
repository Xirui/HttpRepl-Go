package main

import (
	"reflect"
	"testing"
)

func TestSplitArgs(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
		wantErr  bool
	}{
		{
			input:    `get`,
			expected: []string{"get"},
			wantErr:  false,
		},
		{
			input:    `get subpath`,
			expected: []string{"get", "subpath"},
			wantErr:  false,
		},
		{
			input:    `post --content "{\"id\": 1}"`,
			expected: []string{"post", "--content", `{"id": 1}`},
			wantErr:  false,
		},
		{
			input:    `set header "Authorization" "Bearer 12345"`,
			expected: []string{"set", "header", "Authorization", "Bearer 12345"},
			wantErr:  false,
		},
		{
			input:    `post -c '{"name": "hello"}'`,
			expected: []string{"post", "-c", `{"name": "hello"}`},
			wantErr:  false,
		},
		{
			input:    `post --content "unclosed`,
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := splitArgs(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("splitArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("splitArgs() = %v, want %v", got, tt.expected)
			}
		})
	}
}
