# go_k8_helm

## Description
A package of useful functions for managing k8 cluster and helm charts

---
## When to use go_k8_helm
- when you want to install delete k8 manifest files
- get information on services and deployments ect
- deploy / delete /upgrade helm charts
- Add helm repo

---

## Requirements
* go 1.8 [https://go.dev/doc/install](https://go.dev/doc/install) to run and install go_k8_helm

---

## Installation and Basic usage
This will take you through the steps to install and get go_k8_helm up and running.
<details>
<summary>1. Install</summary>

``` bash
go get github.com/Mrpye/go_k8_helm
```
</details>

<details>
<summary>2. Add to your project</summary>


```go
    import "github.com/Mrpye/go_k8_helm"
```
</details>

---

## Examples

<details>
<summary>1. Get workspace items pods deployments ...</summary>

```go
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
	result, err := k8.GetServiceIP("default", "my_service")
	if err != nil {
		panic(err)
	}

	for _, s := range result {
		println(s.ServiceName, s.IP, s.Port, s.ServiceType)
	}

}

```

</details>

<details>
<summary>2. Get service IP</summary>

```go
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
	result, err := k8.GetServiceIP("default", "my_service")
	if err != nil {
		panic(err)
	}

	for _, s := range result {
		println(s.ServiceName, s.IP, s.Port, s.ServiceType)
	}

}


```

</details>


<details>
<summary>3. Add helm repo and install/upgrade uninstall helm chart</summary>

```go
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


```

</details>

---

## To Do
- Write tests


--- 

## Notable 3rd party Libraries

- [https://github.com/helm/helm](https://github.com/helm/helm)
- [https://github.com/kubernetes/client-go](https://github.com/kubernetes/client-go)



## license
go_k8_helm is Apache 2.0 licensed.

## Change Log
### v0.2.0
- Added examples
- Update documents
- Document functions
- Added Delete actions for pods, ns, deployments, service, DemonSet, PV, PVC
- changed way DeleteNS works

