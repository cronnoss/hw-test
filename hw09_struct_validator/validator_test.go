package hw09structvalidator

import (
	"encoding/json"
	"fmt"
	"testing"
)

type UserRole string

// Test the function on different structures and other types.
type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int             `validate:"min:18|max:50"`
		Email  string          `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole        `validate:"in:admin,stuff"`
		Phones []string        `validate:"len:11"`
		meta   json.RawMessage //nolint:unused
	}

	App struct {
		Version string `validate:"len:5"`
	}

	Token struct {
		Header    []byte
		Payload   []byte
		Signature []byte
	}

	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `json:"omitempty"`
	}
)

func TestValidate(t *testing.T) {
	tests := []struct {
		in          interface{}
		expectedErr error
	}{
		{
			in: User{
				ID:     "abcdefghijklmnopqrstuvwxyz0123456789",
				Name:   "Test User",
				Age:    25,
				Email:  "testuser@gmail.com",
				Role:   "admin",
				Phones: []string{"12345678901", "10987654321"},
			},
			expectedErr: nil,
		},
		{
			in: User{
				ID:     "abcdefghijklmnopqrstuvwxyz012345",
				Name:   "Test User",
				Age:    25,
				Email:  "testuser@gmail.com",
				Role:   "admin",
				Phones: []string{"12345678901", "10987654321"},
			},
			expectedErr: ValidationErrors{{Field: "ID", Err: fmt.Errorf("expected length of 36, got 32")}},
		},
		{
			in: User{
				ID:     "abcdefghijklmnopqrstuvwxyz0123456789",
				Name:   "Long name_Long name_Long name_Long name_Long name_Long name_Long name_Long name_Long name_",
				Age:    25,
				Email:  "testuser@gmail.com",
				Role:   "admin",
				Phones: []string{"12345678901", "10987654321"},
			},
			expectedErr: nil,
		},
		{
			in: User{
				ID:     "abcdefghijklmnopqrstuvwxyz0123456789",
				Name:   "Test User",
				Age:    70,
				Email:  "testuser@gmail.com",
				Role:   "admin",
				Phones: []string{"12345678901", "10987654321"},
			},
			expectedErr: ValidationErrors{{Field: "Age", Err: fmt.Errorf("number is more than maximum of 50")}},
		},
		{
			in: User{
				ID:     "abcdefghijklmnopqrstuvwxyz0123456789",
				Name:   "Test User",
				Age:    25,
				Email:  "_test@user",
				Role:   "admin",
				Phones: []string{"12345678901", "10987654321"},
			},
			expectedErr: ValidationErrors{{Field: "Email", Err: fmt.Errorf("string does not match regexp ^\\w+@\\w+\\.\\w+$")}},
		},
		{
			in: User{
				ID:     "abcdefghijklmnopqrstuvwxyz0123456789",
				Name:   "Test User",
				Age:    25,
				Email:  "testuser@gmail.com",
				Role:   "some_role",
				Phones: []string{"12345678901", "10987654321"},
			},
			expectedErr: ValidationErrors{{Field: "Role", Err: fmt.Errorf("string some_role is not in set [admin stuff]")}},
		},
		{
			in: User{
				ID:     "abcdefghijklmnopqrstuvwxyz0123456789",
				Name:   "Test User",
				Age:    25,
				Email:  "testuser@gmail.com",
				Role:   "admin",
				Phones: []string{"12345678901", "4"},
			},
			expectedErr: ValidationErrors{{Field: "Phones", Err: fmt.Errorf("expected length of 11, got 1")}},
		},
		{
			in: User{
				ID:     "abcdefghijklmnopqrstuvwxyz0123456789",
				Name:   "Test User",
				Age:    25,
				Email:  "testuser@gmail.com",
				Role:   "admin",
				Phones: []string{"+12345678901"},
			},
			expectedErr: ValidationErrors{{Field: "Phones", Err: fmt.Errorf("expected length of 11, got 12")}},
		},
		{
			in: App{
				Version: "1.0.0",
			},
			expectedErr: nil,
		},
		{
			in: Token{
				Header:    []byte("header"),
				Payload:   []byte("payload"),
				Signature: []byte("signature"),
			},
			expectedErr: nil,
		},
		{
			in: Response{
				Code: 404,
				Body: "Page not found",
			},
			expectedErr: nil,
		},
		{
			in: Response{
				Code: 400,
				Body: "Bad Request",
			},
			expectedErr: ValidationErrors{{Field: "Code", Err: fmt.Errorf("number 400 is not in set [200 404 500]")}},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			tt := tt
			t.Parallel()

			err := Validate(tt.in)
			if err != nil {
				if tt.expectedErr == nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if err.Error() != tt.expectedErr.Error() {
					t.Fatalf("expected error: %v, got: %v", tt.expectedErr, err)
				}
			} else if tt.expectedErr != nil {
				t.Fatalf("expected error: %v, got: %v", tt.expectedErr, err)
			}
			_ = tt
		})
	}
}
