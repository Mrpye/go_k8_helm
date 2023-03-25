package go_k8_helm

import (
	"context"
	"fmt"

	"k8s.io/client-go/rest"
)

// K8 is the struct for the k8 connection
// If you want to use the kube config file, you need to set the DefaultContext and ConfigPath
// DefaultContext is the default context to use
// ConfigPath is the path to the kube config file
// Host is the host to connect to
// You will need to set the Host and Authorization if you want to use the token connection
// Authorization is the authorization token
// UseTokenConnection if true, use the token connection, otherwise use the kube config file
// Ignore_ssl if true, ignore the ssl connection
type K8 struct {
	DefaultContext     string `json:"default_context" yaml:"default_context" flag:"context c" desc:"The default context to use"`
	ConfigPath         string `json:"config_path" yaml:"config_path" flag:"config_path p" desc:"The path to the kube config file"`
	Host               string `json:"host" yaml:"host" flag:"host h" desc:"The host to connect to"`
	Authorization      string `json:"authorization" yaml:"authorization" flag:"auth a" desc:"The authorization token"`
	UseTokenConnection bool   `json:"use_token_connection" yaml:"use_token_connection" flag:"conn-type u" desc:"Connection type if true, use the token connection, otherwise use the kube config file"` //if true, use the token connection, otherwise use the kube config file
	Ignore_ssl         bool   `json:"ignore_ssl" yaml:"ignore_ssl" flag:"ignore_ssl i" desc:"If true, ignore the ssl connection"`
	dry_run            bool
	verbose            bool
	config             *rest.Config
	ctx                context.Context
}

// buildRestConfig builds the rest config
// And add to the k8 type
// returns error if there is an issue
func (m *K8) CreateConfigAndContext() error {
	m.ctx = context.Background()
	cfg, err := m.buildRestConfig()
	if err != nil {
		return err
	}
	m.config = cfg
	return nil
}

// Verbose returns the verbose flag
func (m *K8) Verbose() bool {
	return m.verbose
}

// SetVerbose sets the verbose flag
// If true, print out the verbose information
func (m *K8) SetVerbose(verbose bool) {
	m.verbose = verbose
}

// DryRun returns the dry_run flag
// If true, do not execute the command, just print out the command
func (m *K8) DryRun() bool {
	return m.dry_run
}

// SetDryRun sets the dry_run flag
// If true, do not execute the command, just print out the command
func (m *K8) SetDryRun(dry_run bool) {
	m.dry_run = dry_run
}

// K8Option is the option for the k8 connection
type K8Option func(*K8)

// OptionK8DefaultContext is the option for the default context
func OptionK8DefaultContext(default_context string) K8Option {
	return func(h *K8) {
		h.DefaultContext = default_context
	}
}

// OptionK8ConfigPath is the option for the config path
func OptionK8ConfigPath(config_path string) K8Option {
	return func(h *K8) {
		h.ConfigPath = config_path
	}
}

// OptionK8Host is the option for the host
func OptionK8Host(host string) K8Option {
	return func(h *K8) {
		h.Host = host
	}
}

// OptionK8Auth is the option for the authorization
func OptionK8Auth(auth string) K8Option {
	return func(h *K8) {
		h.Authorization = auth
	}
}

// OptionK8IgnoreSSL is the option for the ignore ssl
func OptionK8IgnoreSSL(ignore_ssl bool) K8Option {
	return func(h *K8) {
		h.Ignore_ssl = ignore_ssl
	}
}

// OptionK8UseTokenConnection is the option for the use token connection
func OptionK8UseTokenConnection(use_token_connection bool) K8Option {
	return func(h *K8) {
		h.UseTokenConnection = use_token_connection
	}
}

// Update the k8 Type with the options
func (m *K8) Update(opts ...K8Option) {
	// Loop through each option
	for _, opt := range opts {
		// Call the option giving the instantiated
		opt(m)
	}
}

// String returns the string representation of the k8 type
func (m *K8) String() string {
	return fmt.Sprintf("%s,%s", m.DefaultContext, m.ConfigPath)
}

// Create a instance of the k8 type
// default_context is the default context to use
// config_path is the path to the kube config file
func CreateK8KubeConfig(default_context string, config_path string) (*K8, error) {
	k8 := &K8{
		DefaultContext:     default_context,
		ConfigPath:         config_path,
		UseTokenConnection: false,
	}
	k8.ctx = context.Background()
	cfg, err := k8.buildRestConfig()
	if err != nil {
		return nil, err
	}
	k8.config = cfg
	return k8, nil

}

// CreateK8Token creates a instance of the k8 type
// host is the host to connect to
// auth is the authorization token
// ignore_ssl if true, ignore the ssl connection
// returns the k8 type
// returns an error if there is an issue
func CreateK8Token(host string, auth string, ignore_ssl bool) (*K8, error) {
	k8 := &K8{
		Host:               host,
		Authorization:      auth,
		Ignore_ssl:         ignore_ssl,
		UseTokenConnection: true,
	}
	k8.ctx = context.Background()
	cfg, err := k8.buildRestConfig()
	if err != nil {
		return nil, err
	}
	k8.config = cfg
	return k8, nil
}

// CreateK8Options creates a instance of the k8 type
// opts are the options for the k8 type
// returns the k8 type
// returns an error if there is an issue
func CreateK8Options(opts ...K8Option) (*K8, error) {
	k8 := &K8{}
	k8.Update(opts...)
	k8.ctx = context.Background()
	cfg, err := k8.buildRestConfig()
	if err != nil {
		return nil, err
	}
	k8.config = cfg
	return k8, nil
}

// CreateK8 creates a instance of the k8 type
// Does not create the config and context
// Use  CreateConfigAndContext() to create the config and context
// opts are the options for the k8 type
// returns the k8 type
// returns an error if there is an issue
func CreateK8(opts ...K8Option) (*K8, error) {
	k8 := &K8{}
	k8.Update(opts...)
	return k8, nil
}
