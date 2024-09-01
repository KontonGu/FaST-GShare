/*
Copyright 2024 FaST-GShare Authors, KontonGu (Jianfeng Gu), et. al.
@Techinical University of Munich, CAPS Cloud Team

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

package server

import (
	"log"
	"net"
	"net/http"
	"reflect"
	"strings"
	"time"

	clientset "github.com/KontonGu/FaST-GShare/pkg/client/clientset/versioned"
	"github.com/KontonGu/FaST-GShare/pkg/customized_openfaas/k8s"
	"github.com/gorilla/mux"
	"github.com/openfaas/faas-provider/httputil"
	"github.com/openfaas/faas-provider/types"
	klog "k8s.io/klog/v2"
)

const (
	watchdogPort       = "8080"
	defaultContentType = "text/plain"

	maxidleconns        = 0
	maxidleConnsPerHost = 10
	idleConnTimeout     = 400 * time.Millisecond
)

// NewHandlerFunc creates a standard http.HandlerFunc to proxy function requests.
// The returned http.HandlerFunc will ensure:
//
// 	- proper proxy request timeouts
// 	- proxy requests for GET, POST, PATCH, PUT, and DELETE
// 	- path parsing including support for extracing the function name, sub-paths, and query paremeters
// 	- passing and setting the `X-Forwarded-Host` and `X-Forwarded-For` headers
// 	- logging errors and proxy request timing to stdout
//
// Note that this will panic if `resolver` is nil.

func NewHandlerFunc(config types.FaaSConfig, resolver *k8s.FunctionLookup, fastclient clientset.Interface) http.HandlerFunc {
	if resolver == nil {
		panic("NewHandlerFunc: empty proxy handler resolver, cannot be nil")
	}

	proxyClient := NewProxyClientFromConfig(config)

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			defer r.Body.Close()
		}

		switch r.Method {
		case http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodGet,
			http.MethodOptions,
			http.MethodHead:
			proxyRequest(w, r, proxyClient, resolver, fastclient)

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

// NewProxyClientFromConfig creates a new http.Client designed for proxying requests and enforcing
// certain minimum configuration values.
func NewProxyClientFromConfig(config types.FaaSConfig) *http.Client {
	return NewProxyClient(config.GetReadTimeout(), config.GetMaxIdleConns(), config.GetMaxIdleConnsPerHost())
}

// NewProxyClient creates a new http.Client designed for proxying requests, this is exposed as a
// convenience method for internal or advanced uses. Most people should use NewProxyClientFromConfig.
func NewProxyClient(timeout time.Duration, maxIdleConns int, maxIdleConnsPerHost int) *http.Client {
	return &http.Client{
		// these Transport values ensure that the http Client will eventually timeout and prevents
		// infinite retries. The default http.Client configure these timeouts.  The specific
		// values tuned via performance testing/benchmarking
		//
		// Additional context can be found at
		// - https://medium.com/@nate510/don-t-use-go-s-default-http-client-4804cb19f779
		// - https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/
		//
		// Additionally, these overrides for the default client enable re-use of connections and prevent
		// CoreDNS from rate limiting under high traffic
		//
		// See also two similar projects where this value was updated:
		// https://github.com/prometheus/prometheus/pull/3592
		// https://github.com/minio/minio/pull/5860
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   timeout,
				KeepAlive: 1 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          maxidleconns,
			MaxIdleConnsPerHost:   maxidleConnsPerHost,
			IdleConnTimeout:       idleConnTimeout,
			TLSHandshakeTimeout:   5 * time.Second,
			ExpectContinueTimeout: 1500 * time.Millisecond,
		},
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

// proxyRequest handles the actual resolution of and then request to the function service.
func proxyRequest(w http.ResponseWriter, originalReq *http.Request, proxyClient *http.Client, resolver *k8s.FunctionLookup,
	fastclient clientset.Interface) {

	pathVars := mux.Vars(originalReq)
	functionName := pathVars["name"]
	if functionName == "" {
		httputil.Errorf(w, http.StatusBadRequest, "Provide function name in the request path")
		return
	}
	log.Printf("originalReq: %s", originalReq.URL.String())
	slice := strings.Split(originalReq.URL.String(), "/")
	suffix := ""
	log.Printf("after slice: %s", slice)
	if len(slice) > 3 {
		suffix = slice[3]
	}
	log.Printf("request suffix with: %s", suffix)

	start := time.Now()

	argA := pathVars["params"]
	taskName := "index." + functionName
	response, err := resolver.CeleryClient.Delay(taskName, argA)

	if err != nil {
		//log.Printf("error with proxy request to: %s, %s\n", proxyReq.URL.String(), err.Error())
		httputil.Errorf(w, http.StatusInternalServerError, "Can't reach service for: %s.", functionName)
		return
	}

	result, err := response.Get(2 * time.Second)
	seconds := time.Since(start)
	w.Header().Set("Content-Type", defaultContentType)
	if err != nil {
		//go resolver.Update(seconds, functionName, podName, fastclient, true)
		w.WriteHeader(502)
	} else {
		w.WriteHeader(200)
	}

	log.Printf("%s with task id %s took %f seconds\n", functionName, response.TaskID, seconds.Seconds())

	if result != nil {
		//io.Copy(w, result.)
		klog.Infof("result: %+v of type %+v", result, reflect.TypeOf(result))
	}

}
