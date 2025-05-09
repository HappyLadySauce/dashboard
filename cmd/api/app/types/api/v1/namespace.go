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

// CreateNamesapceRequest is the request body for creating a namespace.
// CreateNamesapceRequest 是创建命名空间的请求
type CreateNamesapceRequest struct {
	// Name 是命名空间的名称
	Name                string `json:"name" required:"true"`
	// SkipAutoPropagation 是是否跳过自动传播
	SkipAutoPropagation bool   `json:"skipAutoPropagation"`
}

// CreateNamesapceResponse is the response body for creating a namespace.
// CreateNamesapceResponse 是创建命名空间的响应
type CreateNamesapceResponse struct{}
