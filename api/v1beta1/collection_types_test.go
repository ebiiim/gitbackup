package v1beta1_test

import (
	"reflect"
	"testing"

	v1beta1 "github.com/ebiiim/gitbackup/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/utils/pointer"
)

func Test_CycleCronByMinuteInSameHour(t *testing.T) {
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
		{"1 2 3 4 sun", args{"1 2 3 4 sun"}, "2 2 3 4 sun", false},
		{"@hourly", args{"@hourly"}, "", true},
		{"-1 * * * *", args{"-1 * * * *"}, "", true},
		{"60 * * * *", args{"60 * * * *"}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := v1beta1.CycleCronByMinuteInSameHour(tt.args.cronStr)
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

func TestCollection_GetOwnedConfigMapName(t *testing.T) {
	type fields struct {
		TypeMeta   metav1.TypeMeta
		ObjectMeta metav1.ObjectMeta
		Spec       v1beta1.CollectionSpec
		Status     v1beta1.CollectionStatus
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"a", fields{ObjectMeta: metav1.ObjectMeta{Name: "a"}}, "gitbackup-collection-a-gitconfig"},
		{"b-c", fields{ObjectMeta: metav1.ObjectMeta{Name: "b-c"}}, "gitbackup-collection-b-c-gitconfig"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := v1beta1.Collection{
				TypeMeta:   tt.fields.TypeMeta,
				ObjectMeta: tt.fields.ObjectMeta,
				Spec:       tt.fields.Spec,
				Status:     tt.fields.Status,
			}
			if got := c.GetOwnedConfigMapName(); got != tt.want {
				t.Errorf("Collection.GetOwnedConfigMapName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCollection_GetOwnedRepositoryNames(t *testing.T) {
	type fields struct {
		TypeMeta   metav1.TypeMeta
		ObjectMeta metav1.ObjectMeta
		Spec       v1beta1.CollectionSpec
		Status     v1beta1.CollectionStatus
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{"a/foo", fields{ObjectMeta: metav1.ObjectMeta{Name: "a"}, Spec: v1beta1.CollectionSpec{
			Repos: []v1beta1.CollectionRepoURL{
				{Name: pointer.String("foo"), Src: "http://example.com/hoge/foo", Dst: "http://example.com/fuga/foo"},
			},
		}}, []string{
			"a-foo",
		}},
		{"a/FOO_2022", fields{ObjectMeta: metav1.ObjectMeta{Name: "a"}, Spec: v1beta1.CollectionSpec{
			Repos: []v1beta1.CollectionRepoURL{
				{Name: nil, Src: "http://example.com/hoge/FOO_2022", Dst: "http://example.com/fuga/FOO_2022"},
			},
		}}, []string{
			"a-foo-2022",
		}},
		{"b-c/foo,bar,baz", fields{ObjectMeta: metav1.ObjectMeta{Name: "b-c"}, Spec: v1beta1.CollectionSpec{
			Repos: []v1beta1.CollectionRepoURL{
				{Name: pointer.String("foo"), Src: "http://example.com/hoge/foo", Dst: "http://example.com/fuga/foo"},
				{Name: pointer.String("bar"), Src: "http://example.com/hoge/barbarbar", Dst: "http://example.com/fuga/barbarbar"},
				{Name: nil, Src: "http://example.com/hoge/baz", Dst: "http://example.com/fuga/bazbazbaz"},
			},
		}}, []string{
			"b-c-foo",
			"b-c-bar",
			"b-c-baz",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := v1beta1.Collection{
				TypeMeta:   tt.fields.TypeMeta,
				ObjectMeta: tt.fields.ObjectMeta,
				Spec:       tt.fields.Spec,
				Status:     tt.fields.Status,
			}
			if got := c.GetOwnedRepositoryNames(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Collection.GetOwnedRepositoryNames() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ToRFC1123(t *testing.T) {
	type args struct {
		s   string
		def string
	}
	tests := []struct {
		args args
		want string
	}{
		{args{"", "invalid-name"}, "invalid-name"},
		{args{".-", "invalid-name"}, "invalid-name"},
		{args{"-.", "invalid-name"}, "invalid-name"},
		{args{"a", "invalid-name"}, "a"},
		{args{"ABC_123", "invalid-name"}, "abc-123"},
		{args{"_._.a_a._._", "invalid-name"}, "a-a"},
		{args{"🤤?a?", "invalid-name"}, "a"},
		{args{"a.b.c-d", "invalid-name"}, "a.b.c-d"},
	}
	for _, tt := range tests {
		t.Run(tt.args.s, func(t *testing.T) {
			got := v1beta1.ToRFC1123(tt.args.s, tt.args.def)
			if got != tt.want {
				t.Errorf("toRFC1123() = %v, want %v", got, tt.want)
			}
			if len(validation.IsDNS1123Subdomain(got)) != 0 {
				t.Errorf("validation.IsDNS1123Subdomain() = %v", validation.IsDNS1123Subdomain(got))
			}
		})
	}
}
