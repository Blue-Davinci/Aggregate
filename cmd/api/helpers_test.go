package main

import "testing"

func Test_application_formatDate(t *testing.T) {
	type args struct {
		date string
	}
	tests := []struct {
		name string
		app  *application
		args args
		want string
	}{
		{
			name: "Valid date with time and timezone",
			app:  &application{},
			args: args{date: "2024-08-03T15:02:32.000Z"},
			want: "2024-08-03 15:02:32",
		},
		{
			name: "Valid date without time",
			app:  &application{},
			args: args{date: "2024-08-03"},
			want: "2024-08-03 00:00:00",
		},
		{
			name: "Invalid date",
			app:  &application{},
			args: args{date: "invalid-date"},
			want: "",
		},
		{
			name: "Empty date",
			app:  &application{},
			args: args{date: ""},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.app.formatDate(tt.args.date); got != tt.want {
				t.Errorf("application.formatDate() = %v, want %v", got, tt.want)
			}
		})
	}
}
