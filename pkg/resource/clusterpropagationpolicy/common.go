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

package clusterpropagationpolicy

import (
	"github.com/karmada-io/karmada/pkg/apis/policy/v1alpha1"

	"github.com/karmada-io/dashboard/pkg/dataselect"
)

// ClusterPropagationPolicyCell wraps v1alpha1.ClusterPropagationPolicy for data selection.
// ClusterPropagationPolicyCell 用于数据选择。
type ClusterPropagationPolicyCell v1alpha1.ClusterPropagationPolicy

// GetProperty 返回集群传播策略单元格的属性。
func (c ClusterPropagationPolicyCell) GetProperty(name dataselect.PropertyName) dataselect.ComparableValue {
	switch name {
	case dataselect.NameProperty:
		return dataselect.StdComparableString(c.ObjectMeta.Name)
	case dataselect.CreationTimestampProperty:
		return dataselect.StdComparableTime(c.ObjectMeta.CreationTimestamp.Time)
	default:
		// if name is not supported then just return a constant dummy value, sort will have no effect.
		return nil
	}
}

// toCells 将v1alpha1.ClusterPropagationPolicy对象列表转换为dataselect.DataCell列表。
func toCells(std []v1alpha1.ClusterPropagationPolicy) []dataselect.DataCell {
	cells := make([]dataselect.DataCell, len(std))
	for i := range std {
		cells[i] = ClusterPropagationPolicyCell(std[i])
	}
	return cells
}

// fromCells 将dataselect.DataCell列表转换为v1alpha1.ClusterPropagationPolicy对象列表。
func fromCells(cells []dataselect.DataCell) []v1alpha1.ClusterPropagationPolicy {
	std := make([]v1alpha1.ClusterPropagationPolicy, len(cells))
	for i := range std {
		std[i] = v1alpha1.ClusterPropagationPolicy(cells[i].(ClusterPropagationPolicyCell))
	}
	return std
}
