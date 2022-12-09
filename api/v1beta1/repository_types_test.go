package v1beta1_test

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1beta1 "github.com/ebiiim/gitbackup/api/v1beta1"
)

func TestRepository_GetOwnedConfigMapName(t *testing.T) {
	type fields struct {
		TypeMeta   metav1.TypeMeta
		ObjectMeta metav1.ObjectMeta
		Spec       v1beta1.RepositorySpec
		Status     v1beta1.RepositoryStatus
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"a", fields{ObjectMeta: metav1.ObjectMeta{Name: "a"}}, "gitbackup-repository-a-gitconfig"},
		{"b-c", fields{ObjectMeta: metav1.ObjectMeta{Name: "b-c"}}, "gitbackup-repository-b-c-gitconfig"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := v1beta1.Repository{
				TypeMeta:   tt.fields.TypeMeta,
				ObjectMeta: tt.fields.ObjectMeta,
				Spec:       tt.fields.Spec,
				Status:     tt.fields.Status,
			}
			if got := r.GetOwnedConfigMapName(); got != tt.want {
				t.Errorf("Repository.GetOwnedConfigMapName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepository_GetOwnedCronJobName(t *testing.T) {
	type fields struct {
		TypeMeta   metav1.TypeMeta
		ObjectMeta metav1.ObjectMeta
		Spec       v1beta1.RepositorySpec
		Status     v1beta1.RepositoryStatus
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"a", fields{ObjectMeta: metav1.ObjectMeta{Name: "a"}}, "gitbackup-a"},
		{"b-c", fields{ObjectMeta: metav1.ObjectMeta{Name: "b-c"}}, "gitbackup-b-c"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := v1beta1.Repository{
				TypeMeta:   tt.fields.TypeMeta,
				ObjectMeta: tt.fields.ObjectMeta,
				Spec:       tt.fields.Spec,
				Status:     tt.fields.Status,
			}
			if got := r.GetOwnedCronJobName(); got != tt.want {
				t.Errorf("Repository.GetOwnedCronJobName() = %v, want %v", got, tt.want)
			}
		})
	}
}
