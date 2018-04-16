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

// Code generated by protoc-gen-go.
// source: resource_stats.proto
// DO NOT EDIT!

package firmament

import (
	proto "github.com/golang/protobuf/proto"

	"fmt"
	"math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type ResourceStats struct {
	ResourceId string `protobuf:"bytes,1,opt,name=resource_id,json=resourceId" json:"resource_id,omitempty"`
	Timestamp  uint64 `protobuf:"varint,2,opt,name=timestamp" json:"timestamp,omitempty"`
	// The first entry is the cpu usage of cpu0 and so on.
	CpusStats []*CpuStats `protobuf:"bytes,3,rep,name=cpus_stats,json=cpusStats" json:"cpus_stats,omitempty"`
	// Memory stats (in KB).
	MemAllocatable int64 `protobuf:"varint,4,opt,name=mem_allocatable,json=memAllocatable" json:"mem_allocatable,omitempty"`
	MemCapacity    int64 `protobuf:"varint,5,opt,name=mem_capacity,json=memCapacity" json:"mem_capacity,omitempty"`
	// Memory stats (fraction of total).
	MemReservation float64 `protobuf:"fixed64,6,opt,name=mem_reservation,json=memReservation" json:"mem_reservation,omitempty"`
	MemUtilization float64 `protobuf:"fixed64,7,opt,name=mem_utilization,json=memUtilization" json:"mem_utilization,omitempty"`
	// Disk stats in KB
	DiskBw int64 `protobuf:"varint,8,opt,name=disk_bw,json=diskBw" json:"disk_bw,omitempty"`
	// Network stats in KB
	NetRxBw int64 `protobuf:"varint,9,opt,name=net_rx_bw,json=netRxBw" json:"net_rx_bw,omitempty"`
	NetTxBw int64 `protobuf:"varint,10,opt,name=net_tx_bw,json=netTxBw" json:"net_tx_bw,omitempty"`
}

func (m *ResourceStats) Reset()                    { *m = ResourceStats{} }
func (m *ResourceStats) String() string            { return proto.CompactTextString(m) }
func (*ResourceStats) ProtoMessage()               {}
func (*ResourceStats) Descriptor() ([]byte, []int) { return fileDescriptor7, []int{0} }

func (m *ResourceStats) GetResourceId() string {
	if m != nil {
		return m.ResourceId
	}
	return ""
}

func (m *ResourceStats) GetTimestamp() uint64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

func (m *ResourceStats) GetCpusStats() []*CpuStats {
	if m != nil {
		return m.CpusStats
	}
	return nil
}

func (m *ResourceStats) GetMemAllocatable() int64 {
	if m != nil {
		return m.MemAllocatable
	}
	return 0
}

func (m *ResourceStats) GetMemCapacity() int64 {
	if m != nil {
		return m.MemCapacity
	}
	return 0
}

func (m *ResourceStats) GetMemReservation() float64 {
	if m != nil {
		return m.MemReservation
	}
	return 0
}

func (m *ResourceStats) GetMemUtilization() float64 {
	if m != nil {
		return m.MemUtilization
	}
	return 0
}

func (m *ResourceStats) GetDiskBw() int64 {
	if m != nil {
		return m.DiskBw
	}
	return 0
}

func (m *ResourceStats) GetNetRxBw() int64 {
	if m != nil {
		return m.NetRxBw
	}
	return 0
}

func (m *ResourceStats) GetNetTxBw() int64 {
	if m != nil {
		return m.NetTxBw
	}
	return 0
}

type CpuStats struct {
	// CPU stats in millicores.
	CpuAllocatable int64 `protobuf:"varint,1,opt,name=cpu_allocatable,json=cpuAllocatable" json:"cpu_allocatable,omitempty"`
	CpuCapacity    int64 `protobuf:"varint,2,opt,name=cpu_capacity,json=cpuCapacity" json:"cpu_capacity,omitempty"`
	// CPU stats (fraction of total).
	CpuReservation float64 `protobuf:"fixed64,3,opt,name=cpu_reservation,json=cpuReservation" json:"cpu_reservation,omitempty"`
	CpuUtilization float64 `protobuf:"fixed64,4,opt,name=cpu_utilization,json=cpuUtilization" json:"cpu_utilization,omitempty"`
}

func (m *CpuStats) Reset()                    { *m = CpuStats{} }
func (m *CpuStats) String() string            { return proto.CompactTextString(m) }
func (*CpuStats) ProtoMessage()               {}
func (*CpuStats) Descriptor() ([]byte, []int) { return fileDescriptor7, []int{1} }

func (m *CpuStats) GetCpuAllocatable() int64 {
	if m != nil {
		return m.CpuAllocatable
	}
	return 0
}

func (m *CpuStats) GetCpuCapacity() int64 {
	if m != nil {
		return m.CpuCapacity
	}
	return 0
}

func (m *CpuStats) GetCpuReservation() float64 {
	if m != nil {
		return m.CpuReservation
	}
	return 0
}

func (m *CpuStats) GetCpuUtilization() float64 {
	if m != nil {
		return m.CpuUtilization
	}
	return 0
}

func init() {
	proto.RegisterType((*ResourceStats)(nil), "firmament.ResourceStats")
	proto.RegisterType((*CpuStats)(nil), "firmament.CpuStats")
}

func init() { proto.RegisterFile("resource_stats.proto", fileDescriptor7) }

var fileDescriptor7 = []byte{
	// 330 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x5c, 0x92, 0x4f, 0x4e, 0xe3, 0x30,
	0x1c, 0x85, 0xe5, 0xa6, 0xd3, 0x36, 0xee, 0xcc, 0x20, 0x19, 0x24, 0x2c, 0x84, 0x44, 0xe8, 0x86,
	0xac, 0xba, 0x28, 0x27, 0xa0, 0x5d, 0xb1, 0x35, 0xb0, 0x8e, 0x5c, 0xd7, 0x48, 0x16, 0x71, 0x62,
	0xd9, 0x3f, 0xd3, 0xc2, 0x89, 0xb8, 0x1e, 0x37, 0x40, 0x76, 0x9a, 0x3f, 0x74, 0x99, 0xf7, 0xbe,
	0xf8, 0xe5, 0x8b, 0x8c, 0x2f, 0xac, 0x74, 0xb5, 0xb7, 0x42, 0x16, 0x0e, 0x38, 0xb8, 0xa5, 0xb1,
	0x35, 0xd4, 0x24, 0x7d, 0x55, 0x56, 0x73, 0x2d, 0x2b, 0x58, 0x7c, 0x8f, 0xf0, 0x3f, 0x76, 0x64,
	0x9e, 0x02, 0x42, 0x6e, 0xf0, 0xbc, 0x7b, 0x49, 0xed, 0x28, 0xca, 0x50, 0x9e, 0x32, 0xdc, 0x46,
	0x8f, 0x3b, 0x72, 0x8d, 0x53, 0x50, 0x5a, 0x3a, 0xe0, 0xda, 0xd0, 0x51, 0x86, 0xf2, 0x31, 0xeb,
	0x03, 0xb2, 0xc2, 0x58, 0x18, 0xef, 0x9a, 0x3d, 0x9a, 0x64, 0x49, 0x3e, 0x5f, 0x9d, 0x2f, 0xbb,
	0xc1, 0xe5, 0xc6, 0xf8, 0xb8, 0xc3, 0xd2, 0x80, 0x35, 0x93, 0x77, 0xf8, 0x4c, 0x4b, 0x5d, 0xf0,
	0xb2, 0xac, 0x05, 0x07, 0xbe, 0x2d, 0x25, 0x1d, 0x67, 0x28, 0x4f, 0xd8, 0x7f, 0x2d, 0xf5, 0x43,
	0x9f, 0x92, 0x5b, 0xfc, 0x37, 0x80, 0x82, 0x1b, 0x2e, 0x14, 0x7c, 0xd0, 0x3f, 0x91, 0x9a, 0x6b,
	0xa9, 0x37, 0xc7, 0xa8, 0x3d, 0xcb, 0x4a, 0x27, 0xed, 0x3b, 0x07, 0x55, 0x57, 0x74, 0x92, 0xa1,
	0x1c, 0xc5, 0xb3, 0x58, 0x9f, 0xb6, 0xa0, 0x07, 0x55, 0xaa, 0xcf, 0x06, 0x9c, 0x76, 0xe0, 0x4b,
	0x9f, 0x92, 0x4b, 0x3c, 0xdd, 0x29, 0xf7, 0x56, 0x6c, 0xf7, 0x74, 0x16, 0xf7, 0x26, 0xe1, 0x71,
	0xbd, 0x27, 0x57, 0x38, 0xad, 0x24, 0x14, 0xf6, 0x10, 0xaa, 0x34, 0x56, 0xd3, 0x4a, 0x02, 0x3b,
	0xf4, 0x1d, 0xc4, 0x0e, 0x77, 0xdd, 0xf3, 0x61, 0xbd, 0x5f, 0x7c, 0x21, 0x3c, 0x6b, 0x7f, 0x43,
	0xf8, 0x0c, 0x61, 0xfc, 0x2f, 0x77, 0xd4, 0xb8, 0x0b, 0xe3, 0x4f, 0xdc, 0x03, 0xd8, 0xb9, 0x8f,
	0x1a, 0x77, 0x61, 0xfc, 0xd0, 0x3d, 0x20, 0x43, 0xf7, 0xa4, 0x51, 0x12, 0xc6, 0x9f, 0xb8, 0x07,
	0x70, 0xe8, 0x3e, 0xee, 0xc0, 0x81, 0xfb, 0x76, 0x12, 0x2f, 0xcc, 0xfd, 0x4f, 0x00, 0x00, 0x00,
	0xff, 0xff, 0x99, 0xa2, 0x0c, 0xa3, 0x48, 0x02, 0x00, 0x00,
}
