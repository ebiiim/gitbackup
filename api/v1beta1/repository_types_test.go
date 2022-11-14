/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
		{"a", fields{ObjectMeta: metav1.ObjectMeta{Name: "a"}}, "gitbackup-gitconfig-a"},
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
