/*
Package kube
Copyright (c) 2024 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Manage kubernetes hosted deployments.
*/
package kube

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/avast/retry-go/v4"
	"gitlab.com/ginfra/ginfra/app"
	"gitlab.com/ginfra/ginfra/app/manage"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common"
	"gitlab.com/ginfra/ginfra/common/config"
	"gitlab.com/ginfra/ginfra/common/service/shared"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// #####################################################################################################################
// #

func kubeMangleName(pname string) string {
	var b bytes.Buffer
	for c := range pname {
		if (c >= 48 && c <= 57) || (c >= 65 && c <= 90) || (c >= 97 && c <= 122) || c == 45 || c == 46 {
			b.Write([]byte{byte(c)}) // It will always be under 256
		}
	}
	s := b.String()
	if len(s) > 12 {
		s = s[:10]
	}
	return strings.ToLower(s + base.GenerateCodeword(6))
}

func KubeStart(gctx *app.GContext, name, containerImage string, additional []string) (err error) {

	var (
		kube   manage.HostFile
		cprv   config.ConfigProvider
		cpauth string
		rcp    string
		cspc   *shared.Specifics      // Container specifics.
		sspc   map[string]interface{} // Service specifics.
		cmi    *apiv1.ConfigMap
		cdepi  *appsv1.Deployment
		svci   *apiv1.Service
		sha    string
	)

	// -- Get deployment info and config provider.
	if kube, cprv, err = prepKube(gctx); err != nil {
		return err
	}

	if gctx.Port <= 0 {
		gctx.Port = PrimaryPortDefault
	}

	// -- Container specifics. -----------------------------------------------------------------------------------
	if sspc, err = manage.GetServiceSpecifics(gctx, additional); err != nil {
		return base.NewGinfraErrorChild("Could not process specifics for service.", err)
	}

	if cspc, sha, err = getContainerSpecifics(gctx, containerImage, sspc); err != nil {
		// It's ok to not have specifics.
		if ((base.GetErrorFlags(err) & base.ERRFLAG_NOT_FOUND) > 0) ||
			((base.GetErrorFlags(err) & base.ERRFLAG_RECOVERABLE) > 0) {
			gctx.GetLogger().Warn("Bad specifics for container.  Ignoring.", zap.String(base.LM_CID, name), zap.Error(err))
		} else {
			return err
		}
	}

	if cspc, err = manage.PrepSpecifics(gctx, cspc, sspc); err != nil {
		return err
	}

	// -- Deal with security specifics -------------------------------------------------------------------------------

	if gctx.Bundle != "" {
		if ok, cert, key, errr := shared.BundleExtractServerTLS(base.GetGinfraHome()); ok {
			sspc[config.CRD_SSERVICE__CONFIG__SPECIFICS__SRVCERT] = cert
			sspc[config.CRD_SSERVICE__CONFIG__SPECIFICS__SRVKEY] = key
		} else if errr != nil {
			gctx.GetLogger().Warnw("Bundle configured but could not be extracted.", zap.Error(errr))
		}
	}

	// -- Register service. --------------------------------------------------------------------------------------

	// Service class
	sclass := gctx.ServiceClass
	if sclass == "" {
		sclass = name
	}

	auth := gctx.Authtoken
	if auth == "" {
		auth = base.GenerateCodeword(base.ConfigAuthLength)
		gctx.GetLogger().Info("Using generated auth token.")
	} else {
		gctx.GetLogger().Info("Using provided auth token.")
	}

	if cpauth, err = cprv.RegisterService(config.ITKube, name, kube.DeploymentName, auth, containerImage, rcp,
		sclass, gctx.Port, sspc); err != nil {
		return base.NewGinfraErrorChild("Could not register service", err)
	}

	// Early register instance.
	if _, err = cprv.RegisterInstance(config.ITKube, name, kube.DeploymentName, cpauth, name,
		"0.0.0.0", "0.0.0.0"); err != nil {
		manage.RollbackRegistration(gctx, kube.DeploymentName, name, cpauth, cprv)
		return base.NewGinfraErrorChild("Could not register service instance.", err)
	}

	// =====  CONFIG ===================================================================================
	icont := gctx.InitContainer
	if icont == "" {
		icont = InitContainerImage
	}

	dctx := context.Background()
	var cfg = make(map[string]interface{})
	cfg[base.ConfigAuth] = auth
	cfg[base.ConfigUri] = kube.ConfigUrl
	cfg[base.ConfigPortalUri] = kube.ConfigPortalHostUrl
	cfg[base.ConfigAccessToken] = cpauth
	cfg[base.ConfigDeployment] = gctx.Deployment
	cfg[base.ConfigName] = name
	cfg[base.ConfigId] = name
	cfg[base.ConfigPort] = gctx.Port
	cfg[base.ConfigType] = config.CN_INSTANCE_TYPE__KUBE
	if gctx.Clear {
		cfg[base.ConfigClear] = "true"
	}
	cfg[base.ConfigSpecifics] = sspc

	d, _ := json.Marshal(cfg)
	cmap := apiv1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: kubeServiceConfigMapName(name),
		},
		Data: map[string]string{
			kubeServiceConfigMapBlobName(name): string(d),
		},
	}

	ppn := strings.ToLower(PrimaryPortName + base.GenerateCodeword(6))
	rep := int32(1)
	cdep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: kube.DeploymentName,
			Labels: map[string]string{
				base.LM_SERVICE_NAME: name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &rep,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					base.LM_SERVICE_NAME: name,
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						base.LM_SERVICE_NAME: name,
					},
				},
				Spec: apiv1.PodSpec{
					Volumes: []apiv1.Volume{
						{
							Name: name,
							VolumeSource: apiv1.VolumeSource{
								ConfigMap: &apiv1.ConfigMapVolumeSource{
									LocalObjectReference: apiv1.LocalObjectReference{
										Name: kubeServiceConfigMapName(name),
									},
									Items: []apiv1.KeyToPath{
										{
											Key:  kubeServiceConfigMapBlobName(name),
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
							Image:           icont,
							ImagePullPolicy: apiv1.PullIfNotPresent,
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name:      name,
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
							Name:            name,
							Image:           containerImage + "@" + sha,
							ImagePullPolicy: apiv1.PullIfNotPresent,
							Ports: []apiv1.ContainerPort{
								{
									Name:          ppn,
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: int32(gctx.Port),
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
			Name: name,
		},
		Spec: apiv1.ServiceSpec{
			Selector: map[string]string{
				base.LM_SERVICE_NAME: name,
			},
			Ports: []apiv1.ServicePort{
				{
					Name:       ppn,
					Protocol:   apiv1.ProtocolTCP,
					Port:       int32(gctx.Port),
					TargetPort: intstr.IntOrString{Type: intstr.String, StrVal: ppn},
				},
			},
		},
	}

	if cspc != nil {
		for pname, v := range cspc.Exposed {

			// Scrub the name against the server specifics.
			nn, _ := common.SimpleReplacer(pname, sspc, true)
			pp, err := strconv.Atoi(nn)
			if err != nil {
				return base.NewGinfraErrorChild("Bad container specifics value for export port number", err)
			}
			mn := kubeMangleName(v.Name)

			// Container port
			cp := apiv1.ContainerPort{
				Name:          mn,
				Protocol:      apiv1.ProtocolTCP,
				ContainerPort: int32(pp),
			}
			cdep.Spec.Template.Spec.Containers[0].Ports = append(cdep.Spec.Template.Spec.Containers[0].Ports, cp)

			// Service port
			sp := apiv1.ServicePort{
				Name:       mn,
				Protocol:   apiv1.ProtocolTCP,
				Port:       int32(pp),
				TargetPort: intstr.IntOrString{Type: intstr.String, StrVal: mn},
			}
			svc.Spec.Ports = append(svc.Spec.Ports, sp)
		}
	}

	// ===== CONFIG =====================================================================================
	// From here rollbacks become necessary.

	client, err := getKubeClient()
	if err != nil {
		return base.NewGinfraErrorChild("Could not get kubernetes client.", err)
	}

	// Delete previous one if it exists.  Ignore errors.
	_ = client.CoreV1().ConfigMaps(kube.DeploymentName).Delete(dctx, cmap.Name, metav1.DeleteOptions{})

	cmi, err = client.CoreV1().ConfigMaps(kube.DeploymentName).Create(dctx, &cmap, metav1.CreateOptions{})
	if err != nil {
		return base.NewGinfraErrorChild("Could not create configuration map.", err)
	}

	// Make sure we rollback even if there is a panic from something else.
	defer func() {
		if r := recover(); r != nil {
			switch val := r.(type) {
			case *base.GinfraError:
				err = val
			case retry.Error:
				if len(val) > 0 {
					err = base.NewGinfraErrorA("Failed to start service.", base.LM_CAUSE, fmt.Sprintf("%v", val[0]))
				} else {
					err = base.NewGinfraError("Failed to start service due to unknown reason in retry.  This is probably a bug.")
				}
			default:
				err = base.NewGinfraErrorA("Failed to start service.", base.LM_CAUSE, fmt.Sprintf("%v", r))
			}

			manage.RollbackRegistration(gctx, kube.DeploymentName, name, cpauth, cprv)
			if !gctx.NoRollback {
				// Need to ask for the deployment again.
				cdepi, _ = client.AppsV1().Deployments(kube.DeploymentName).Get(dctx, cdep.Name, metav1.GetOptions{})
				_ = destroyServiceClients(gctx, client, name, cmi, cdepi, svci) // Failures are already logged.
			}

			// TODO reap services no longer needed.
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

	// =====  CREATE SERVICE ===================================================================

	svci, err = client.CoreV1().Services(kube.DeploymentName).Create(dctx, &svc, metav1.CreateOptions{})
	if err != nil {
		panic(base.NewGinfraErrorChild("Could not create Init service.", err))
	}

	el, err := getEndpoints(client, kube.DeploymentName, svci.Name)
	if err != nil {
		panic(err)
	}

	// Just grab the first one for now.
	//ip := el.Subsets[0].Addresses[0].IP
	addy := el.Subsets[0].Addresses[0].IP // Just IPs for now

	// =====  DISCOVERY =========================================================================

	// Add discovery and portal for primary port.  This is messy unless there is a loadbalancer or dispatcher.
	// This will get cleaned up when we address that.
	err = cprv.DiscoveryAdd(gctx.Deployment, sclass, fmt.Sprintf("http://%s:%d/", addy, gctx.Port))
	if err != nil {
		gctx.GetLogger().Warn("Could not add discovery for primary port.", zap.Error(err))
	}

	// Portals should be added with ingress if not the default.
	//_, err = cprv.PortalAdd(kube.DeploymentName, sclass, addy)
	//if err != nil {
	//	panic(err)
	//}

	// Add discovery for exports.
	if cspc != nil {
		for nport, e := range cspc.Exposed {

			if e.ServiceClass.Name != "" {
				if nport != "" && e.ServiceClass.Url != "" {
					nurl, err := common.SimpleReplacer(e.ServiceClass.Url,
						map[string]interface{}{"host": addy, "port": nport}, false)
					if err != nil {
						gctx.GetLogger().Panic("Container specifics are corrupt.", zap.Error(err))
					}

					err = cprv.DiscoveryAdd(gctx.Deployment, name+"/"+e.ServiceClass.Name, nurl)
					if err != nil {
						gctx.GetLogger().Warn("Could not add discovery for exposed port.", zap.Error(err))
					}

				}
				_, err = cprv.PortalAdd(kube.DeploymentName, e.ServiceClass.Name, addy)
				if err != nil {
					panic(err)
				}
			}
		}
	}

	gctx.GetLogger().Info("Hosted service started.", base.LM_SERVICE_IP, kube.ConfigUrl, base.LM_SERVICE_PORT,
		gctx.Port)
	return nil
}

func KubeStop(gctx *app.GContext, name string) (err error) {

	kube, cprv, err := prepKube(gctx)
	if err != nil {
		return err
	}
	client, err := getKubeClient()
	if err != nil {
		return base.NewGinfraErrorChild("Could not get kubernetes client.", err)
	}

	// These will log their own errors.
	manage.RollbackRegistration(gctx, kube.DeploymentName, name, "x", cprv)
	_ = destroyService(gctx, client, kube.DeploymentName, kube.InitConfigMapName, ConfigServiceName)
	return nil
}
