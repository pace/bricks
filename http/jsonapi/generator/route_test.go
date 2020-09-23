// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.
// Created at 2020/07/14 by Marius Neugebauer

package generator

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSortableRouteList(t *testing.T) {
	paths := []string{
		"/beta/payment-method-kinds",
		"/beta/payment-methods/paypal",
		"/beta/payment-methods/{paymentMethodId}/notification",
		"/beta/payment-methods?filter[status]=valid",
		"/beta/payment-methods/confirm/{token}",
		"/beta/payment-methods/{paymentMethodId}/authorize",
		"/beta/payment-methods/dkv",
		"/beta/payment-methods/paydirekt",
		"/beta/payment-methods",
		"/beta/payment-methods/creditcard",
		"/beta/receipts/{transactionID}",
		"/beta/payment-method-kinds/applepay/authorize",
		"/beta/transactions/{transactionId}",
		"/beta/payment-tokens/{paymentTokenId}",
		"/beta/payment-methods/hoyer",
		"/beta/payment-methods/{paymentMethodId}",
		"/beta/payment-methods",
		"/beta/transactions",
		"/beta/transactions/{transactionId}/cancel",
		"/beta/payment-methods/sepa-direct-debit",
		"/beta/receipts/{transactionID}.{fileFormat}",
	}
	list := make(sortableRouteList, len(paths))
	for i, path := range paths {
		route := &route{pattern: path}
		require.NoError(t, route.parseURL())
		list[i] = route
	}
	sort.Sort(&list)
	actual := make([]string, len(paths))
	for i, route := range list {
		actual[i] = route.pattern
	}
	assert.Equal(t, []string{
		"/beta/payment-method-kinds/applepay/authorize",
		"/beta/transactions/{transactionId}/cancel",
		"/beta/payment-methods/{paymentMethodId}/notification",
		"/beta/payment-methods/confirm/{token}",
		"/beta/payment-methods/{paymentMethodId}/authorize",
		"/beta/receipts/{transactionID}.{fileFormat}",
		"/beta/payment-methods/hoyer",
		"/beta/payment-methods/paydirekt",
		"/beta/payment-methods/dkv",
		"/beta/payment-methods/creditcard",
		"/beta/payment-methods/sepa-direct-debit",
		"/beta/payment-methods/paypal",
		"/beta/transactions/{transactionId}",
		"/beta/payment-tokens/{paymentTokenId}",
		"/beta/payment-methods/{paymentMethodId}",
		"/beta/receipts/{transactionID}",
		"/beta/payment-methods?filter[status]=valid",
		"/beta/payment-methods",
		"/beta/transactions",
		"/beta/payment-methods",
		"/beta/payment-method-kinds",
	}, actual)
}
