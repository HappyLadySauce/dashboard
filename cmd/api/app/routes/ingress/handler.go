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

package ingress

import (
	"github.com/gin-gonic/gin"

	"github.com/karmada-io/dashboard/cmd/api/app/router"
	"github.com/karmada-io/dashboard/cmd/api/app/types/common"
	"github.com/karmada-io/dashboard/pkg/client"
	"github.com/karmada-io/dashboard/pkg/resource/ingress"
)

// 获取ingress列表
func handleGetIngress(c *gin.Context) {
	k8sClient := client.InClusterClientForKarmadaAPIServer()
	dataSelect := common.ParseDataSelectPathParameter(c)
	nsQuery := common.ParseNamespacePathParameter(c)
	result, err := ingress.GetIngressList(k8sClient, nsQuery, dataSelect)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.Success(c, result)
}

// 获取ingress详情
func handleGetIngressDetail(c *gin.Context) {
	k8sClient := client.InClusterClientForKarmadaAPIServer()
	namespace := c.Param("namespace")
	name := c.Param("service")
	result, err := ingress.GetIngressDetail(k8sClient, namespace, name)
	if err != nil {
		common.Fail(c, err)
		return
	}
	common.Success(c, result)
}

// 初始化路由
func init() {
	r := router.V1()
	// 获取ingress列表
	r.GET("/ingress", handleGetIngress)
	// 获取ingress列表
	r.GET("/ingress/:namespace", handleGetIngress)
	// 获取ingress详情
	r.GET("/ingress/:namespace/:service", handleGetIngressDetail)
}
