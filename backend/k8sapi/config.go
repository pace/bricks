// Copyright © 2022 by PACE Telematics GmbH. All rights reserved.

package k8sapi

// Config gathers the required kubernetes system configuration to use the
// kubernetes API
type Config struct {
	Host          string `env:"KUBERNETES_SERVICE_HOST" envDefault:"localhost"`
	Port          int    `env:"KUBERNETES_PORT_443_TCP_PORT" envDefault:"433"`
	NamespaceFile string `env:"KUBERNETES_NAMESPACE_FILE" envDefault:"/run/secrets/kubernetes.io/serviceaccount/namespace"`
	CACertFile    string `env:"KUBERNETES_API_CA_FILE" envDefault:"/run/secrets/kubernetes.io/serviceaccount/ca.crt"`
	TokenFile     string `env:"KUBERNETES_API_TOKEN_FILE" envDefault:"/run/secrets/kubernetes.io/serviceaccount/token"`
}
