// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.

package oidc

// Config for OIDC based on swagger
type Config struct {
	Description      string
	OpenIdConnectURL string `json:"openIdConnectUrl"`
}
