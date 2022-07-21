package core

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func LogClusterError(msg string, err error) {
	if err != nil {
		panic(fmt.Sprintf("[cluster] fail to %s: %v", msg, err))
	}
}

type Controller struct {
	clientset *kubernetes.Clientset
}

func (c *Controller) GetDeploymentLatestUpdatedTime(ns string) {
	deployments, err := c.clientset.AppsV1().Deployments(ns).List(context.Background(), metav1.ListOptions{})
	LogClusterError("get deployment in "+ns, err)
	for _, deploy := range deployments.Items {
		if !metav1.HasAnnotation(deploy.ObjectMeta, "deployment-flipper.watch") {
			continue
		}

		conditions := deploy.Status.Conditions
		for _, cond := range conditions {
			fmt.Println(cond)
		}
	}
}

func Main(clientset *kubernetes.Clientset) {
	controller := Controller{clientset}
	for {
		// Core logic:
		// 1. Get all namespace names
		// 	1.1 filter out namespaces with given include/exclude options

		// namespaces, err := clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
		// LogClusterError("get namespaces", err)
		// fmt.Println(len(namespaces.Items))

		// 2. Get all Deployment in that namespaces
		// for _, ns := range namespaces.Items {
		// 	deployments, err := clientset.AppsV1().Deployments(ns.Name).List(context.Background(), metav1.ListOptions{})
		// 	LogClusterError("get deployment in "+ns.Name, err)

		// 	fmt.Println(ns.Name, len(deployments.Items))
		// 	for _, deploy := range deployments.Items {
		// 		fmt.Println(deploy.Namespace, deploy.Name)
		// 		conditions := deploy.Status.Conditions
		// 		for _, cond := range conditions {
		// 			fmt.Println(cond)
		// 		}
		// 	}
		// }

		controller.GetDeploymentLatestUpdatedTime("default")

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

		fmt.Println("---------------------- [cluster] sleep 3s ----------------------")
		time.Sleep(time.Second * 3)
	}
}
