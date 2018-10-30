// Copyright © 2018 by PACE Telematics GmbH. All rights reserved.
// Created at 2018/10/30 by Vincent Landgraf

package jarvisv1

import (
	"lab.jamit.de/pace/go-microservice/client/pace/client"
)

// Client implements a Jarvis v1 API client
type Client struct {
	*client.Client
}

// New creates new client with the passed endpoint
func New(endpoint string) *Client {
	return &Client{Client: client.New(endpoint)}
}

// EndpointDevelopment URL for the development environment
const EndpointDevelopment = "https://j-1-dev.pacelink.net"

// EndpointStaging URL for the staging environment
const EndpointStaging = "https://j-1-stage.pacelink.net"

// EndpointProduction URL for the production environment
const EndpointProduction = "https://j-1-prod.pacelink.net"
