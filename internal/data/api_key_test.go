package data

import (
	"testing"
	"time"
)

func TestGenerateAPI(t *testing.T) {
	ApiKey := []struct {
		Name   string
		UserID int64
		Expiry time.Time
		Scope  string
		Test   int
		Size   int
		want   int64
	}{
		{Name: "User ID Test", UserID: 1, Expiry: time.Now().Add(72 * time.Hour), Scope: "authentication", Test: 1, Size: 20, want: 1},
		{Name: "User ID Test", UserID: 2, Expiry: time.Now().Add(72 * time.Hour), Scope: "authentication", Test: 1, Size: 20, want: 2},
		{Name: "User ID Test", UserID: 3, Expiry: time.Now().Add(72 * time.Hour), Scope: "authentication", Test: 1, Size: 20, want: 3},
		{Name: "Plaintext Test", UserID: 1, Expiry: time.Now().Add(72 * time.Hour), Scope: "authentication", Test: 2, Size: 20, want: 32},
		{Name: "Plaintext Test", UserID: 2, Expiry: time.Now().Add(72 * time.Hour), Scope: "authentication", Test: 2, Size: 20, want: 32},
		{Name: "Plaintext Test", UserID: 3, Expiry: time.Now().Add(72 * time.Hour), Scope: "authentication", Test: 2, Size: 20, want: 32},
		{Name: "Token Test", UserID: 1, Expiry: time.Now().Add(72 * time.Hour), Scope: "authentication", Test: 2, Size: 16, want: 26},
		{Name: "Token Test", UserID: 2, Expiry: time.Now().Add(72 * time.Hour), Scope: "authentication", Test: 2, Size: 16, want: 26},
		{Name: "Token Test", UserID: 3, Expiry: time.Now().Add(72 * time.Hour), Scope: "authentication", Test: 2, Size: 16, want: 26},
	}
	for _, test := range ApiKey {
		switch test.Test {
		case 1:
			t.Run(test.Name, func(t *testing.T) {
				got, _ := generateAPI(test.UserID, 72*time.Hour, "authentication", test.Size)
				want := test.want

				if got.UserID != want {
					t.Errorf("Got:%v But Wanted:%v", got, want)
				}
			})
		case 2:
			t.Run(test.Name, func(t *testing.T) {
				got, _ := generateAPI(test.UserID, 72*time.Hour, "authentication", test.Size)
				want := test.want
				if len(got.Plaintext) != int(want) {
					t.Errorf("Got:%v But Wanted:%v", got, want)
				}
			})
		}

	}
}
