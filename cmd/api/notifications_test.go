package main

import (
	"testing"
)

func Test_application_refineInterval(t *testing.T) {
	type args struct {
		want     int64
		interval *int64
	}
	values := []int64{10, 1, 2, -5, 100, 234, 67, 0, -1}
	tests := []struct {
		name string
		app  *application
		args args
	}{
		{"Test1", &application{}, args{want: 10, interval: &values[0]}},
		{"Test2", &application{}, args{want: 1, interval: &values[1]}},
		{"Test3", &application{}, args{want: 10, interval: &values[3]}},
		{"Test4", &application{}, args{want: 10, interval: &values[8]}},
		{"Test5", &application{}, args{want: 100, interval: &values[4]}},
		{"Test6", &application{}, args{want: 10, interval: &values[5]}},
		{"Test7", &application{}, args{want: 10, interval: &values[7]}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//configs
			tt.app.config.notifier.interval = 10
			tt.app.config.notifier.deleteinterval = 100
			//t.Log("Settings: ", tt.app.config.notifier.interval, tt.app.config.notifier.deleteinterval)
			got := tt.app.refineInterval(tt.args.interval)
			if got != tt.args.want {
				t.Errorf("refineInterval() = %v, want %v", got, tt.args.want)
			}
		})
	}
}
