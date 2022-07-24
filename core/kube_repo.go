package core

import (
	"context"
	"errors"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

type KubeRepo interface {
	ListWatchedDeploys(ns string) ([]*appsv1.Deployment, error)
	GetConfigMapLatestUpdatedTime(name string, ns string) (time.Time, error)
	GetSecretLatestUpdatedTime(name string, ns string) (time.Time, error)
	RolloutRestartDeployment(name string, ns string) error
}

type kubeRepo struct {
	clientset kubernetes.Interface
	cache     ConfigStore
}

func NewKubeRepo(clientset kubernetes.Interface, cache ConfigStore) KubeRepo {
	return &kubeRepo{clientset, cache}
}

func (k *kubeRepo) ListWatchedDeploys(ns string) ([]*appsv1.Deployment, error) {
	deployments, err := k.clientset.AppsV1().Deployments(ns).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	deploys := make([]*appsv1.Deployment, 0)
	for _, deploy := range deployments.Items {
		if metav1.HasAnnotation(deploy.ObjectMeta, _AnnotationKey_watch) {
			deploys = append(deploys, &deploy)
		}
	}
	return deploys, nil
}

func (k *kubeRepo) GetConfigMapLatestUpdatedTime(name string, ns string) (time.Time, error) {
	cm, err := k.clientset.CoreV1().ConfigMaps(ns).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return time.Time{}, err
	}

	currentApplied, ok := cm.GetAnnotations()["kubectl.kubernetes.io/last-applied-configuration"]
	if !ok {
		return time.Time{}, errors.New("last applied configmap not found")
	}
	t, err := k.getLastTime(ns, _ConfigurationType_ConfigMap, name, currentApplied)
	if err != nil {
		return time.Time{}, err
	}

	return t, nil
}

func (k *kubeRepo) GetSecretLatestUpdatedTime(name string, ns string) (time.Time, error) {
	sc, err := k.clientset.CoreV1().Secrets(ns).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return time.Time{}, err
	}

	currentApplied, ok := sc.GetAnnotations()["kubectl.kubernetes.io/last-applied-configuration"]
	if !ok {
		return time.Time{}, errors.New("last applied secret not found")
	}
	t, err := k.getLastTime(ns, _ConfigurationType_Secret, name, currentApplied)
	if err != nil {
		return time.Time{}, err
	}

	return t, nil
}

func (k *kubeRepo) RolloutRestartDeployment(name string, ns string) error {
	_, err := k.clientset.AppsV1().Deployments(ns).Patch(
		context.Background(),
		name,
		types.StrategicMergePatchType,
		[]byte(fmt.Sprintf(`
		{
			"spec": {
				"template": {
					"metadata": {
						"annotations": {
							"%s": "%s"
						}
					}
				}
			}
		}`,
			_NativeDeploymentAnnotation_RestartedAt, time.Now().Format(time.RFC3339))),
		metav1.PatchOptions{})
	// fmt.Println("patched time:", result.Spec.Template.GetAnnotations()[_NativeDeploymentAnnotation_RestartedAt])
	return err
}

func (k *kubeRepo) getLastTime(ns, typ, name, value string) (time.Time, error) {
	data, err := k.cache.Get(ns, typ, name)
	if err == ErrNotFound {
		data = &Data{
			Value: value,
			Time:  time.Now(),
		}

		if err := k.cache.Set(ns, typ, name, *data); err != nil {
			return time.Time{}, err
		}
	} else if err != nil {
		return time.Time{}, err
	}

	if value != data.Value {
		data = &Data{
			Value: value,
			Time:  time.Now(),
		}
		if err := k.cache.Set(ns, typ, name, *data); err != nil {
			return time.Time{}, err
		}
	}
	return data.Time, nil
}
