// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.
// Created at 2020/05/08 by Vincent Landgraf

package oidc

// Config for OIDC based on swagger
type Config struct {
	Description      string
	OpenIdConnectURL string `json:"openIdConnectUrl"`
}
