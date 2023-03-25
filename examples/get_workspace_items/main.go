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
	result, err := k8.GetPods("default")
	//result, err := k8.GetDemonSet("default")
	//result, err := k8.GetDeployments("default")
	//result, err := k8.GetSecrets("default")
	//result, err := k8.GetStatefulSets("default")
	//result, err := k8.GetServices("default")
	if err != nil {
		panic(err)
	}

	for _, pod := range result.Items {
		println(pod.Name)
	}

}
