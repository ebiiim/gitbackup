package v1beta1

import "testing"

func Test_testIsValidURLSet(t *testing.T) {
	type args struct {
		s []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"0", args{s: []string{}}, true},
		{"1", args{s: []string{"http://example.com/src/foo"}}, true},
		{"2", args{s: []string{"http://example.com/src/foo", "http://example.com/dst/foo"}}, true},
		{"3", args{s: []string{"http://example.com/src/foo", "http://example.com/dst/foo", "http://example.com/dst2/foo"}}, true},
		{"1x", args{s: []string{"http://example.com/s c/foo"}}, false},
		{"2x", args{s: []string{"http://example.com/src/foo", "http://example.com/src/foo"}}, false},
		{"3x", args{s: []string{"http://example.com/src/foo", "http://example.com/dst/foo", "http://example.com/dst/foo"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidURLSet(tt.args.s...); got != tt.want {
				t.Errorf("testURLSet() = %v, want %v", got, tt.want)
			}
		})
	}
}
