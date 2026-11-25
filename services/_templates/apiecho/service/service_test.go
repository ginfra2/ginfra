//go:build functional_test

/*
Package service
// >A#

You may license this file however you wish.

// >A#
*/
package service

import (
	"testing"
	// # YOUR IMPORTS START HERE
	// >B###############################################################################################################
	// >B###############################################################################################################
	// #  YOUR IMPORTS END
)

func yourSetup(t *testing.T, context *TestServiceContext) error {

	// # YOUR SETUP STARTS HERE
	// >C###############################################################################################################

	return nil

	// >C###############################################################################################################
	// #  YOUR SETUP ENDS HERE
}

func yourTeardown(t *testing.T, context *TestServiceContext) error {

	// # YOUR TEARDOWN STARTS HERE
	// >D###############################################################################################################

	return nil

	// >D###############################################################################################################
	// #  YOUR TEARDOWN ENDS HERE
}

/*
	IMPORTANT: Add the following two lines to the start of any Test* function.
		teardown, tcontext := setupTest(t, yourSetup)
		defer teardown(t, tcontext, yourTeardown)
*/
// ^^###################################################################################################################
// # ADDITIONAL CODE GOES BELOW HERE
