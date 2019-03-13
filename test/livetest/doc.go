// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/02/01 by Vincent Landgraf

// Package livetest implements a set of helpers that ease writing of a
// sidecar that tests the functions of a service.
//
// Assuring the functional correctness of a service is an important
// task for a production grade system. This package aims to provide
// helpers that allow a go test like experience for building functional
// health tests in production.
//
// Test functions need to be written similarly to the regular go test
// function format. Only difference is the use of the testing.TB
// interface.
//
// If a test failed, all other tests are still executed. So tests
// should not build on each other. Sub tests should be used for that
// purpose.
//
// The result for the tests is exposed via prometheus metrics.
//
// The interval is configured using PACE_LIVETEST_INTERVAL (duration format).
package livetest
