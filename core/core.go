package core

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type empty struct{}

const (
	_AnnotationKey_watch = "deployment-flipper.watch"
)

func LogClusterError(msg string, err error) {
	if err != nil {
		panic(fmt.Sprintf("[cluster] fail to %s: %v", msg, err))
	}
}

type Controller struct {
	clientset *kubernetes.Clientset
}

func (c *Controller) ListWatchedDeploys(ns string) ([]*appsv1.Deployment, error) {
	deployments, err := c.clientset.AppsV1().Deployments(ns).List(context.Background(), metav1.ListOptions{})
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

func (c *Controller) GetCMLatestUpdatedTime(name string, ns string) (*corev1.ConfigMap, error) {
	return c.clientset.CoreV1().ConfigMaps(ns).Get(context.Background(), name, metav1.GetOptions{})
}

func checkStableAndGetLatestUpdatedTime(deploy *appsv1.Deployment) (time.Time, bool) {
	var available bool
	var availableLastUpdateTime time.Time
	var processing bool
	var processingLastUpdateTime time.Time
	for _, cond := range deploy.Status.Conditions {
		if cond.Type == appsv1.DeploymentAvailable && cond.Status == corev1.ConditionTrue {
			available = true
			availableLastUpdateTime = cond.LastUpdateTime.Time
		}
		if cond.Type == appsv1.DeploymentProgressing && cond.Status == corev1.ConditionTrue {
			processing = true
			processingLastUpdateTime = cond.LastUpdateTime.Time
		}
	}

	if available && processing {
		if availableLastUpdateTime.After(processingLastUpdateTime) {
			return availableLastUpdateTime, true
		} else {
			return processingLastUpdateTime, true
		}
	}
	return time.Time{}, false
}

func listReferencedConfigurations(deploy *appsv1.Deployment) (map[string]empty, map[string]empty) {
	configMapNames := make(map[string]empty, 0)
	secretNames := make(map[string]empty, 0)

	for _, container := range deploy.Spec.Template.Spec.Containers {
		for _, env := range container.Env {
			if env.ValueFrom == nil {
				continue
			}

			if env.ValueFrom.ConfigMapKeyRef != nil {
				configMapNames[env.ValueFrom.ConfigMapKeyRef.Name] = empty{}
			}
			if env.ValueFrom.SecretKeyRef != nil {
				secretNames[env.ValueFrom.SecretKeyRef.Name] = empty{}
			}

			for _, envFrom := range container.EnvFrom {
				if envFrom.ConfigMapRef != nil {
					configMapNames[envFrom.ConfigMapRef.Name] = empty{}
				}
				if envFrom.SecretRef != nil {
					secretNames[envFrom.SecretRef.Name] = empty{}
				}
			}
		}
	}
	return configMapNames, secretNames
}

func (c *Controller) Do() {
	ns := "default"
	deploys, err := c.ListWatchedDeploys(ns)
	LogClusterError("get deployment in "+ns, err)

	stableDeploys := make(map[*appsv1.Deployment]time.Time, 0)
	for _, deploy := range deploys {
		if lastUpdateTime, ok := checkStableAndGetLatestUpdatedTime(deploy); ok {
			stableDeploys[deploy] = lastUpdateTime
		}
	}

	for deploy, lastUpdateTime := range stableDeploys {
		fmt.Println("deployment:", deploy.Name, "lastUpdateTime:", lastUpdateTime)
		configMapNames, secretNames := listReferencedConfigurations(deploy)
		if len(configMapNames) > 0 {
			// fmt.Println(configMapNames)
			for name := range configMapNames {
				cm, err := c.GetCMLatestUpdatedTime(name, ns)
				LogClusterError("get configmap "+name, err)
				fmt.Println(cm.Name, cm.CreationTimestamp)
			}
		}
		if len(secretNames) > 0 {
			fmt.Println(secretNames)
		}
	}
}

func Main(clientset *kubernetes.Clientset) {
	controller := Controller{clientset}
	for {
		// Core logic:
		// 1. Get all namespace names
		// 	1.1 filter out namespaces with given include/exclude options
		// 2. Get all Deployment in that namespaces
		controller.Do()

		// 3. Get all referenced Secrets/ConfigMaps in each Deployment
		// 4. Compare updated time on Deployment and Secrets/ConfigMaps
		// 5. Restart the Deployment if its updated time is older than its referenced Secrets/ConfigMaps

		// pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
		// if err != nil {
		// 	panic(err.Error())
		// }
		// fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

		// // Examples for error handling:
		// // - Use helper functions like e.g. errors.IsNotFound()
		// // - And/or cast to StatusError and use its properties like e.g. ErrStatus.Message
		// namespace := "default"
		// pod := "example-xxxxx"
		// _, err = clientset.CoreV1().Pods(namespace).Get(context.TODO(), pod, metav1.GetOptions{})
		// if errors.IsNotFound(err) {
		// 	fmt.Printf("Pod %s in namespace %s not found\n", pod, namespace)
		// } else if statusError, isStatus := err.(*errors.StatusError); isStatus {
		// 	fmt.Printf("Error getting pod %s in namespace %s: %v\n",
		// 		pod, namespace, statusError.ErrStatus.Message)
		// } else if err != nil {
		// 	panic(err.Error())
		// } else {
		// 	fmt.Printf("Found pod %s in namespace %s\n", pod, namespace)
		// }

		fmt.Println("---------------------- [cluster] sleep ----------------------")
		time.Sleep(time.Second * 2)
	}
}
