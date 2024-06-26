package data

import (
	"fmt"
	"testing"
)

func TestValidateUUID(t *testing.T) {
	myUUID := "7a467466-e92c-44ed-a206-3a460c867f3b"
	myUUID2 := "c5f950a6-290d-4a38-a817-f49f2bb52b2f"
	myUUID3 := "abcd123342123"
	uuidTests := []struct {
		Name string
		UUID string
		want bool
	}{
		{Name: "User UUID Test", UUID: myUUID, want: true},
		{Name: "User UUID Test1", UUID: myUUID2, want: true},
		{Name: "User UUID Test2", UUID: myUUID3, want: false},
	}
	for _, test := range uuidTests {
		t.Run(test.Name, func(t *testing.T) {
			_, got := ValidateUUID(test.UUID)
			want := test.want
			fmt.Println(test.UUID)
			if got != want {
				t.Errorf("[%v]Got:%v But Wanted:%v", test.UUID, got, want)
			}
		})
	}

}
