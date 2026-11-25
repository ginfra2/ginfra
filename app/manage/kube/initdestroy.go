/*
Package kube
Copyright (c) 2024 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Manage kubernetes hosted deployments.
*/
package kube

import (
	"context"
	"encoding/json"
	"fmt"
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/app/manage"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common"
	"gitlab.com/ginfra/ginfra/common/config"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strings"

	//"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"k8s.io/client-go/kubernetes"

	"os"
)

const (
	ConfigServiceName     = "configservice"
	ConfigContainerName   = "giconfig"
	ConfigPortName        = "sc-pri-port"
	ConfigServiceCmapBlob = "configservice-blob"
	ConfigServiceVolume   = "configservice-volume"
)

// #####################################################################################################################
// #

// KubeHostInit initializes a kubernetes deployment.  The directory it is initialized in will be for this deployment only.
//
// Parameters:
// - gctx: The ginfra app context.
// - configImage: The Docker image for the configuration service container.
//
// Returns:
// - err: An error, if any.
func KubeHostInit(gctx *app.GContext, configImage string) (err error) {

	var (
		kube = manage.HostFile{
			DeploymentName:    gctx.Deployment,
			ConfigAccessToken: config.GetNewAuthToken(false),
			InitConfigMapName: gctx.Deployment + "-scgiconfig",
		}

		cmi   *apiv1.ConfigMap
		cdepi *appsv1.Deployment
		svci  *apiv1.Service
	)

	if _, err = os.Stat(manage.GetHostFilePath(gctx)); err == nil {
		return base.NewGinfraErrorA("Kube host deployment metadata file already exists.  Is there another deployment?",
			base.LM_FILEPATH, app.PathHostingFilePath)
	}

	dctx := context.Background()
	var scfg = make(map[string]interface{})
	scfg[base.ConfigAuth] = base.GenerateCodeword(base.ConfigAuthLength)
	scfg[base.ConfigUri] = "config://"
	scfg[base.ConfigPortalUri] = "config://"
	scfg[base.ConfigAccessToken] = kube.ConfigAccessToken
	scfg[base.ConfigType] = config.CN_INSTANCE_TYPE__KUBE
	d, _ := json.Marshal(scfg)

	cmap := apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: kube.InitConfigMapName,
		},
		Data: map[string]string{
			ConfigServiceCmapBlob: string(d),
		},
	}

	rep := int32(1)
	cdep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ConfigServiceName,
			Namespace: kube.DeploymentName,
			Labels: map[string]string{
				base.LM_SERVICE_NAME: ConfigServiceName,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &rep,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					base.LM_SERVICE_NAME: ConfigServiceName,
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						base.LM_SERVICE_NAME: ConfigServiceName,
					},
				},
				Spec: apiv1.PodSpec{
					Volumes: []apiv1.Volume{
						{
							Name: ConfigServiceVolume,
							VolumeSource: apiv1.VolumeSource{
								ConfigMap: &apiv1.ConfigMapVolumeSource{
									LocalObjectReference: apiv1.LocalObjectReference{
										Name: kube.InitConfigMapName,
									},
									Items: []apiv1.KeyToPath{
										{
											Key:  ConfigServiceCmapBlob,
											Path: base.ConfigFileName,
										},
									},
								},
							},
						},
						{
							Name: ConfigVolume,
						},
					},
					InitContainers: []apiv1.Container{
						{
							Name:            InitContainerName,
							Image:           InitContainerImage,
							ImagePullPolicy: apiv1.PullIfNotPresent,
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      ConfigServiceVolume,
									MountPath: InitMountPoint,
								},
								{
									Name:      ConfigVolume,
									MountPath: ConfigMountPoint,
								},
							},
						},
					},
					Containers: []apiv1.Container{
						{
							Name:            ConfigContainerName,
							Image:           configImage,
							ImagePullPolicy: apiv1.PullIfNotPresent,
							Ports: []apiv1.ContainerPort{
								{
									Name:          ConfigPortName,
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: config.RESERVED_CONFIG_SERVICE_PORT,
								},
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      ConfigVolume,
									MountPath: ConfigMountPoint,
								},
							},
						},
					},
				},
			},
		},
	}

	svc := apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: ConfigServiceName,
		},
		Spec: apiv1.ServiceSpec{
			Selector: map[string]string{
				base.LM_SERVICE_NAME: ConfigServiceName,
			},
			Ports: []apiv1.ServicePort{
				{
					Name:       ConfigPortName,
					Protocol:   apiv1.ProtocolTCP,
					Port:       config.RESERVED_CONFIG_SERVICE_PORT,
					TargetPort: intstr.IntOrString{Type: intstr.String, StrVal: ConfigPortName},
				},
			},
		},
	}

	// =====  CLIENT AND CONFIG ========================================================
	client, err := getKubeClient()
	if err != nil {
		return base.NewGinfraErrorChild("Could not get kubernetes client.", err)
	}

	// Delete previous one if it exists.  Ignore errors.
	_ = client.CoreV1().ConfigMaps(kube.DeploymentName).Delete(dctx, kube.InitConfigMapName, metav1.DeleteOptions{})

	cmi, err = client.CoreV1().ConfigMaps(kube.DeploymentName).Create(dctx, &cmap, metav1.CreateOptions{})
	if err != nil {
		return base.NewGinfraErrorChild("Could not create configuration map.", err)
	}

	// Maybe do rollback from this point
	defer func() {
		if r := recover(); r != nil {
			if perr, ok := r.(*base.GinfraError); ok { // TODO might be dangerous if another object comes out.
				err = perr
			} else {
				err = base.NewGinfraErrorA("Failed to init host.", base.LM_CAUSE, fmt.Sprintf("%v", r))
			}
			_ = destroyServiceClients(gctx, client, kube.DeploymentName, cmi, cdepi, svci)
		}
	}()

	// =====  CREATE DEPLOYMENT ==========================================================

	cdepi, err = client.AppsV1().Deployments(kube.DeploymentName).Create(dctx, cdep, metav1.CreateOptions{})
	if err != nil {
		panic(base.NewGinfraErrorChild("Could not deploy.", err))
	}

	cdepi, err = whenDeploymentReady(client, cdepi)
	if err != nil {
		panic(base.NewGinfraErrorChild("Init deployment did not start.", err))
	}

	// =====  CREATE SERVICE ==========================================================

	svci, err = client.CoreV1().Services(kube.DeploymentName).Create(dctx, &svc, metav1.CreateOptions{})
	if err != nil {
		panic(base.NewGinfraErrorChild("Could not create Init service.", err))
	}

	el, err := getEndpoints(client, kube.DeploymentName, svci.Name)
	if err != nil {
		panic(err)
	}

	// Just grab the first one for now.
	kube.ConfigUrl = fmt.Sprintf("http://%s:%d", el.Subsets[0].Addresses[0].IP,
		config.RESERVED_CONFIG_SERVICE_PORT)

	// =====  HOSTING CONFIG FILE ========================================================
	if gctx.PortalHostnameUrl == "" {
		kube.ConfigPortalHostUrl = kube.ConfigUrl
	} else {
		if !strings.HasPrefix(gctx.PortalHostnameUrl, "http:") {
			panic("Bad --portal url")
		}
		kube.ConfigPortalHostUrl = gctx.PortalHostnameUrl
	}

	if err == nil {
		// Write out file.
		if err = common.WriteJsonFile(kube, manage.GetHostFilePath(gctx),
			"hosting metadata"); err != nil {
			panic(base.NewGinfraErrorChildA("Could not write kube hosting metadata file.", err, base.LM_FILEPATH,
				app.PathHostingFilePath))
		}
		gctx.GetLogger().Info("Hosted deployment initialized.  Configuration service alive.",
			base.LM_SERVICE_PORT, config.RESERVED_CONFIG_SERVICE_PORT)
	}

	return err
}

func KubeHostDestroy(gctx *app.GContext) (err error) {

	kube, cprv, err := prepKube(gctx)
	if err != nil {
		return err
	}
	client, err := getKubeClient()
	if err != nil {
		return base.NewGinfraErrorChild("Could not get kubernetes client.", err)
	}

	// Destroy services.
	svc := cprv.FetchConfig(config.CR_DEPLOYMENTS, kube.DeploymentName, config.CRD_SERVICES)
	if svc.Err == nil {
		for k, v := range svc.Config {
			if _, ok := v.(map[string]interface{}); ok {
				// It has the deplopment for the service in question.  Zap it.
				if err = destroyService(gctx, client, kube.DeploymentName, kubeServiceConfigMapName(k), k); err != nil {
					gctx.GetLogger().Errorw("Failed to remove a service for the deployment.",
						zap.String(base.LM_DEPLOYMENT_NAME, kube.DeploymentName),
						zap.String(base.LM_SERVICE_NAME, k),
						zap.String(base.LM_CAUSE, err.Error()))
				}
			} else {
				panic("Corrupt configuration.  This should never happen.")
			}
		}
		gctx.GetLogger().Info("Done stopping services.")
	}

	// Destroy configuration provider.  Ignore the error, it will already be logged.
	_ = destroyService(gctx, client, kube.DeploymentName, kube.InitConfigMapName, ConfigServiceName)

	// Delete host spec.
	err = os.Remove(manage.GetHostFilePath(gctx))

	if err == nil {
		gctx.GetLogger().Info("Destroy complete.")
	}

	return err
}
