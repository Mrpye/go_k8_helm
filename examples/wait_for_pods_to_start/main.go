package main

import "github.com/Mrpye/go_k8_helm"

func main() {

	//*****************************************************************************
	// Create the k8 connection
	//you can pass in the kube config file path or leave black for default location
	//*****************************************************************************
	k8, err := go_k8_helm.CreateK8KubeConfig("minikube", "")

	// ********************************
	//Can also use the token connection
	// ********************************
	//k8, err := go_k8_helm.CreateK8Token("localhost","auth",true)

	if err != nil {
		panic(err)
	}
	//****************************
	//get items from the workspace
	//****************************
	running, result, err := k8.CheckStatusOf("default", []interface{}{"deployment:nginx(.*)"}, false)
	if err != nil {
		panic(err)
	}

	if running {
		println("All pods are running")
	}

	for _, s := range result {
		println(s)
	}

}
