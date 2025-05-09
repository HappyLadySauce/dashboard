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

package v1

// CreateDeploymentRequest defines the request structure for creating a deployment.
// CreateDeploymentRequest 是创建部署的请求
type CreateDeploymentRequest struct {
	// Namespace 是命名空间
	Namespace string `json:"namespace"`
	// Name 是名称
	Name      string `json:"name"`
	// Content 是内容
	Content   string `json:"content"`
}

// CreateDeploymentResponse defines the response structure for creating a deployment.
// CreateDeploymentResponse 是创建部署的响应
type CreateDeploymentResponse struct{}
