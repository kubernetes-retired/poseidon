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
// source: whare_map_stats.proto
// DO NOT EDIT!

package firmament

import (
	"fmt"
	"math"

	"github.com/golang/protobuf/proto"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type WhareMapStats struct {
	NumIdle    uint64 `protobuf:"varint,1,opt,name=num_idle,json=numIdle" json:"num_idle,omitempty"`
	NumDevils  uint64 `protobuf:"varint,2,opt,name=num_devils,json=numDevils" json:"num_devils,omitempty"`
	NumRabbits uint64 `protobuf:"varint,3,opt,name=num_rabbits,json=numRabbits" json:"num_rabbits,omitempty"`
	NumSheep   uint64 `protobuf:"varint,4,opt,name=num_sheep,json=numSheep" json:"num_sheep,omitempty"`
	NumTurtles uint64 `protobuf:"varint,5,opt,name=num_turtles,json=numTurtles" json:"num_turtles,omitempty"`
}

func (m *WhareMapStats) Reset()                    { *m = WhareMapStats{} }
func (m *WhareMapStats) String() string            { return proto.CompactTextString(m) }
func (*WhareMapStats) ProtoMessage()               {}
func (*WhareMapStats) Descriptor() ([]byte, []int) { return fileDescriptor14, []int{0} }

func (m *WhareMapStats) GetNumIdle() uint64 {
	if m != nil {
		return m.NumIdle
	}
	return 0
}

func (m *WhareMapStats) GetNumDevils() uint64 {
	if m != nil {
		return m.NumDevils
	}
	return 0
}

func (m *WhareMapStats) GetNumRabbits() uint64 {
	if m != nil {
		return m.NumRabbits
	}
	return 0
}

func (m *WhareMapStats) GetNumSheep() uint64 {
	if m != nil {
		return m.NumSheep
	}
	return 0
}

func (m *WhareMapStats) GetNumTurtles() uint64 {
	if m != nil {
		return m.NumTurtles
	}
	return 0
}

func init() {
	proto.RegisterType((*WhareMapStats)(nil), "firmament.WhareMapStats")
}

func init() { proto.RegisterFile("whare_map_stats.proto", fileDescriptor14) }

var fileDescriptor14 = []byte{
	// 184 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x3c, 0x8e, 0xb1, 0xaa, 0xc2, 0x30,
	0x14, 0x86, 0xe9, 0xbd, 0xbd, 0x57, 0x1b, 0x71, 0x09, 0x08, 0x11, 0x11, 0xc5, 0xc9, 0xc9, 0xc5,
	0x57, 0x70, 0x71, 0x70, 0x69, 0x05, 0xc7, 0x90, 0xd2, 0x23, 0x0d, 0x24, 0x69, 0x48, 0x4e, 0xf4,
	0x95, 0x7c, 0x4c, 0x39, 0x29, 0x3a, 0xe6, 0xff, 0xbe, 0x8f, 0x1c, 0xb6, 0x78, 0xf6, 0x2a, 0x80,
	0xb4, 0xca, 0xcb, 0x88, 0x0a, 0xe3, 0xc1, 0x87, 0x01, 0x07, 0x5e, 0xdd, 0x75, 0xb0, 0xca, 0x82,
	0xc3, 0xdd, 0xab, 0x60, 0xf3, 0x1b, 0x49, 0x17, 0xe5, 0x1b, 0x52, 0xf8, 0x92, 0x4d, 0x5d, 0xb2,
	0x52, 0x77, 0x06, 0x44, 0xb1, 0x2d, 0xf6, 0x65, 0x3d, 0x71, 0xc9, 0x9e, 0x3b, 0x03, 0x7c, 0xcd,
	0x18, 0xa1, 0x0e, 0x1e, 0xda, 0x44, 0xf1, 0x93, 0x61, 0xe5, 0x92, 0x3d, 0xe5, 0x81, 0x6f, 0xd8,
	0x8c, 0x70, 0x50, 0x6d, 0xab, 0x31, 0x8a, 0xdf, 0xcc, 0xa9, 0xa8, 0xc7, 0x85, 0xaf, 0x18, 0xd9,
	0x32, 0xf6, 0x00, 0x5e, 0x94, 0x19, 0xd3, 0x5f, 0x0d, 0xbd, 0x3f, 0x35, 0xa6, 0x80, 0x06, 0xa2,
	0xf8, 0xfb, 0xd6, 0xd7, 0x71, 0x69, 0xff, 0xf3, 0xf1, 0xc7, 0x77, 0x00, 0x00, 0x00, 0xff, 0xff,
	0xe3, 0x72, 0xcd, 0xaa, 0xd5, 0x00, 0x00, 0x00,
}
