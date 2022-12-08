package controllers

import "testing"

func Test_cycleCronByMinuteInSameHour(t *testing.T) {
	type args struct {
		cronStr string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"0 * * * *", args{"0 * * * *"}, "1 * * * *", false},
		{"1 * * * *", args{"1 * * * *"}, "2 * * * *", false},
		{"30 * * * *", args{"30 * * * *"}, "31 * * * *", false},
		{"58 * * * *", args{"58 * * * *"}, "59 * * * *", false},
		{"59 * * * *", args{"59 * * * *"}, "0 * * * *", false},
		{"1 2 3 4 5", args{"1 2 3 4 5"}, "2 2 3 4 5", false},
		{"@hourly", args{"@hourly"}, "", true},
		{"-1 * * * *", args{"-1 * * * *"}, "", true},
		{"60 * * * *", args{"60 * * * *"}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cycleCronByMinuteInSameHour(tt.args.cronStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("cycleCronByMinuteInSameHour() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("cycleCronByMinuteInSameHour() = %v, want %v", got, tt.want)
			}
		})
	}
}
