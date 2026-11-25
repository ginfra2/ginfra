/*
Package providers
Copyright (c) 2023 Erich Gatejen
[LICENSE]

Providers setup.

[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt
*/
package providers

import (
	"<<<Module>>>/local"
	// # YOUR IMPORTS START HERE
	// >A###############################################################################################################
	// >A###############################################################################################################
	// #  YOUR IMPORTS END
)

type ProvidersLoaded struct {

	// # ADD POINTERS (without * because they are interfaces) TO YOUR PROVIDERS START HERE
	// >B###############################################################################################################

	// >B###############################################################################################################
	// # ADD POINTERS (without * because they are interfaces) TO YOUR PROVIDERS END HERE

}

func ProvidersLoad(gcontext *local.GContext) (*ProvidersLoaded, error) {
	var (
		l   ProvidersLoaded
		err error
	)

	// # ADD INSTANTIATING AND REMEMBERING YOUR PROVIDERS START HERE
	// >C###############################################################################################################

	// >C###############################################################################################################
	// # ADD INSTANTIATING AND REMEMBERING YOUR PROVIDERS END HERE

	return &l, err
}
