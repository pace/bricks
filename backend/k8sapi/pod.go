// Copyright © 2022 by PACE Telematics GmbH. All rights reserved.
// Created at 2022/01/24 by Vincent Landgraf

package k8sapi

import (
	"context"
	"fmt"
	"net/http"
)

// SetCurrentPodLabel set the label for the current pod in the current
// namespace (requires patch on pods resource)
func (c *Client) SetCurrentPodLabel(ctx context.Context, label, value string) error {
	return c.SetPodLabel(ctx, c.Namespace, c.Podname, label, value)
}

// SetPodLabel sets the label and value for the pod of the given namespace
// (requires patch on pods resource in the given namespace)
func (c *Client) SetPodLabel(ctx context.Context, namespace, podname, label, value string) error {
	pr := []struct {
		Op    string `json:"op"`
		Path  string `json:"path"`
		Value string `json:"value"`
	}{
		{
			Op:    "add",
			Path:  "/metadata/labels/" + label,
			Value: value,
		},
	}
	url := fmt.Sprintf("https://%s:%d/api/v1/namespaces/%s/pods/%s",
		c.cfg.Host, c.cfg.Port, namespace, podname)
	var resp interface{}

	return c.SimpleRequest(ctx, http.MethodPatch, url, pr, resp)
}
