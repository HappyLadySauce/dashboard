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

package router

import (
	"github.com/gin-gonic/gin"

	"github.com/karmada-io/dashboard/pkg/environment"
)

var (
	router *gin.Engine
	v1     *gin.RouterGroup
)

// init 初始化函数
func init() {
	if !environment.IsDev() {
		gin.SetMode(gin.ReleaseMode)
	}

	router = gin.Default()
	_ = router.SetTrustedProxies(nil)
	v1 = router.Group("/api/v1")

	router.GET("/livez", func(c *gin.Context) {
		c.String(200, "livez")
	})
	router.GET("/readyz", func(c *gin.Context) {
		c.String(200, "readyz")
	})
}

// V1 返回 /api/v1 的路由组
func V1() *gin.RouterGroup {
	return v1
}

// Router 返回主 Gin 引擎实例
func Router() *gin.Engine {
	return router
}
