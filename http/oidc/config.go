// Copyright © 2020–2026 by PACE Mobility GmbH. All rights reserved.

package oidc

// Config for OIDC based on swagger
type Config struct {
	Description      string
	OpenIdConnectURL string `json:"openIdConnectUrl"`
}
