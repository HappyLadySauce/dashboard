/*
Copyright 2024 The Karmada Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package client

import (
	"errors"
	"fmt"
	"os"
	"sync"

	karmadaclientset "github.com/karmada-io/karmada/pkg/generated/clientset/versioned"
	kubeclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/klog/v2"
)

// proxyURL 代理 URL
const proxyURL = "/apis/cluster.karmada.io/v1alpha1/clusters/%s/proxy/"

var (
	// kubernetesRestConfig 是 Kubernetes 的 rest.Config
	kubernetesRestConfig               *rest.Config
	// kubernetesAPIConfig 是 Kubernetes 的 clientcmdapi.Config
	kubernetesAPIConfig                *clientcmdapi.Config
	// inClusterClient 是 Kubernetes 的客户端
	inClusterClient                    kubeclient.Interface
	// karmadaRestConfig 是 Karmada 的 rest.Config
	karmadaRestConfig                  *rest.Config
	// karmadaAPIConfig 是 Karmada 的 clientcmdapi.Config
	karmadaAPIConfig                   *clientcmdapi.Config
	// karmadaMemberConfig 是 Karmada 的 rest.Config
	karmadaMemberConfig                *rest.Config
	// inClusterKarmadaClient 是 Karmada 的客户端
	inClusterKarmadaClient             karmadaclientset.Interface
	// inClusterClientForKarmadaAPIServer 是 Karmada 的客户端
	inClusterClientForKarmadaAPIServer kubeclient.Interface
	// inClusterClientForMemberAPIServer 是 Karmada 的客户端
	inClusterClientForMemberAPIServer  kubeclient.Interface
	// memberClients 是成员集群的客户端
	memberClients                      sync.Map
)

// configBuilder 是 config 的构建器
type configBuilder struct {
	kubeconfigPath string
	kubeContext    string
	insecure       bool
	userAgent      string
}

// Option 是 configBuilder 的配置选项
type Option func(*configBuilder)

// WithUserAgent 是设置 user agent 的选项
func WithUserAgent(agent string) Option {
	return func(c *configBuilder) {
		c.userAgent = agent
	}
}

// WithKubeconfig 是设置 kubeconfig 路径的选项
func WithKubeconfig(path string) Option {
	return func(c *configBuilder) {
		c.kubeconfigPath = path
	}
}

// WithKubeContext 是设置 kubeconfig 上下文的选项
func WithKubeContext(kubecontext string) Option {
	return func(c *configBuilder) {
		c.kubeContext = kubecontext
	}
}

// WithInsecureTLSSkipVerify 是设置不安全的 TLS 跳过验证的选项
func WithInsecureTLSSkipVerify(insecure bool) Option {
	return func(c *configBuilder) {
		c.insecure = insecure
	}
}

// newConfigBuilder 是创建 configBuilder 的函数
func newConfigBuilder(options ...Option) *configBuilder {
	builder := &configBuilder{}

	for _, opt := range options {
		opt(builder)
	}

	return builder
}

// buildRestConfig 是构建 rest.Config 的函数
func (in *configBuilder) buildRestConfig() (*rest.Config, error) {
	if len(in.kubeconfigPath) == 0 {
		return nil, errors.New("must specify kubeconfig")
	}
	klog.InfoS("Using kubeconfig", "kubeconfig", in.kubeconfigPath)

	restConfig, err := LoadRestConfig(in.kubeconfigPath, in.kubeContext)
	if err != nil {
		return nil, err
	}

	restConfig.QPS = DefaultQPS
	restConfig.Burst = DefaultBurst
	// TODO: make clear that why karmada apiserver seems only can use application/json, however kubernetest apiserver can use "application/vnd.kubernetes.protobuf"
	restConfig.UserAgent = DefaultUserAgent + "/" + in.userAgent
	restConfig.TLSClientConfig.Insecure = in.insecure

	return restConfig, nil
}

// buildAPIConfig 是构建 clientcmdapi.Config 的函数
func (in *configBuilder) buildAPIConfig() (*clientcmdapi.Config, error) {
	if len(in.kubeconfigPath) == 0 {
		return nil, errors.New("must specify kubeconfig")
	}
	klog.InfoS("Using kubeconfig", "kubeconfig", in.kubeconfigPath)
	apiConfig, err := LoadAPIConfig(in.kubeconfigPath, in.kubeContext)
	if err != nil {
		return nil, err
	}
	return apiConfig, nil
}

// isKubeInitialized 检查 Kubernetes config 是否已初始化
func isKubeInitialized() bool {
	if kubernetesRestConfig == nil || kubernetesAPIConfig == nil {
		klog.Errorf(`karmada/karmada-dashboard/client' package has not been initialized properly. Run 'client.InitKubeConfig(...)' to initialize it. `)
		return false
	}
	return true
}

// InitKubeConfig 初始化 Kubernetes 客户端配置
func InitKubeConfig(options ...Option) {
	builder := newConfigBuilder(options...)
	// prefer InClusterConfig, if something wrong, use explicit kubeconfig path
	restConfig, err := rest.InClusterConfig()
	if err == nil {
		klog.Infof("InitKubeConfig by InClusterConfig method")
		restConfig.UserAgent = DefaultUserAgent + "/" + builder.userAgent
		restConfig.TLSClientConfig.Insecure = builder.insecure
		kubernetesRestConfig = restConfig

		apiConfig := ConvertRestConfigToAPIConfig(restConfig)
		kubernetesAPIConfig = apiConfig
	} else {
		klog.Infof("InClusterConfig error: %+v", err)
		klog.Infof("InitKubeConfig by explicit kubeconfig path")
		restConfig, err = builder.buildRestConfig()
		if err != nil {
			klog.Errorf("Could not init client config: %s", err)
			os.Exit(1)
		}
		kubernetesRestConfig = restConfig
		apiConfig, err := builder.buildAPIConfig()
		if err != nil {
			klog.Errorf("Could not init api config: %s", err)
			os.Exit(1)
		}
		kubernetesAPIConfig = apiConfig
	}
}

// InClusterClient 返回一个 Kubernetes 客户端
func InClusterClient() kubeclient.Interface {
	if !isKubeInitialized() {
		return nil
	}

	if inClusterClient != nil {
		return inClusterClient
	}

	// init on-demand only
	c, err := kubeclient.NewForConfig(kubernetesRestConfig)
	if err != nil {
		klog.ErrorS(err, "Could not init kubernetes in-cluster client")
		os.Exit(1)
	}
	// initialize in-memory client
	inClusterClient = c
	return inClusterClient
}

// GetKubeConfig 返回 Kubernetes 客户端配置
func GetKubeConfig() (*rest.Config, *clientcmdapi.Config, error) {
	if !isKubeInitialized() {
		return nil, nil, fmt.Errorf("client package not initialized")
	}
	return kubernetesRestConfig, kubernetesAPIConfig, nil
}

// isKarmadaInitialized 检查 Karmada config 是否已初始化
func isKarmadaInitialized() bool {
	if karmadaRestConfig == nil || karmadaAPIConfig == nil {
		klog.Errorf(`karmada/karmada-dashboard/client' package has not been initialized properly. Run 'client.InitKarmadaConfig(...)' to initialize it. `)
		return false
	}
	return true
}

// InitKarmadaConfig 初始化 Karmada 客户端配置
func InitKarmadaConfig(options ...Option) {
	builder := newConfigBuilder(options...)
	restConfig, err := builder.buildRestConfig()
	if err != nil {
		klog.Errorf("Could not init client config: %s", err)
		os.Exit(1)
	}
	karmadaRestConfig = restConfig

	apiConfig, err := builder.buildAPIConfig()
	if err != nil {
		klog.Errorf("Could not init api config: %s", err)
		os.Exit(1)
	}
	karmadaAPIConfig = apiConfig

	memberConfig, err := builder.buildRestConfig()
	if err != nil {
		klog.Errorf("Could not init member config: %s", err)
		os.Exit(1)
	}
	karmadaMemberConfig = memberConfig
}

// InClusterKarmadaClient 返回一个 Karmada 客户端
func InClusterKarmadaClient() karmadaclientset.Interface {
	if !isKarmadaInitialized() {
		return nil
	}
	if inClusterKarmadaClient != nil {
		return inClusterKarmadaClient
	}
	// init on-demand only
	c, err := karmadaclientset.NewForConfig(karmadaRestConfig)
	if err != nil {
		klog.ErrorS(err, "Could not init karmada in-cluster client")
		os.Exit(1)
	}
	// initialize in-memory client
	inClusterKarmadaClient = c
	return inClusterKarmadaClient
}

// GetKarmadaConfig 返回 Karmada 客户端配置
func GetKarmadaConfig() (*rest.Config, *clientcmdapi.Config, error) {
	if !isKarmadaInitialized() {
		return nil, nil, fmt.Errorf("client package not initialized")
	}
	return karmadaRestConfig, karmadaAPIConfig, nil
}

// GetMemberConfig 返回成员集群的客户端配置
func GetMemberConfig() (*rest.Config, error) {
	if !isKarmadaInitialized() {
		return nil, fmt.Errorf("client package not initialized")
	}
	return karmadaMemberConfig, nil
}

// InClusterClientForKarmadaAPIServer 返回一个 Karmada API 服务器的 Kubernetes 客户端
func InClusterClientForKarmadaAPIServer() kubeclient.Interface {
	if !isKarmadaInitialized() {
		return nil
	}
	if inClusterClientForKarmadaAPIServer != nil {
		return inClusterClientForKarmadaAPIServer
	}
	restConfig, _, err := GetKarmadaConfig()
	if err != nil {
		klog.ErrorS(err, "Could not get karmada restConfig")
		return nil
	}
	c, err := kubeclient.NewForConfig(restConfig)
	if err != nil {
		klog.ErrorS(err, "Could not init kubernetes in-cluster client for karmada apiserver")
		return nil
	}
	inClusterClientForKarmadaAPIServer = c
	return inClusterClientForKarmadaAPIServer
}

// InClusterClientForMemberCluster 返回一个成员集群的 Kubernetes 客户端
func InClusterClientForMemberCluster(clusterName string) kubeclient.Interface {
	if !isKarmadaInitialized() {
		return nil
	}

	// Load and return Interface for member apiserver if already exist
	// 如果成员 API 服务器已经存在，则加载并返回 Interface
	if value, ok := memberClients.Load(clusterName); ok {
		if inClusterClientForMemberAPIServer, ok = value.(kubeclient.Interface); ok {
			return inClusterClientForMemberAPIServer
		}
		klog.Error("Could not get client for member apiserver")
		return nil
	}

	// 为新的成员 API 服务器创建客户端
	restConfig, _, err := GetKarmadaConfig()
	if err != nil {
		klog.ErrorS(err, "Could not get karmada restConfig")
		return nil
	}
	memberConfig, err := GetMemberConfig()
	if err != nil {
		klog.ErrorS(err, "Could not get member restConfig")
		return nil
	}
	memberConfig.Host = restConfig.Host + fmt.Sprintf(proxyURL, clusterName)
	c, err := kubeclient.NewForConfig(memberConfig)
	if err != nil {
		klog.ErrorS(err, "Could not init kubernetes in-cluster client for member apiserver")
		return nil
	}
	inClusterClientForMemberAPIServer = c
	memberClients.Store(clusterName, inClusterClientForMemberAPIServer)
	return inClusterClientForMemberAPIServer
}

// ConvertRestConfigToAPIConfig 将 rest.Config 转换为 clientcmdapi.Config
func ConvertRestConfigToAPIConfig(restConfig *rest.Config) *clientcmdapi.Config {
	// 将 rest.Config 转换为 clientcmdapi.Config
	clientcmdConfig := clientcmdapi.NewConfig()
	clientcmdConfig.Clusters["clusterName"] = &clientcmdapi.Cluster{
		Server:                   restConfig.Host,
		InsecureSkipTLSVerify:    restConfig.Insecure,
		CertificateAuthorityData: restConfig.TLSClientConfig.CAData,
	}

	clientcmdConfig.AuthInfos["authInfoName"] = &clientcmdapi.AuthInfo{
		ClientCertificateData: restConfig.TLSClientConfig.CertData,
		ClientKeyData:         restConfig.TLSClientConfig.KeyData,
	}
	clientcmdConfig.Contexts["contextName"] = &clientcmdapi.Context{
		Cluster:  "clusterName",
		AuthInfo: "authInfoName",
	}
	clientcmdConfig.CurrentContext = "contextName"
	return clientcmdConfig
}
