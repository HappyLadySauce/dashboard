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
	"fmt"
	"net/http"

	karmadaclientset "github.com/karmada-io/karmada/pkg/generated/clientset/versioned"
	kubeclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/klog/v2"
)

// LoadRestConfig creates a rest.Config using the passed kubeconfig. If context is empty, current context in kubeconfig will be used.
// LoadRestConfig 使用传递的 kubeconfig 创建一个 rest.Config。如果 context 为空，则使用 kubeconfig 中的当前上下文。
func LoadRestConfig(kubeconfig string, context string) (*rest.Config, error) {
	loader := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig}
	loadedConfig, err := loader.Load()
	if err != nil {
		return nil, err
	}

	if context == "" {
		context = loadedConfig.CurrentContext
	}
	klog.Infof("Use context %v", context)

	return clientcmd.NewNonInteractiveClientConfig(
		*loadedConfig,
		context,
		&clientcmd.ConfigOverrides{},
		loader,
	).ClientConfig()
}

// LoadAPIConfig creates a clientcmdapi.Config using the passed kubeconfig. If currentContext is empty, current context in kubeconfig will be used.
// LoadAPIConfig 使用传递的 kubeconfig 创建一个 clientcmdapi.Config。如果 currentContext 为空，则使用 kubeconfig 中的当前上下文。
func LoadAPIConfig(kubeconfig string, currentContext string) (*clientcmdapi.Config, error) {
	config, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		return nil, err
	}
	if currentContext == "" {
		currentContext = config.CurrentContext
	}
	context, exist := config.Contexts[currentContext]
	if !exist {
		return nil, fmt.Errorf("context:%s not exist", currentContext)
	}
	clusterName := context.Cluster
	authInfoName := context.AuthInfo
	cluster := config.Clusters[clusterName]
	authInfo := config.AuthInfos[authInfoName]

	apiConfig := &clientcmdapi.Config{
		Clusters: map[string]*clientcmdapi.Cluster{
			clusterName: cluster,
		},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			authInfoName: authInfo,
		},
		Contexts: map[string]*clientcmdapi.Context{
			currentContext: {
				Cluster:  clusterName,
				AuthInfo: authInfoName,
			},
		},
		CurrentContext: currentContext,
	}
	return apiConfig, nil
}

// LoadRestConfigFromKubeConfig creates a rest.Config from a kubeconfig string.
// LoadRestConfigFromKubeConfig 从 kubeconfig 字符串创建一个 rest.Config。
func LoadRestConfigFromKubeConfig(kubeconfig string) (*rest.Config, error) {
	apiConfig, err := clientcmd.Load([]byte(kubeconfig))
	if err != nil {
		return nil, err
	}
	clientConfig := clientcmd.NewDefaultClientConfig(*apiConfig, &clientcmd.ConfigOverrides{})
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	return restConfig, nil
}

// KubeClientSetFromKubeConfig creates a Kubernetes clientset from a kubeconfig string.
// KubeClientSetFromKubeConfig 从 kubeconfig 字符串创建一个 Kubernetes 客户端。
func KubeClientSetFromKubeConfig(kubeconfig string) (*kubeclient.Clientset, error) {
	restConfig, err := LoadRestConfigFromKubeConfig(kubeconfig)
	if err != nil {
		return nil, err
	}
	kubeClient := kubeclient.NewForConfigOrDie(restConfig)
	return kubeClient, nil
}

// GetKarmadaClientFromRequest creates a Karmada clientset from an HTTP request.
// GetKarmadaClientFromRequest 从 HTTP 请求创建一个 Karmada 客户端
func GetKarmadaClientFromRequest(request *http.Request) (karmadaclientset.Interface, error) {
	// 检查 Karmada 是否已初始化
	if !isKarmadaInitialized() {
		return nil, fmt.Errorf("client package not initialized")
	}
	// 从 HTTP 请求创建 Karmada 客户端
	return karmadaClientFromRequest(request)
}

// karmadaClientFromRequest 从 HTTP 请求创建一个 Karmada 客户端
func karmadaClientFromRequest(request *http.Request) (karmadaclientset.Interface, error) {
	// 从 HTTP 请求创建 Karmada 配置
	config, err := karmadaConfigFromRequest(request)
	if err != nil {
		return nil, err
	}

	return karmadaclientset.NewForConfig(config)
}
