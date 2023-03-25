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
	//**********
	//Add a repo
	//**********
	err = k8.RepoAdd("my_repo", "https://kubernetes-charts.storage.googleapis.com", "user", "password")
	if err != nil {
		panic(err)
	}
	//***************
	//Update the repo
	//***************
	err = k8.RepoUpdate()
	if err != nil {
		panic(err)
	}

	//***************
	//Deploy a chart
	//***************
	err = k8.DeployHelmChart("my_repo/char", "release_name", "default", map[string]interface{}{"key": "value"})
	if err != nil {
		panic(err)
	}
	//***************
	//Upgrade a chart
	//***************
	err = k8.UpgradeHelmChart("my_repo/char", "release_name", "default", map[string]interface{}{"key": "value"})
	if err != nil {
		panic(err)
	}
	//***************
	//Uninstall a chart
	//***************
	err = k8.UninstallHelmChart("my_repo/char", "release_name")
	if err != nil {
		panic(err)
	}
}
