/*
Package config
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Various configuration types and points.
*/
package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfigTypes(t *testing.T) {

	t.Run("Instance from String", func(t *testing.T) {
		c, _ := GetInstanceTypeFromString(CN_INSTANCE_TYPE__HOST)
		assert.Equal(t, ITHost, c)
	})
	t.Run("Bad Instance from String", func(t *testing.T) {
		c, _ := GetInstanceTypeFromString("bork")
		assert.Equal(t, ITBad, c)
	})
	t.Run("String from Instance", func(t *testing.T) {
		c, _ := GetStringFromInstanceType(ITHost)
		assert.Equal(t, CN_INSTANCE_TYPE__HOST, c)
	})
	t.Run("String from Bad Instance", func(t *testing.T) {
		_, err := GetStringFromInstanceType(-1)
		assert.Error(t, err)
	})
	t.Run("String from Bad Instance 2", func(t *testing.T) {
		_, err := GetStringFromInstanceType(99999999)
		assert.Error(t, err)
	})

	t.Run("Reservation Type from String", func(t *testing.T) {
		c, _ := GetReservationTypeFromString(CN_RESERVATION_TYPE__PORT)
		assert.Equal(t, RTPort, c)
	})
	t.Run("Bad Reservation Type from String", func(t *testing.T) {
		c, _ := GetReservationTypeFromString("bork")
		assert.Equal(t, RTBad, c)
	})
	t.Run("String from Reservation Type", func(t *testing.T) {
		c, _ := GetStringFromReservationType(RTPort)
		assert.Equal(t, CN_RESERVATION_TYPE__PORT, c)
	})
	t.Run("String from Bad Reservation Type", func(t *testing.T) {
		_, err := GetStringFromReservationType(-1)
		assert.Error(t, err)
	})
	t.Run("String from Bad Reservation Type 2", func(t *testing.T) {
		_, err := GetStringFromReservationType(9999999)
		assert.Error(t, err)
	})
}
