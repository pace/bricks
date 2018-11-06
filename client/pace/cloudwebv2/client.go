// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/10/30 by Vincent Landgraf

package cloudwebv2

import "lab.jamit.de/pace/go-microservice/client/pace/client"

// Client implements a Cloud Web v2 API client
type Client struct {
	*client.Client
}

// New creates new client with the passed endpoint
func New(endpoint string) *Client {
	return &Client{Client: client.New(endpoint)}
}

// EndpointDevelopment URL for the development environment
const EndpointDevelopment = "https://cl-1-dev.pacelink.net"

// EndpointStaging URL for the staging environment
const EndpointStaging = "https://cl-1-stage.pacelink.net"

// EndpointProduction URL for the production environment
const EndpointProduction = "https://cl-1-prod.pacelink.net"
