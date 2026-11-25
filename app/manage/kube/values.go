/*
Package kube
Copyright (c) 2024 Erich Gatejen
[LICENSE] https://www.apache.org/licenses/LICENSE-2.0.txt

Manage kubernetes deployments.
*/
package kube

import "gitlab.com/ginfra/ginfra/base"

// #####################################################################################################################
// #

const (
	TimeoutWaitingForDeploymenMS = 5000
	PrimaryPortDefault           = 8900
	PrimaryPortName              = "primary"
	KubernetesNamespace          = "k8s.io"

	EndpointReadRetries = 5
	EndpointReadDelay   = 200

	// -- Naming --

	InitContainerName  = "initcontainer"
	InitContainerImage = "ginfra/initcontainer" // Make this configurable.
	InitMountPoint     = "/initconfig"
	ConfigVolume       = "config-volume"
	ConfigMountPoint   = base.ConfigMountDir
)

func kubeServiceConfigMapName(name string) string {
	return "config" + name
}

func kubeServiceConfigMapBlobName(name string) string {
	return "configblob" + name
}
