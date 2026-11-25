/*
Package config
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Various configuration types and points.
*/
package config

import (
	"gitlab.com/ginfra/ginfra/base"
)

// =============================================================================================
// = INSTANCE TYPE

type InstanceType int

const (
	ITBad InstanceType = 0

	// ---- Both 'Host types' and 'Specific instance' types. --------------------------------

	// ITLocal indicates a local processes running on a local machine.
	ITLocal InstanceType = 1

	// ITHost indicates docker containers running in a docker host on a local machine.
	ITHost InstanceType = 2

	// ITManaged indicates docker containers running in a managed service.
	ITManaged InstanceType = 3

	// ITKube indicates containers running in a kubernetes cluster.
	ITKube InstanceType = 4

	// ---- Host types ----------------------------------------------------------------------
	// reserved

	// ---- Specific instance types ---------------------------------------------------------
	// reserved

)

var (
	instanceTypeMap = map[string]InstanceType{
		CN_INSTANCE_TYPE__LOCAL:   ITLocal,
		CN_INSTANCE_TYPE__HOST:    ITHost,
		CN_INSTANCE_TYPE__MANAGED: ITManaged,
		CN_INSTANCE_TYPE__KUBE:    ITKube,
	}

	instanceTypeArray = []string{
		"bad",
		CN_INSTANCE_TYPE__LOCAL,
		CN_INSTANCE_TYPE__HOST,
		CN_INSTANCE_TYPE__MANAGED,
		CN_INSTANCE_TYPE__KUBE,
	}
)

// GetInstanceTypeFromString converts a string representation to the corresponding InstanceType.
// Returns ITBad and an error if the string does not match any known InstanceType.
func GetInstanceTypeFromString(t string) (InstanceType, error) {
	if c, ok := instanceTypeMap[t]; ok {
		return c, nil
	}
	return ITBad, base.NewGinfraErrorA("Bad instance type.", base.LM_TYPE, t)
}

// GetStringFromInstanceType converts an InstanceType to its corresponding string representation or returns an error if invalid.
func GetStringFromInstanceType(itype InstanceType) (string, error) {
	if itype < 0 || int(itype) >= len(instanceTypeArray) {
		return "", base.NewGinfraErrorA("Bad instance type.", base.LM_TYPE, itype)
	}
	return instanceTypeArray[itype], nil
}

// =============================================================================================
// = RESERVATION TYPE

// ReservationType defines an integer-based type representing different kinds of reservations.
type ReservationType int

const (
	RTBad ReservationType = 0

	RTPort ReservationType = 1
)

var (
	reservationTypeMap = map[string]ReservationType{
		CN_RESERVATION_TYPE__PORT: RTPort,
	}

	reservationTypeArray = []string{
		"bad",
		CN_RESERVATION_TYPE__PORT,
	}
)

// GetReservationTypeFromString converts a string to its corresponding ReservationType.
// Returns RTBad and an error if the string is not recognized.
func GetReservationTypeFromString(t string) (ReservationType, error) {
	if c, ok := reservationTypeMap[t]; ok {
		return c, nil
	}
	return RTBad, base.NewGinfraErrorA("Bad instance type.", base.LM_TYPE, t)
}

// GetStringFromReservationType converts a ReservationType to its corresponding string representation.
// Returns an error if the provided ReservationType is invalid or out of range.
func GetStringFromReservationType(itype ReservationType) (string, error) {
	if itype < 0 || int(itype) >= len(reservationTypeArray) {
		return "", base.NewGinfraErrorA("Bad reservation type.", base.LM_TYPE, itype)
	}
	return reservationTypeArray[itype], nil
}
