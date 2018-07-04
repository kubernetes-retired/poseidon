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
	"context"
	"encoding/json"
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
	PathHealth  = "/healthz"
)

// generateMetricsHandler generates metrics handlers.
func generateMetricsHandler() map[string]http.Handler {
	metrics.Register()
	m := make(map[string]http.Handler)
	m[pathMetrics] = promhttp.Handler()
	return m
}

// generateHealthzHandler generates healthz handlers.
func generateHealthzHandler(fc firmament.FirmamentSchedulerClient) map[string]http.Handler {
	m := make(map[string]http.Handler)
	m[PathHealth] = newHealthzHandler(func() Health { return checkHealth(fc) })
	return m
}

// newHealthHandler handles '/healthz' requests.
func newHealthzHandler(hfunc func() Health) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		h := hfunc()
		d, err := json.Marshal(h)
		if err != nil {
			glog.Errorf("Marshal failed, err: %v", err)
			return
		}
		if h.Health != "true" {
			http.Error(w, string(d), http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(d)
	}
}

type Health struct {
	Health string `json:"health"`
}

// checkHealth checks the status of firmament service and generate the status of poseidon
func checkHealth(fc firmament.FirmamentSchedulerClient) Health {
	h := Health{Health: "false"}
	res, err := fc.Check(context.Background(), &firmament.HealthCheckRequest{})
	if err != nil {
		glog.Errorf("checkHealth: health check err: %v", err)
		return h
	}

	if res.GetStatus() == firmament.ServingStatus_SERVING {
		h.Health = "true"
	}
	return h
}

// buildAddrMap adds handler map to addrMap
func buildAddrMap(addr string,
	handlerMap map[string]http.Handler,
	addrMap map[string][]map[string]http.Handler) {
	_, ok := addrMap[addr]
	if ok {
		addrMap[addr] = append(addrMap[addr], handlerMap)
	} else {
		hList := []map[string]http.Handler{}
		hList = append(hList, handlerMap)
		addrMap[addr] = hList
	}
}

// Serve starts the http service for metrics/healthz/pprof
func Serve(fc firmament.FirmamentSchedulerClient) {
	cfg := config.GetConfig()
	// addrMap is a map to store the port addrs, key is the port name and value is the ip:port
	addrMap := make(map[string][]map[string]http.Handler)
	addrMap[cfg.MetricsBindAddress] = append(addrMap[cfg.MetricsBindAddress], generateMetricsHandler())

	if cfg.EnablePprof {
		glog.Infof("pprof is enabled under %s", config.GetPprofAddress()+debugutil.HTTPPrefixPProf)
		go debugutil.RuntimeStack()
		buildAddrMap(cfg.PprofAddress, debugutil.PProfHandlers(), addrMap)
	}
	// add healthz handler map to addrMap
	buildAddrMap(cfg.HealthCheckAddress, generateHealthzHandler(fc), addrMap)

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
