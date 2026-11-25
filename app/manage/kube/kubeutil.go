/*
Package kube
Copyright (c) 2024 Erich Gatejen
[LICENSE] https://www.apache.org/licenses/LICENSE-2.0.txt

Manage kubernetes deployments.  Utilities.
*/
package kube

import (
	"context"
	retry "github.com/avast/retry-go/v4"
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/base"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	"time"
)

// #####################################################################################################################
// #

func getKubeClient() (*kubernetes.Clientset, error) {

	config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
	if err != nil {
		// Leave as panic for now.
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	return clientset, nil
}

func watchDeploymentReady(ctx context.Context, client *kubernetes.Clientset, deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	var (
		err     error
		to      int64 = 25 // This was probably causing the hangups, but I can't tell for sure.
		watcher watch.Interface
		ndep    *appsv1.Deployment
	)

	// Watch deployment until ready
	l := base.LM_SERVICE_NAME + "=" + deployment.Name

	watcher, err = client.AppsV1().Deployments(deployment.Namespace).Watch(context.TODO(), metav1.ListOptions{
		LabelSelector:  l,
		TimeoutSeconds: &to,
	})
	defer watcher.Stop()

	done := false
	for err == nil && !done {
		select {
		case ev, ok := <-watcher.ResultChan():
			if !ok {
				//Hangup.
				return nil, nil
			}
			if ndep, ok = ev.Object.(*appsv1.Deployment); ok {

				// Leaving for some odd debugging.
				left := *ndep.Spec.Replicas
				right := ndep.Status.ReadyReplicas

				if ndep != nil && ndep.Spec.Replicas != nil && left == right {
					done = true
					break
				}
			}
		case <-ctx.Done():
			return ndep, base.NewGinfraError("Timeout: Deployment not ready in time.")
		}
	}

	return ndep, err
}

func whenDeploymentReady(client *kubernetes.Clientset, deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	var (
		dep *appsv1.Deployment
		err error
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*2)
	defer cancel()

	for dep == nil && err == nil {
		dep, err = watchDeploymentReady(ctx, client, deployment)
	}

	return dep, err
}

func destroyService(gctx *app.GContext, client *kubernetes.Clientset, deployment, cmapname, servicename string) error {
	var errored bool

	dctx := context.Background()

	cmi, err := client.CoreV1().ConfigMaps(deployment).Get(dctx, cmapname, metav1.GetOptions{})
	if err != nil || cmi == nil {
		errored = true
		gctx.GetLogger().Errorw("Could not get ConfigMap for service during destroy.  The ConfigMap doesn't exist or may be ghosted",
			zap.Error(err), zap.String(base.LM_DEPLOYMENT_NAME, deployment), zap.String(base.LM_SERVICE_NAME, servicename))
	}

	cdepi, err := client.AppsV1().Deployments(deployment).Get(dctx, servicename, metav1.GetOptions{})
	if err != nil || cdepi == nil {
		errored = true
		gctx.GetLogger().Errorw("Could not get kubernetes Deployment for service during destroy.  The Deployment doesn't exist or may be ghosted",
			zap.Error(err), zap.String(base.LM_DEPLOYMENT_NAME, deployment), zap.String(base.LM_SERVICE_NAME, servicename))
	}

	svci, err := client.CoreV1().Services(deployment).Get(dctx, servicename, metav1.GetOptions{})
	if err != nil || svci == nil {
		errored = true
		gctx.GetLogger().Errorw("Could not get kubernetes Service for service during destroy.  The Service doesn't exist or may be ghosted",
			zap.Error(err), zap.String(base.LM_DEPLOYMENT_NAME, deployment), zap.String(base.LM_SERVICE_NAME, servicename))
	}

	err = destroyServiceClients(gctx, client, deployment, cmi, cdepi, svci)
	if errored || err != nil {
		return base.NewGinfraErrorA("Error while destroying service.", base.LM_SERVICE_NAME, servicename)
	}
	return nil
}

func destroyServiceClients(gctx *app.GContext, client *kubernetes.Clientset, deployment string, cmi *apiv1.ConfigMap,
	cdepi *appsv1.Deployment, svci *apiv1.Service) error {

	var errored bool
	dctx := context.Background()

	if svci != nil {
		err := client.CoreV1().Services(deployment).Delete(dctx, svci.Name, metav1.DeleteOptions{})
		if err != nil {
			errored = true
			gctx.GetLogger().Errorw("Could not get Delete the kubernetes Service during destroy.  It may be ghosted",
				zap.Error(err), zap.String(base.LM_DEPLOYMENT_NAME, deployment), zap.String(base.LM_SERVICE_NAME, svci.Name))
		}
	}
	if cmi != nil {
		err := client.CoreV1().ConfigMaps(deployment).Delete(dctx, cmi.Name, metav1.DeleteOptions{})
		if err != nil {
			errored = true
			gctx.GetLogger().Errorw("Could not get Delete the ConfigMap for service during destroy.  It may be ghosted",
				zap.Error(err), zap.String(base.LM_DEPLOYMENT_NAME, deployment), zap.String(base.LM_SERVICE_NAME, cmi.Name))
		}
	}
	// && cdepi.Status.ObservedGeneration > 0
	if cdepi != nil {
		err := client.AppsV1().Deployments(deployment).Delete(dctx, cdepi.Name, metav1.DeleteOptions{})
		if err != nil {
			errored = true
			gctx.GetLogger().Errorw("Could not get Delete the kubernetes Deployment for service during destroy.  It may be ghosted",
				zap.Error(err), zap.String(base.LM_DEPLOYMENT_NAME, deployment), zap.String(base.LM_SERVICE_NAME, cdepi.Name))
		}
	}

	if errored {
		return base.NewGinfraErrorA("Error while destroying service clients.")
	}
	return nil
}

func getEndpoints(client *kubernetes.Clientset, deployment, name string) (*apiv1.Endpoints, error) {
	var (
		err error
		el  *apiv1.Endpoints
	)

	err = retry.Do(
		func() error {
			el, err = client.CoreV1().Endpoints(deployment).Get(context.Background(), name, metav1.GetOptions{})
			if err != nil {
				return base.NewGinfraErrorChild("Could not get Endpoints for the service.", err)
			}
			if len(el.Subsets) < 1 || len(el.Subsets[0].Addresses) < 1 {
				return base.NewGinfraErrorChild("No Endpoints for the service.", err)
			}
			return nil
		},
		retry.DelayType(func(n uint, err error, config *retry.Config) time.Duration {
			return retry.BackOffDelay(n, err, config)
		}),
		retry.Attempts(EndpointReadRetries),
		retry.Delay(time.Millisecond*EndpointReadDelay),
	)

	return el, err
}
