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
	ResponseArray struct {
		Code []int `validate:"in:200,404,500"`
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
				Age:    70,
				Email:  "_test@user",
				Role:   "some_role",
				Phones: []string{"212", "495"},
			},
			expectedErr: NewValidationErrors(
				ValidationError{
					Field: "ID",
					Err:   fmt.Errorf("len must be 36, got 32 for 'abcdefghijklmnopqrstuvwxyz012345'"),
				},
				ValidationError{
					Field: "Age",
					Err:   fmt.Errorf("maximum value 50, got 70"),
				},
				ValidationError{
					Field: "Email",
					Err:   fmt.Errorf("fieldValue must match regexp '^\\w+@\\w+\\.\\w+$', actual value '_test@user'"),
				},
				ValidationError{
					Field: "Role",
					Err:   fmt.Errorf("fieldValue must be one of [admin stuff] values, given 'some_role'"),
				},
				ValidationError{
					Field: "Phones",
					Err:   fmt.Errorf("len must be 11, got 3 for '212'"),
				},
				ValidationError{
					Field: "Phones",
					Err:   fmt.Errorf("len must be 11, got 3 for '495'"),
				},
			),
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
			expectedErr: ValidationErrors{{Field: "Code", Err: fmt.Errorf(
				"fieldValue must be one of [200 404 500] values, given 400")}},
		},
		{
			in: ResponseArray{
				Code: []int{200, 404, 500},
			},
			expectedErr: nil,
		},
		{
			in: ResponseArray{
				Code: []int{200, 201, 404, 420},
			},
			expectedErr: NewValidationErrors(
				ValidationError{
					Field: "Code",
					Err:   fmt.Errorf("fieldValue must be one of [200 404 500] values, given 201"),
				},
				ValidationError{
					Field: "Code",
					Err:   fmt.Errorf("fieldValue must be one of [200 404 500] values, given 420"),
				},
			),
		},
		{
			in:          "1",
			expectedErr: NewIllegalArgumentError("expected a struct"),
		},
		{
			in:          struct{}{},
			expectedErr: nil,
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
