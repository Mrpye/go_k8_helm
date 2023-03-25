// This package contains helper functions for managing K8s cluster and Helm charts
//
//	Creating and deleting K8 manifests yamls
//	Installing and uninstalling Helm charts
//	Getting the status of K8s services
//	Getting the status of K8s deployments
//	Getting the status of K8s pods
//	Getting the status of K8s services
//	Managing Helm releases
//	Managing Helm repositories
package go_k8_helm

import (
	"flag"
	"fmt"
	"path"
	"strings"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// buildRestConfig builds the rest config
// Reads the kube config file and returns the rest config
// returns error if there is an issue
func (m *K8) buildRestConfig() (*rest.Config, error) {
	var kube_config string

	//**********************************
	//Shall we use the token connection?
	//**********************************
	if m.UseTokenConnection {
		if m.Host == "" || m.Authorization == "" {
			return nil, fmt.Errorf("host and authorization are required for token connection")
		}
		config := &rest.Config{
			Host:            m.Host,
			BearerToken:     m.Authorization,
			TLSClientConfig: rest.TLSClientConfig{Insecure: true},
		}
		return config, nil
	}
	//**************************************************
	//Use the kube config file to connect to the cluster
	//**************************************************
	if m.ConfigPath == "" {
		home := homedir.HomeDir()
		kube_config = path.Join(home, ".kube", "config") // flag.String("kubeconfig", path.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		if strings.HasSuffix(m.ConfigPath, "/") {
			kube_config = path.Join(m.ConfigPath, "config") //flag.String("kubeconfig", m.ConfigPath, "absolute path to the kubeconfig file")
		} else {
			kube_config = m.ConfigPath
		}
	}
	flag.Parse()

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.ExplicitPath = kube_config
	// if you want to change the loading rules (which files in which order), you can do so here
	var configOverrides clientcmd.ConfigOverrides

	if m.DefaultContext != "" {
		configOverrides = clientcmd.ConfigOverrides{
			CurrentContext: m.DefaultContext,
		}
	} else {
		configOverrides = clientcmd.ConfigOverrides{}
	}
	// if you want to change override values or bind them to flags, there are methods to help you
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &configOverrides)
	config, err := kubeConfig.ClientConfig()

	if err != nil {
		return nil, fmt.Errorf("unable to load kube config %s with context %s", kube_config, m.DefaultContext)

	}
	return config, nil
}
