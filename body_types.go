package go_k8_helm

type ServiceDetails struct {
	ServiceName string `json:"service_name" yaml:"service_name"`
	ServiceType string `json:"service_type" yaml:"service_type"`
	IP          string `json:"ip" yaml:"ip"`
	Port        int32  `json:"port" yaml:"port"`
}
