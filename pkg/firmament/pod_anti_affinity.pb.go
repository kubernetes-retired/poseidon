/*
Copyright 2018 The Kubernetes Authors.

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

// Code generated by protoc-gen-go. DO NOT EDIT.
// source: pod_anti_affinity.proto

package firmament

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type MatchLabelsAntiAff struct {
	Key   string `protobuf:"bytes,1,opt,name=key" json:"key,omitempty"`
	Value string `protobuf:"bytes,2,opt,name=value" json:"value,omitempty"`
}

func (m *MatchLabelsAntiAff) Reset()                    { *m = MatchLabelsAntiAff{} }
func (m *MatchLabelsAntiAff) String() string            { return proto.CompactTextString(m) }
func (*MatchLabelsAntiAff) ProtoMessage()               {}
func (*MatchLabelsAntiAff) Descriptor() ([]byte, []int) { return fileDescriptor8, []int{0} }

func (m *MatchLabelsAntiAff) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

func (m *MatchLabelsAntiAff) GetValue() string {
	if m != nil {
		return m.Value
	}
	return ""
}

type LabelSelectorRequirementAntiAff struct {
	Key      string   `protobuf:"bytes,1,opt,name=key" json:"key,omitempty"`
	Operator string   `protobuf:"bytes,2,opt,name=operator" json:"operator,omitempty"`
	Values   []string `protobuf:"bytes,3,rep,name=values" json:"values,omitempty"`
}

func (m *LabelSelectorRequirementAntiAff) Reset()                    { *m = LabelSelectorRequirementAntiAff{} }
func (m *LabelSelectorRequirementAntiAff) String() string            { return proto.CompactTextString(m) }
func (*LabelSelectorRequirementAntiAff) ProtoMessage()               {}
func (*LabelSelectorRequirementAntiAff) Descriptor() ([]byte, []int) { return fileDescriptor8, []int{1} }

func (m *LabelSelectorRequirementAntiAff) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

func (m *LabelSelectorRequirementAntiAff) GetOperator() string {
	if m != nil {
		return m.Operator
	}
	return ""
}

func (m *LabelSelectorRequirementAntiAff) GetValues() []string {
	if m != nil {
		return m.Values
	}
	return nil
}

type LabelSelectorAntiAff struct {
	MatchLabels      *MatchLabelsAntiAff                `protobuf:"bytes,1,opt,name=matchLabels" json:"matchLabels,omitempty"`
	MatchExpressions []*LabelSelectorRequirementAntiAff `protobuf:"bytes,2,rep,name=matchExpressions" json:"matchExpressions,omitempty"`
}

func (m *LabelSelectorAntiAff) Reset()                    { *m = LabelSelectorAntiAff{} }
func (m *LabelSelectorAntiAff) String() string            { return proto.CompactTextString(m) }
func (*LabelSelectorAntiAff) ProtoMessage()               {}
func (*LabelSelectorAntiAff) Descriptor() ([]byte, []int) { return fileDescriptor8, []int{2} }

func (m *LabelSelectorAntiAff) GetMatchLabels() *MatchLabelsAntiAff {
	if m != nil {
		return m.MatchLabels
	}
	return nil
}

func (m *LabelSelectorAntiAff) GetMatchExpressions() []*LabelSelectorRequirementAntiAff {
	if m != nil {
		return m.MatchExpressions
	}
	return nil
}

type PodAffinityTermAntiAff struct {
	LabelSelector *LabelSelectorAntiAff `protobuf:"bytes,1,opt,name=labelSelector" json:"labelSelector,omitempty"`
	Namespaces    []string              `protobuf:"bytes,2,rep,name=namespaces" json:"namespaces,omitempty"`
	// Empty topologyKey is not allowed.
	TopologyKey string `protobuf:"bytes,3,opt,name=topologyKey" json:"topologyKey,omitempty"`
}

func (m *PodAffinityTermAntiAff) Reset()                    { *m = PodAffinityTermAntiAff{} }
func (m *PodAffinityTermAntiAff) String() string            { return proto.CompactTextString(m) }
func (*PodAffinityTermAntiAff) ProtoMessage()               {}
func (*PodAffinityTermAntiAff) Descriptor() ([]byte, []int) { return fileDescriptor8, []int{3} }

func (m *PodAffinityTermAntiAff) GetLabelSelector() *LabelSelectorAntiAff {
	if m != nil {
		return m.LabelSelector
	}
	return nil
}

func (m *PodAffinityTermAntiAff) GetNamespaces() []string {
	if m != nil {
		return m.Namespaces
	}
	return nil
}

func (m *PodAffinityTermAntiAff) GetTopologyKey() string {
	if m != nil {
		return m.TopologyKey
	}
	return ""
}

type WeightedPodAffinityTermAntiAff struct {
	// weight associated with matching the corresponding podAffinityTerm,
	// in the range 1-100.
	Weight int32 `protobuf:"varint,1,opt,name=weight" json:"weight,omitempty"`
	// A pod affinity term, associated with the corresponding weight.
	PodAffinityTerm *PodAffinityTermAntiAff `protobuf:"bytes,2,opt,name=podAffinityTerm" json:"podAffinityTerm,omitempty"`
}

func (m *WeightedPodAffinityTermAntiAff) Reset()                    { *m = WeightedPodAffinityTermAntiAff{} }
func (m *WeightedPodAffinityTermAntiAff) String() string            { return proto.CompactTextString(m) }
func (*WeightedPodAffinityTermAntiAff) ProtoMessage()               {}
func (*WeightedPodAffinityTermAntiAff) Descriptor() ([]byte, []int) { return fileDescriptor8, []int{4} }

func (m *WeightedPodAffinityTermAntiAff) GetWeight() int32 {
	if m != nil {
		return m.Weight
	}
	return 0
}

func (m *WeightedPodAffinityTermAntiAff) GetPodAffinityTerm() *PodAffinityTermAntiAff {
	if m != nil {
		return m.PodAffinityTerm
	}
	return nil
}

type PodAntiAffinity struct {
	RequiredDuringSchedulingIgnoredDuringExecution  []*PodAffinityTermAntiAff         `protobuf:"bytes,1,rep,name=requiredDuringSchedulingIgnoredDuringExecution" json:"requiredDuringSchedulingIgnoredDuringExecution,omitempty"`
	PreferredDuringSchedulingIgnoredDuringExecution []*WeightedPodAffinityTermAntiAff `protobuf:"bytes,2,rep,name=preferredDuringSchedulingIgnoredDuringExecution" json:"preferredDuringSchedulingIgnoredDuringExecution,omitempty"`
}

func (m *PodAntiAffinity) Reset()                    { *m = PodAntiAffinity{} }
func (m *PodAntiAffinity) String() string            { return proto.CompactTextString(m) }
func (*PodAntiAffinity) ProtoMessage()               {}
func (*PodAntiAffinity) Descriptor() ([]byte, []int) { return fileDescriptor8, []int{5} }

func (m *PodAntiAffinity) GetRequiredDuringSchedulingIgnoredDuringExecution() []*PodAffinityTermAntiAff {
	if m != nil {
		return m.RequiredDuringSchedulingIgnoredDuringExecution
	}
	return nil
}

func (m *PodAntiAffinity) GetPreferredDuringSchedulingIgnoredDuringExecution() []*WeightedPodAffinityTermAntiAff {
	if m != nil {
		return m.PreferredDuringSchedulingIgnoredDuringExecution
	}
	return nil
}

func init() {
	proto.RegisterType((*MatchLabelsAntiAff)(nil), "firmament.MatchLabelsAntiAff")
	proto.RegisterType((*LabelSelectorRequirementAntiAff)(nil), "firmament.LabelSelectorRequirementAntiAff")
	proto.RegisterType((*LabelSelectorAntiAff)(nil), "firmament.LabelSelectorAntiAff")
	proto.RegisterType((*PodAffinityTermAntiAff)(nil), "firmament.PodAffinityTermAntiAff")
	proto.RegisterType((*WeightedPodAffinityTermAntiAff)(nil), "firmament.WeightedPodAffinityTermAntiAff")
	proto.RegisterType((*PodAntiAffinity)(nil), "firmament.PodAntiAffinity")
}

func init() { proto.RegisterFile("pod_anti_affinity.proto", fileDescriptor8) }

var fileDescriptor8 = []byte{
	// 425 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x93, 0xc1, 0x8e, 0xd3, 0x30,
	0x10, 0x86, 0xe5, 0x46, 0x5b, 0x91, 0x89, 0xd0, 0xae, 0xac, 0x55, 0x89, 0x90, 0xd8, 0x0d, 0x39,
	0x15, 0x0e, 0x41, 0x2a, 0x57, 0x24, 0x54, 0x89, 0x1e, 0xd0, 0x82, 0x84, 0xbc, 0x08, 0x8e, 0x2b,
	0x6f, 0x32, 0x49, 0x2d, 0x12, 0xdb, 0x38, 0x0e, 0x6c, 0x1e, 0x80, 0x03, 0x67, 0x9e, 0x80, 0x27,
	0xe0, 0xc0, 0x0b, 0xa2, 0x3a, 0x69, 0x49, 0x59, 0x0a, 0xed, 0x2d, 0x33, 0xf1, 0xff, 0xfd, 0xff,
	0x8c, 0x13, 0xb8, 0xa7, 0x55, 0x76, 0xc5, 0xa5, 0x15, 0x57, 0x3c, 0xcf, 0x85, 0x14, 0xb6, 0x4d,
	0xb4, 0x51, 0x56, 0x51, 0x3f, 0x17, 0xa6, 0xe2, 0x15, 0x4a, 0x1b, 0x3f, 0x03, 0xfa, 0x9a, 0xdb,
	0x74, 0xf9, 0x8a, 0x5f, 0x63, 0x59, 0xcf, 0xa5, 0x15, 0xf3, 0x3c, 0xa7, 0x27, 0xe0, 0x7d, 0xc0,
	0x36, 0x24, 0x11, 0x99, 0xfa, 0x6c, 0xf5, 0x48, 0x4f, 0xe1, 0xe8, 0x13, 0x2f, 0x1b, 0x0c, 0x47,
	0xae, 0xd7, 0x15, 0x71, 0x01, 0xe7, 0x4e, 0x78, 0x89, 0x25, 0xa6, 0x56, 0x19, 0x86, 0x1f, 0x1b,
	0x61, 0x70, 0x45, 0xde, 0x8d, 0xba, 0x0f, 0x77, 0x94, 0x46, 0xc3, 0xad, 0x32, 0x3d, 0x6d, 0x53,
	0xd3, 0x09, 0x8c, 0x1d, 0xb9, 0x0e, 0xbd, 0xc8, 0x9b, 0xfa, 0xac, 0xaf, 0xe2, 0x1f, 0x04, 0x4e,
	0xb7, 0x9c, 0xd6, 0xf8, 0xe7, 0x10, 0x54, 0xbf, 0xf3, 0x3b, 0x9b, 0x60, 0xf6, 0x20, 0xd9, 0x0c,
	0x98, 0xdc, 0x9e, 0x8e, 0x0d, 0x15, 0xf4, 0x1d, 0x9c, 0xb8, 0x72, 0x71, 0xa3, 0x0d, 0xd6, 0xb5,
	0x50, 0xb2, 0x0e, 0x47, 0x91, 0x37, 0x0d, 0x66, 0x8f, 0x07, 0x94, 0xff, 0x4c, 0xc9, 0x6e, 0x31,
	0xe2, 0xef, 0x04, 0x26, 0x6f, 0x54, 0x36, 0xef, 0x37, 0xff, 0x16, 0x4d, 0xb5, 0xce, 0xbc, 0x80,
	0xbb, 0xe5, 0x90, 0xd7, 0xa7, 0x3e, 0xdf, 0xe5, 0xb7, 0x36, 0xd9, 0x56, 0xd1, 0x33, 0x00, 0xc9,
	0x2b, 0xac, 0x35, 0x4f, 0xb1, 0xcb, 0xec, 0xb3, 0x41, 0x87, 0x46, 0x10, 0x58, 0xa5, 0x55, 0xa9,
	0x8a, 0xf6, 0x02, 0xdb, 0xd0, 0x73, 0xab, 0x1e, 0xb6, 0xe2, 0x2f, 0x04, 0xce, 0xde, 0xa3, 0x28,
	0x96, 0x16, 0xb3, 0x1d, 0x59, 0x27, 0x30, 0xfe, 0xec, 0x4e, 0xb8, 0x90, 0x47, 0xac, 0xaf, 0xe8,
	0x05, 0x1c, 0xeb, 0x6d, 0x85, 0xbb, 0xcb, 0x60, 0xf6, 0x70, 0x30, 0xc5, 0xdf, 0x99, 0xec, 0x4f,
	0x65, 0xfc, 0x73, 0x04, 0xc7, 0xab, 0xb3, 0xdd, 0x7b, 0xd7, 0xa7, 0x5f, 0x09, 0x24, 0xa6, 0x5b,
	0x74, 0xf6, 0xa2, 0x31, 0x42, 0x16, 0x97, 0xe9, 0x12, 0xb3, 0xa6, 0x14, 0xb2, 0x78, 0x59, 0x48,
	0xb5, 0x69, 0x2f, 0x6e, 0x30, 0x6d, 0xac, 0x50, 0x32, 0x24, 0xee, 0xda, 0xf6, 0x08, 0x70, 0x20,
	0x98, 0x7e, 0x23, 0xf0, 0x44, 0x1b, 0xcc, 0xd1, 0xec, 0x1f, 0xa6, 0xfb, 0x86, 0x1e, 0x0d, 0xc2,
	0xfc, 0x7b, 0xd3, 0xec, 0x50, 0x87, 0xeb, 0xb1, 0xfb, 0x99, 0x9f, 0xfe, 0x0a, 0x00, 0x00, 0xff,
	0xff, 0x61, 0x16, 0x84, 0x1d, 0xe7, 0x03, 0x00, 0x00,
}