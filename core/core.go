package core

import (
	"time"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes"
)

type empty struct{}

const (
	_AnnotationKey_watch         = "deployment-flipper.watch"
	_ConfigurationType_ConfigMap = "configmap"
	_ConfigurationType_Secret    = "secret"

	_NativeDeploymentAnnotation_RestartedAt = "kubectl.kubernetes.io/restartedAt"
	_NewRSAvailableReason                   = "NewReplicaSetAvailable"
)

type Controller struct {
	logger *zap.Logger
	repo   KubeRepo
}

func (c *Controller) Do() {
	ns := "default"
	deploys, err := c.repo.ListWatchedDeploys(ns)
	if err != nil {
		c.logger.Error("list watched deploys error:", zap.Error(err))
		return
	}

	stableDeploys := make(map[*appsv1.Deployment]time.Time, 0)
	for _, deploy := range deploys {
		if deployTime, ok := checkStableAndGetLatestUpdatedTime(deploy); ok {
			stableDeploys[deploy] = deployTime
		}
	}
	var restart bool
	for deploy, deployTime := range stableDeploys {
		logger := c.logger.With(
			zap.String("ns", ns),
			zap.String("deploy_name", deploy.Name),
			zap.Time("deploy_time", deployTime),
		)
		logger.Info("found stable watched deploy")

		configMapNames, secretNames := listReferencedConfigurations(deploy)
		if len(configMapNames) > 0 {
			for name := range configMapNames {
				configMapTime, err := c.repo.GetConfigMapLatestUpdatedTime(name, ns)
				if err != nil {
					logger.Error("get updated time error:",
						zap.String("config_name", name),
						zap.Error(err))
					continue
				}

				if configMapTime.After(deployTime) {
					logger.Info("found updated ConfigMap:",
						zap.String("config_name", name),
						zap.Time("config_time", configMapTime),
					)
					restart = true
				}
			}
		}
		if len(secretNames) > 0 {
			for name := range secretNames {
				secretTime, err := c.repo.GetSecretLatestUpdatedTime(name, ns)
				if err != nil {
					logger.Error("get updated time error:",
						zap.String("secret_name", name),
						zap.Error(err))
					continue
				}

				if secretTime.After(deployTime) {
					logger.Info("found updated ConfigMap:",
						zap.String("secret_name", name),
						zap.Time("secret_time", secretTime),
					)
					restart = true
				}
			}
		}
		if restart {
			if err := c.repo.RolloutRestartDeployment(deploy.Name, ns); err != nil {
				logger.Error("rollout restart deployment error:", zap.Error(err))
			} else {
				logger.Info("deployment is restarted")
			}
		}
	}
}

func Main(clientset kubernetes.Interface) {
	logger, _ := zap.NewDevelopment()
	controller := Controller{logger, NewKubeRepo(clientset, NewConfigStore())}
	for {
		// Initialization
		// load configurations from its CRD

		// Core logic:
		// 1. Get all namespace names
		// 	1.1 filter out namespaces with given include/exclude options
		// 2. Get all Deployment in that namespaces
		// 3. Get all referenced Secrets/ConfigMaps in each Deployment
		// 4. Compare updated time on Deployment and Secrets/ConfigMaps
		// 5. Restart the Deployment if its updated time is older than its referenced Secrets/ConfigMaps
		controller.Do()

		logger.Info("---------------------- [cluster] sleep ----------------------")
		time.Sleep(time.Second * 2)
	}
}
