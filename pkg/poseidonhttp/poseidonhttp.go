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

package poseidonhttp

import (
	"net/http"

	"github.com/golang/glog"
	"github.com/kubernetes-sigs/poseidon/pkg/config"
	"github.com/kubernetes-sigs/poseidon/pkg/debugutil"
	"github.com/kubernetes-sigs/poseidon/pkg/firmament"
	"github.com/kubernetes-sigs/poseidon/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	pathMetrics = "/metrics"
	PathHealth  = "/health"
)

// generateMetricsHandler generates metrics handlers.
func generateMetricsHandler() map[string]http.Handler {
	metrics.Register()
	m := make(map[string]http.Handler)
	m[pathMetrics] = promhttp.Handler()
	return m
}

// TODO(jiaxuanzhou): generateHealthzHandler() map[string]http.Handler
// func generateHealthzHandler() map[string]http.Handler()

// Serve starts the http service for metrics/healthz/pprof
func Serve(fc *firmament.FirmamentSchedulerClient) {
	cfg := config.GetConfig()
	// addrMap is a map to store the port addrs, key is the port name and value is the ip:port
	addrMap := make(map[string][]map[string]http.Handler)
	addrMap[cfg.MetricsBindAddress] = append(addrMap[cfg.MetricsBindAddress], generateMetricsHandler())

	if cfg.EnablePprof {
		glog.Infof("pprof is enabled under %s", config.GetPprofAddress()+debugutil.HTTPPrefixPProf)
		go debugutil.RuntimeStack()
		_, ok := addrMap[cfg.PprofAddress]
		if ok {
			addrMap[cfg.MetricsBindAddress] = append(addrMap[cfg.MetricsBindAddress], debugutil.PProfHandlers())
		} else {
			handlerMap := []map[string]http.Handler{}
			handlerMap = append(handlerMap, debugutil.PProfHandlers())
			addrMap[cfg.PprofAddress] = handlerMap
		}
	}
	// TODO: add healthz handler map to addrMap

	// start http services
	for addr, handlersList := range addrMap {
		go startHttpServices(addr, handlersList)
	}
}

// startHttpServices register handlers and start port services
func startHttpServices(addr string, hList []map[string]http.Handler) {
	mux := http.NewServeMux()
	for _, hMap := range hList {
		for p, h := range hMap {
			mux.Handle(p, h)
		}
	}
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	glog.Fatal(server.ListenAndServe())
}
