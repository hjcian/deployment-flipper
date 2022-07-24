package core

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_KubeRepo_ListWatchedDeploys(t *testing.T) {
	tests := []struct {
		name          string
		giveNamespace string
		giveDeploys   []runtime.Object
		wantDeploys   []*appsv1.Deployment
	}{
		{
			name:          "empty",
			giveNamespace: "default",
			giveDeploys:   []runtime.Object{},
			wantDeploys:   []*appsv1.Deployment{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fakeClientset := fake.NewSimpleClientset(tt.giveDeploys...)
			k := NewKubeRepo(fakeClientset, NewConfigStore())

			gotDeploys, err := k.ListWatchedDeploys(tt.giveNamespace)
			if err != nil {
				t.Errorf("ListWatchedDeploys() error = %v", err)
				return
			}
			if len(gotDeploys) != len(tt.wantDeploys) {
				t.Errorf("ListWatchedDeploys() gotDeploys = %v, want %v", gotDeploys, tt.wantDeploys)
			}
		})
	}
}
