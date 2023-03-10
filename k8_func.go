package go_k8_helm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	lib_log "github.com/Mrpye/golib/log"
	"github.com/Mrpye/helm-api/modules/body_types"
	"github.com/gookit/color"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/kubectl/pkg/cmd/cp"
	"k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/scheme"
)

var decUnstructured = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

// PodExec executes a command in a pod
// ns: namespace
// pod_name: pod name
// command: command to execute
// return: stdout, error
func (m *K8) PodExec(ns string, pod_name string, command string) (string, error) {
	//***************
	//Load the Config
	//***************
	//***********************
	//Split the command lines
	//***********************
	quoted := false
	cmd := strings.FieldsFunc(command, func(r rune) bool {
		if r == '"' {
			quoted = !quoted
		}
		return !quoted && r == ' '
	})

	for x := range cmd {
		cmd[x] = strings.ReplaceAll(cmd[x], "\"", "")
	}

	//************
	//Setup the K8
	//************
	config, err := m.buildRestConfig()
	if err != nil {
		return "", err
	}
	//**********************
	// creates the clientset
	//**********************
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", err
	}

	// get pods in all the namespaces by omitting namespace
	// Or specify namespace to get pods in particular namespace
	req := clientset.CoreV1().RESTClient().Post().Resource("pods").Name(pod_name).Namespace(ns).SubResource("exec")
	option := &v1.PodExecOptions{
		Command: cmd,
		Stdin:   true,
		Stdout:  true,
		Stderr:  false,
		TTY:     true,
	}
	req.VersionedParams(
		option,
		scheme.ParameterCodec,
	)
	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return "", err
	}

	//***************
	//Run the command
	//***************

	l := &lib_log.LogStreamer{}

	err = exec.StreamWithContext(context.Background(), remotecommand.StreamOptions{
		Stdin:  os.Stdin,
		Stdout: l,
		Stderr: os.Stderr,
		Tty:    true,
	})

	return string(l.String()), err
}

// PodCopy copies a file to and from a pod
// ns: namespace
// src: source file
// dst: destination file
// container_name: container name
// return: stdout, error
func (m *K8) PodCopy(ns string, src string, dst string, container_name string) (string, error) {
	if ns == "" {
		ns = "default"
	}
	//************
	//Setup the K8
	//************
	config, err := m.buildRestConfig()
	if err != nil {
		return "", err
	}

	config.APIPath = "/api"                                                 // Make sure we target /api and not just /
	config.GroupVersion = &schema.GroupVersion{Group: "api", Version: "v1"} // this targets the core api groups so the url path will be /api/v1
	config.NegotiatedSerializer = serializer.WithoutConversionCodecFactory{CodecFactory: scheme.Codecs}
	//**********************
	// creates the clientset
	//**********************
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", err
	}

	ioStreams, _, out, errOut := genericclioptions.NewTestIOStreams()
	copyOptions := cp.NewCopyOptions(ioStreams)

	var copt genericclioptions.RESTClientGetter = &genericclioptions.ConfigFlags{}

	nf := util.NewFactory(copt)
	cobra := cp.NewCmdCp(nf, ioStreams)
	cobra.ResetFlags()
	copyOptions.Complete(nf, cobra, []string{src, dst})

	copyOptions.Clientset = clientset
	copyOptions.ClientConfig = config
	copyOptions.Container = container_name
	copyOptions.Namespace = ns

	copyOptions.Validate()
	err = copyOptions.Run()

	if err != nil {
		return "", err
	}

	error_str := errOut.String()
	if error_str != "" {
		return "", errors.New(error_str)
	}
	out_str := out.String()

	return out_str, nil
}

// dryRun returns the dry-run value for the given boolean value.
func (m *K8) dryRun(dry_run bool) string {
	if dry_run {
		return metav1.DryRunAll
	}
	return ""
}

// DeleteYaml deletes a resource using a yaml manifest
// ctx: context
// cfg: k8 config
// yaml: yaml manifest
// ns: namespace
// return: error
func (m *K8) DeleteYaml(ctx context.Context, cfg *rest.Config, yaml string, ns string) error {

	// 1. Prepare a RESTMapper to find GVR
	dc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	// 2. Prepare the dynamic client
	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return err
	}

	// 3. Decode YAML manifest into unstructured.Unstructured
	obj := &unstructured.Unstructured{}
	_, gvk, err := decUnstructured.Decode([]byte(yaml), nil, obj)
	if err != nil {
		return err
	}

	if ns != "" {
		obj.SetNamespace(ns)
	}

	// 4. Find GVR
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return err
	}

	//Get the namespace
	namespace := ""
	if obj.GetNamespace() != "" {
		namespace = obj.GetNamespace()
	} else {
		namespace = "default"
	}

	// 5. Obtain REST interface for the GVR
	var dr dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		// namespaced resources should specify the namespace
		dr = dyn.Resource(mapping.Resource).Namespace(namespace)
	} else {
		// for cluster-wide resources
		dr = dyn.Resource(mapping.Resource)
	}

	log.Printf("Info: Deleting Kind(%s) Namespace(%s) Name(%s)\n", obj.GetKind(), obj.GetNamespace(), obj.GetName())

	//********************
	//Lets delete the item
	//********************
	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}
	err = dr.Delete(ctx, obj.GetName(), deleteOptions)
	if err != nil {
		return fmt.Errorf("info: Failed to delete Kind(%s) Namespace(%s) Name(%s) Error(%s)", obj.GetKind(), obj.GetNamespace(), obj.GetName(), err.Error())
	}
	fmt.Printf("Info: Deleted to delete Kind(%s) Namespace(%s) Name(%s)\n", obj.GetKind(), obj.GetNamespace(), obj.GetName())
	return err
}

// ApplyYaml applies a resource using a yaml manifest
// ctx: context
// cfg: k8 config
// yaml: yaml manifest
// ns: namespace
// return: error
func (m *K8) ApplyYaml(ctx context.Context, cfg *rest.Config, yaml string, ns string) error {

	// 1. Prepare a RESTMapper to find GVR
	dc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

	// 2. Prepare the dynamic client
	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return err
	}

	// 3. Decode YAML manifest into unstructured.Unstructured
	obj := &unstructured.Unstructured{}
	_, gvk, err := decUnstructured.Decode([]byte(yaml), nil, obj)
	if err != nil {
		return err
	}

	if ns != "" {
		obj.SetNamespace(ns)
		//log.Print("Info: Setting Namespace " + obj.GetNamespace() + "")
	}

	// 4. Find GVR
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return err
	}

	//Get the namespace
	namespace := ""
	if obj.GetNamespace() != "" {
		namespace = obj.GetNamespace()
	} else {
		namespace = "default"
	}

	// 5. Obtain REST interface for the GVR
	var dr dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		// namespaced resources should specify the namespace
		dr = dyn.Resource(mapping.Resource).Namespace(namespace)
	} else {
		// for cluster-wide resources
		dr = dyn.Resource(mapping.Resource)
	}

	// 6. Marshal object into JSON
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	log.Printf("Info: Patch Kind(%s) Namespace(%s) Name(%s)\n", obj.GetKind(), obj.GetNamespace(), obj.GetName())

	//Show
	if m.dry_run || m.verbose {
		log.Println("Info: Dry Run")
		//lib.FormatResults("**Payload**", yaml)
	}

	_, err = dr.Patch(ctx, obj.GetName(), types.ApplyPatchType, data, metav1.PatchOptions{
		FieldManager: "package-manager",
		DryRun:       []string{m.dryRun(m.dry_run)},
	})
	if err == nil {
		fmt.Printf("Info: Created Kind(%s) Namespace(%s) Name(%s)", obj.GetKind(), obj.GetNamespace(), obj.GetName())
	}
	//*****************
	//Dry Run Show Info
	//*****************
	if m.dry_run || m.verbose {
		if err != nil {
			log.Print(err)
			return err
		}
	}

	if !m.dry_run {
		//Last resort delete and create
		if err != nil {
			fmt.Printf("Info: Error patching Kind(%s) Namespace(%s) Name(%s) Error(%s)", obj.GetKind(), obj.GetNamespace(), obj.GetName(), err.Error())
			//********************
			//Lets delete the item
			//********************
			deletePolicy := metav1.DeletePropagationForeground
			deleteOptions := metav1.DeleteOptions{
				PropagationPolicy: &deletePolicy,
			}

			log.Printf("Info: Cleaning Kind(%s) Namespace(%s) Name(%s)\n", obj.GetKind(), obj.GetNamespace(), obj.GetName())
			dr.Delete(ctx, obj.GetName(), deleteOptions)

			//********
			//Recreate
			//********
			log.Printf("Info: Creating Kind(%s) Namespace(%s) Name(%s)\n", obj.GetKind(), obj.GetNamespace(), obj.GetName())
			_, err = dr.Create(ctx, obj, metav1.CreateOptions{})
			if err != nil {
				return err
			}
			if err == nil {
				fmt.Printf("Info: Created Kind(%s) Namespace(%s) Name(%s) ", obj.GetKind(), obj.GetNamespace(), obj.GetName())
			}
		}
	}

	return err
}

// GetSecrets gets secrets from a k8 cluster
// ns: namespace
// return: v1.SecretList, error
func (m *K8) GetSecrets(ns string) (*v1.SecretList, error) {
	//************
	//Setup the K8
	//************
	config, err := m.buildRestConfig()
	if err != nil {
		return nil, err
	}

	//**********************
	// creates the clientset
	//**********************
	client_set, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// get pods in all the namespaces by omitting namespace
	// Or specify namespace to get pods in particular namespace
	pods, err := client_set.CoreV1().Secrets(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return pods, nil
}

// GetPods gets pods from a k8 cluster
// ns: namespace
// return: v1.PodList, error
func (m *K8) GetPods(ns string) (*v1.PodList, error) {
	//************
	//Setup the K8
	//************
	config, err := m.buildRestConfig()
	if err != nil {
		return nil, err
	}
	//**********************
	// creates the clientset
	//**********************
	client_set, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// get pods in all the namespaces by omitting namespace
	// Or specify namespace to get pods in particular namespace
	pods, err := client_set.CoreV1().Pods(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return pods, nil
}

// GetServices gets services from a k8 cluster
// ns: namespace
//	return: v1.ServiceList, error
func (m *K8) GetServices(ns string) (*v1.ServiceList, error) {
	//************
	//Setup the K8
	//************
	config, err := m.buildRestConfig()
	if err != nil {
		return nil, err
	}
	//**********************
	// creates the clientset
	//**********************
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// get pods in all the namespaces by omitting namespace
	// Or specify namespace to get pods in particular namespace
	pods, err := clientset.CoreV1().Services(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return pods, nil
}

// GetDeployments gets deployments from a k8 cluster
// ns: namespace
// return: appsv1.DeploymentList, error
func (m *K8) GetDeployments(ns string) (*appsv1.DeploymentList, error) {
	//************
	//Setup the K8
	//************
	config, err := m.buildRestConfig()
	if err != nil {
		return nil, err
	}
	//**********************
	// creates the clientset
	//**********************
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// get pods in all the namespaces by omitting namespace
	// Or specify namespace to get pods in particular namespace
	pods, err := clientset.AppsV1().Deployments(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return pods, nil
}

// GetStatefulSets gets statefulsets from a k8 cluster
// ns: namespace
// return: appsv1.StatefulSetList, error
func (m *K8) GetStatefulSets(ns string) (*appsv1.StatefulSetList, error) {
	//************
	//Setup the K8
	//************
	config, err := m.buildRestConfig()
	if err != nil {
		return nil, err
	}
	//**********************
	// creates the clientset
	//**********************
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// get pods in all the namespaces by omitting namespace
	// Or specify namespace to get pods in particular namespace
	pods, err := clientset.AppsV1().StatefulSets(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return pods, nil
}

// GetDemonSet gets demonsets from a k8 cluster
// ns: namespace
// return: appsv1.DaemonSetList, error
func (m *K8) GetDemonSet(ns string) (*appsv1.DaemonSetList, error) {
	//************
	//Setup the K8
	//************
	config, err := m.buildRestConfig()
	if err != nil {
		return nil, err
	}
	//**********************
	// creates the clientset
	//**********************
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// get pods in all the namespaces by omitting namespace
	// Or specify namespace to get pods in particular namespace
	pods, err := clientset.AppsV1().DaemonSets(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return pods, nil
}

// GetServiceIP gets service ip from a k8 cluster
// ns: namespace
// regex_service_name: regex to match service name
// return: v1.ServiceList, error
func (m *K8) GetServiceIP(ns string, regex_service_name string) ([]body_types.ServiceDetails, error) {
	//***************
	//Load the Config
	//***************
	config, err := m.buildRestConfig()
	if err != nil {
		return nil, err
	}

	//**********************
	// creates the clientset
	//**********************
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// get pods in all the namespaces by omitting namespace
	// Or specify namespace to get pods in particular namespace
	services, err := clientset.CoreV1().Services(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var ports []body_types.ServiceDetails
	for _, o := range services.Items {
		if len(o.Status.LoadBalancer.Ingress) > 0 {
			for _, i := range o.Status.LoadBalancer.Ingress {
				res, _ := regexp.MatchString(regex_service_name, o.Name)
				if res {
					if len(o.Spec.Ports) > 0 {

						ports = append(ports, body_types.ServiceDetails{ServiceType: "LoadBalancer", ServiceName: o.Name, IP: i.IP, Port: o.Spec.Ports[0].Port})
						log.Printf("Info: %s; %s:%v", o.Name, i.IP, o.Spec.Ports[0].Port)
					} else {
						ports = append(ports, body_types.ServiceDetails{ServiceType: "LoadBalancer", ServiceName: o.Name, IP: i.IP})
						log.Printf("Info: %s; %s", o.Name, i.IP)
					}

				}
			}
		} else if len(o.Spec.ClusterIPs) > 0 {
			for _, i := range o.Spec.ClusterIPs {
				res, _ := regexp.MatchString(regex_service_name, o.Name)
				if res {
					if len(o.Spec.Ports) > 0 {
						ports = append(ports, body_types.ServiceDetails{ServiceType: "ClusterIP", ServiceName: o.Name, IP: i, Port: o.Spec.Ports[0].Port})
						log.Printf("Info: %s; %s", o.Name, i)
					} else {
						ports = append(ports, body_types.ServiceDetails{ServiceType: "ClusterIP", ServiceName: o.Name, IP: i})
						log.Printf("Info: %s; %s", o.Name, i)
					}

				}
			}
		}
	}

	return ports, nil
}

// CreateNS creates a namespace in a k8 cluster
// ns: namespace
// return: error
func (m *K8) CreateNS(ns string) error {
	//***************
	//Load the Config
	//***************
	config, err := m.buildRestConfig()
	if err != nil {
		return err
	}
	if strings.ToLower(ns) == "default" {
		return errors.New("cannot create default name space")
	}
	ctx := context.Background()

	if ns != "" {
		namespace := "kind: Namespace\napiVersion: v1\nmetadata:\n  name: " + ns + "\n  labels:\n    name: " + ns
		log.Print("Info: Creating Namespace " + ns + " **")
		//**********
		//ApplyYaml
		//**********
		m.ApplyYaml(ctx, config, namespace, ns)
		/*if err != nil {
			return err
		}*/
	}

	return nil
}

// DeleteNS deletes a namespace in a k8 cluster
// ns: namespace
// return: error
// Does not delete default namespace
func (m *K8) DeleteNS(ns string) error {
	//***************
	//Load the Config
	//***************
	config, err := m.buildRestConfig()
	if err != nil {
		return err
	}
	if strings.ToLower(ns) == "default" {
		return errors.New("cannot delete default name space")
	}
	ctx := context.Background()

	if ns != "" {
		namespace := "kind: Namespace\napiVersion: v1\nmetadata:\n  name: " + ns + "\n  labels:\n    name: " + ns
		log.Print("Info: Deleting Namespace " + ns + " **")
		//**********
		//DeleteYaml
		//**********
		m.DeleteYaml(ctx, config, namespace, ns)
		/*if err != nil {
			return err
		}*/
	}

	return nil
}

// CheckStatusOf checks the status of a list of checks
// ns: namespace
// checks: list of checks to perform
// return: bool, []string, error
// bool: true if all checks passed
// []string: list of the results of the checks
// error: error if any
/*
	Example:
		checks := []interface{}{
			"deployment:nginx(.*)",
			"replica:nginx2(.*)",
			"stateful:nginx3(.*)",
			"demon:nginx4(.*)",
			"service:nginx(.*)",
		}
*/
func (m *K8) CheckStatusOf(ns string, checks []interface{}) (bool, []string, error) {
	green := color.FgGreen.Render
	red := color.FgRed.Render
	var results []string
	all_completed := true

	//type:name
	deployments, err := m.GetDeployments(ns)
	if err != nil {
		return false, nil, err
	}
	stateful, err := m.GetStatefulSets(ns)
	if err != nil {
		return false, nil, err
	}
	demonset, err := m.GetDemonSet(ns)
	if err != nil {
		return false, nil, err
	}
	services, err := m.GetServices(ns)
	if err != nil {
		return false, nil, err
	}
	//Loop through the checks
	for _, check := range checks {
		//Split the check into type and name
		checks := strings.Split(check.(string), ":")
		switch checks[0] {
		case "deployment", "replica":
			for _, o := range deployments.Items {
				res, _ := regexp.MatchString(checks[1], o.Name)
				if res {
					//loop through deployment and see if the status is ready
					if o.Status.ReadyReplicas == *o.Spec.Replicas {
						results = append(results, fmt.Sprintf("%s: %s (%v/%v) %s", checks[0], o.GetName(), o.Status.ReadyReplicas, *o.Spec.Replicas, green("Ready")))
					} else {
						all_completed = false
						results = append(results, fmt.Sprintf("%s: %s (%v/%v) %s", checks[0], o.GetName(), o.Status.ReadyReplicas, *o.Spec.Replicas, red("Not Ready")))
					}
					//get the statefulset

				}
			}
		case "stateful":
			for _, o := range stateful.Items {
				res, _ := regexp.MatchString(checks[1], o.Name)
				if res {
					//loop through deployment and see if the status is ready
					if o.Status.ReadyReplicas == *o.Spec.Replicas {
						results = append(results, fmt.Sprintf("%s: %s (%v/%v) %s", checks[0], o.GetName(), o.Status.ReadyReplicas, *o.Spec.Replicas, green("Ready")))
					} else {
						all_completed = false
						results = append(results, fmt.Sprintf("%s: %s (%v/%v) %s", checks[0], o.GetName(), o.Status.ReadyReplicas, *o.Spec.Replicas, red("Not Ready")))
					}
					//get the statefulset

				}
			}
		case "service":
			for _, o := range services.Items {
				res, _ := regexp.MatchString(checks[1], o.Name)
				if res {
					//loop through service and see if the status is ready
					if o.Spec.Type == "LoadBalancer" {
						if len(o.Status.LoadBalancer.Ingress) > 0 {
							results = append(results, fmt.Sprintf("%s: %s %s", checks[0], o.GetName(), green("Ready")))
						} else {
							all_completed = false
							results = append(results, fmt.Sprintf("%s: %s %s", checks[0], o.GetName(), red("Not Ready")))

						}
					} else {
						results = append(results, fmt.Sprintf("%s: %s %s", checks[0], o.GetName(), green("Ready")))
					}
				}
			}
		case "demonset":
			for _, o := range demonset.Items {
				res, _ := regexp.MatchString(checks[1], o.Name)
				if res {
					//loop through deployment and see if the status is ready
					if o.Status.DesiredNumberScheduled == o.Status.NumberReady {
						results = append(results, fmt.Sprintf("%s: %s (%v/%v) %s", checks[0], o.GetName(), o.Status.DesiredNumberScheduled, o.Status.NumberReady, green("Ready")))
					} else {
						all_completed = false
						results = append(results, fmt.Sprintf("%s: %s (%v/%v) %s", checks[0], o.GetName(), o.Status.DesiredNumberScheduled, o.Status.NumberReady, red("Not Ready")))
					}
				}
			}
		}
	}

	return all_completed, results, nil
}
