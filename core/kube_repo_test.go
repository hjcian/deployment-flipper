package core

import (
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_KubeRepo_ListWatchedDeploys(t *testing.T) {
	tests := []struct {
		name          string
		giveNamespace string
		giveDeploys   []runtime.Object
		wantError     error
		wantDeploys   []*appsv1.Deployment
	}{
		{
			name:          "empty",
			giveNamespace: "default",
			giveDeploys:   []runtime.Object{},
			wantDeploys:   []*appsv1.Deployment{},
		},
		{
			name:          "test one watched deployment",
			giveNamespace: "default",
			giveDeploys: []runtime.Object{
				&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "watched-deploy",
						Namespace: "default",
						Annotations: map[string]string{
							"deployment-flipper.watch": "true",
						},
					},
				},
				&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "non-watched-deploy",
						Namespace: "default",
					},
				},
				&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "non-watched-deploy",
						Namespace: "others",
						Annotations: map[string]string{
							"deployment-flipper.watch": "true",
						},
					},
				},
			},
			wantDeploys: []*appsv1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "watched-deploy",
						Namespace: "default",
						Annotations: map[string]string{
							"deployment-flipper.watch": "true",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fakeClientset := fake.NewSimpleClientset(tt.giveDeploys...)
			k := NewKubeRepo(fakeClientset, NewConfigStore())

			gotDeploys, err := k.ListWatchedDeploys(tt.giveNamespace)
			if tt.wantError != nil {
				require.Error(t, err)
			} else {
				require.Equal(t, len(tt.wantDeploys), len(gotDeploys))
				require.ElementsMatch(t, tt.wantDeploys, gotDeploys)
			}
		})
	}
}
