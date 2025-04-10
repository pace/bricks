// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.

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

	sort.Stable(&list)

	actual := make([]string, len(paths))
	for i, route := range list {
		actual[i] = route.pattern
	}

	assert.Equal(t, []string{
		"/beta/payment-method-kinds/applepay/authorize",
		"/beta/payment-methods/{paymentMethodId}/notification",
		"/beta/payment-methods/confirm/{token}",
		"/beta/payment-methods/{paymentMethodId}/authorize",
		"/beta/transactions/{transactionId}/cancel",
		"/beta/receipts/{transactionID}.{fileFormat}",
		"/beta/payment-methods/paypal",
		"/beta/payment-methods/dkv",
		"/beta/payment-methods/paydirekt",
		"/beta/payment-methods/creditcard",
		"/beta/payment-methods/hoyer",
		"/beta/payment-methods/sepa-direct-debit",
		"/beta/receipts/{transactionID}",
		"/beta/transactions/{transactionId}",
		"/beta/payment-tokens/{paymentTokenId}",
		"/beta/payment-methods/{paymentMethodId}",
		"/beta/payment-methods?filter[status]=valid",
		"/beta/payment-method-kinds",
		"/beta/payment-methods",
		"/beta/payment-methods",
		"/beta/transactions",
	}, actual)
}
