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

package config

import (
	"flag"
	"strconv"
	"strings"

	"github.com/golang/glog"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var config poseidonConfig

type poseidonConfig struct {
	SchedulerName      string `json:"schedulerName,omitempty"`
	FirmamentAddress   string `json:"firmamentAddress,omitempty"`
	KubeConfig         string `json:"kubeConfig,omitempty"`
	KubeVersion        string `json:"kubeVersion,omitempty"`
	StatsServerAddress string `json:"statsServerAddress,omitempty"`
	SchedulingInterval int    `json:"schedulingInterval,omitempty"`
	FirmamentPort      string `json:"firmamentPort,omitempty"`
	ConfigPath         string `json:"configPath,omitempty"`
}

// GetSchedulerName returns the SchedulerName from config
func GetSchedulerName() string {
	return config.SchedulerName
}

// GetFirmamentAddress returns the FirmamentAddress from config
func GetFirmamentAddress() string {
	// join the firmament address and port with a colon separator
	// Passing the firmament address with port and colon separator throws an error
	// for conversion from yaml to json
	values := []string{config.FirmamentAddress, config.FirmamentPort}
	return strings.Join(values, ":")
}

// GetKubeConfig returns the KubeConfig from config
func GetKubeConfig() string {
	return config.KubeConfig
}

// GetKubeVersion returns the KubeMajor and Minor version from the config
func GetKubeVersion() (int, int) {
	kubeVer := strings.Split(config.KubeVersion, ".")
	if len(kubeVer) < 2 {
		glog.Fatalf("Incorrect content in --kubeVersion %s, kubeVersion should be in the format of X.Y", config.KubeVersion)
	}
	kubeMajorVer, err := strconv.Atoi(kubeVer[0])
	if err != nil {
		glog.Fatalf("Incorrect content in --kubeVersion %s, kubeVersion should be in the format of X.Y and X should be an non-negative integer", config.KubeVersion)
	}
	kubeMinorVer, err := strconv.Atoi(kubeVer[1])
	if err != nil {
		glog.Fatalf("Incorrect content in --kubeVersion %s, kubeVersion should be in the format of X.Y and Y should be an non-negative integer", config.KubeVersion)
	}
	return kubeMajorVer, kubeMinorVer
}

// GetStatsServerAddress returns the StatsServerAddress from the config
// TODO(shiv): We need to have separate port and IP for stats server too like firmament address amd port.
// This separation is required when passing address as command line flags in deployment yaml,
// the IP:PORT in the yaml throws an error.
func GetStatsServerAddress() string {
	return config.StatsServerAddress
}

// GetSchedulingInterval return the scheduling interval from config
func GetSchedulingInterval() int {
	return config.SchedulingInterval
}

// GetConfigPath returns the config path from  config
func GetConfigPath() string {
	return config.ConfigPath
}

// ReadFromConfigFile to read from yaml,json,toml etc poseidonConfig file
// Note:
//  The poseidonConfig values will be overwritten if flag for the same key are present
func ReadFromConfigFile() {
	viper.AddConfigPath(".")
	viper.AddConfigPath(config.ConfigPath)
	viper.SetConfigName("poseidon_config")
	err := viper.ReadInConfig()
	if err != nil {
		glog.Warning(err, "unable to read poseidon_config, using command flags/default values")
		return
	}
	err = viper.Unmarshal(&config)
	if err != nil {
		glog.Fatal("unmarshal poseidon_config file failed", err)
	}
	glog.Info("ReadFromConfigFile", config)
}

// ReadFromCommandLineFlags reads command line flags and these will override poseidonConfig file flags.
func ReadFromCommandLineFlags() {
	pflag.StringVar(&config.SchedulerName, "schedulerName", "poseidon", "The scheduler name with which pods are labeled")
	pflag.StringVar(&config.FirmamentAddress, "firmamentAddress", "firmament-service.kube-system", "Firmament scheduler service address")
	pflag.StringVar(&config.FirmamentPort, "firmamentPort", "9090", "Firmament scheduler service port")
	pflag.StringVar(&config.KubeConfig, "kubeConfig", "kubeconfig.cfg", "Path to the kubeconfig file")
	pflag.StringVar(&config.KubeVersion, "kubeVersion", "1.6", "Kubernetes version")
	pflag.StringVar(&config.StatsServerAddress, "statsServerAddress", "0.0.0.0:9091", "Address on which the stats server listens")
	pflag.IntVar(&config.SchedulingInterval, "schedulingInterval", 10, "Time between scheduler runs (in seconds)")
	pflag.StringVar(&config.ConfigPath, "configPath", ".",
		"The path to the config file (i.e poseidon_cfg) without filename or extension, supported extensions/formats are Yaml, Json")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	// This is required to make flag package suppress the below error msg
	// ERROR: logging before flag.Parse:
	// please refer https://github.com/kubernetes/kubernetes/issues/17162
	flag.CommandLine.Parse([]string{})

	viper.BindPFlags(pflag.CommandLine)
	glog.Info("ReadFromCommandLineFlags", config)
}

func init() {
	ReadFromCommandLineFlags()
	ReadFromConfigFile()
}
