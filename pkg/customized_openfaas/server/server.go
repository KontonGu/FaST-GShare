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
	"net/http"
	"os"

	clientset "github.com/KontonGu/FaST-GShare/pkg/client/clientset/versioned"
	listers "github.com/KontonGu/FaST-GShare/pkg/client/listers/fastgshare.caps.in.tum/v1"
	"github.com/KontonGu/FaST-GShare/pkg/config"
	"github.com/KontonGu/FaST-GShare/pkg/customized_openfaas/k8s"
	bootstrap "github.com/openfaas/faas-provider"
	"github.com/openfaas/faas-provider/logs"
	"github.com/openfaas/faas-provider/types"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	coreinformerv1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	klog "k8s.io/klog/v2"
)

func New(fastclient clientset.Interface,
	kubeclient kubernetes.Interface,
	podInformer coreinformerv1.PodInformer,
	fastpodLister listers.FaSTPodLister,
	clusterRole bool,
	fastpodLookup *k8s.FunctionLookup,
	cfg config.BootstrapConfig) *Server {

	fastpodNamespace := "fast-gshare-fn"

	if namespace, exists := os.LookupEnv("function_namespace"); exists {
		fastpodNamespace = namespace
	}

	pprof := "false"

	if val, exists := os.LookupEnv("pprof"); exists {
		pprof = val
	}
	klog.Infof("pprof: %v.", pprof)

	bootstrapConfig := types.FaaSConfig{
		ReadTimeout:  cfg.FaaSConfig.ReadTimeout,
		WriteTimeout: cfg.FaaSConfig.WriteTimeout,
		TCPPort:      cfg.FaaSConfig.TCPPort,
		EnableHealth: true,
	}

	bootstrapHandlers := types.FaaSHandlers{
		FunctionProxy:  NewHandlerFunc(bootstrapConfig, fastpodLookup, fastclient),
		DeleteHandler:  makeDeleteHandler(fastpodNamespace, fastclient),
		DeployHandler:  makeApplyHandler(fastpodNamespace, fastclient),
		FunctionReader: makeListHandler(fastpodNamespace, fastclient, fastpodLister),
		ReplicaReader:  makeReplicaReader(fastpodNamespace, fastclient, fastpodLister),
		ReplicaUpdater: makeReplicaHandler(fastpodNamespace, fastclient, fastpodLookup),
		UpdateHandler:  makeHealthReader(),
		HealthHandler:  makeHealthReader(),
		InfoHandler:    makeInfoHandler(),
		//SecretHandler: ,
		LogHandler:           logs.NewLogHandlerFunc(k8s.NewLogRequestor(kubeclient, fastpodNamespace), bootstrapConfig.WriteTimeout),
		ListNamespaceHandler: MakeNamespacesLister(fastpodNamespace, clusterRole, kubeclient),
	}

	if pprof == "true" {
		bootstrap.Router().PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
	}

	bootstrap.Router().Path("/metrics").Handler(promhttp.Handler())
	klog.Infof("Using namespace '%s'", fastpodNamespace)

	return &Server{
		BootstrapConfig:   &bootstrapConfig,
		BootstrapHandlers: &bootstrapHandlers,
	}
}

type Server struct {
	BootstrapHandlers *types.FaaSHandlers
	BootstrapConfig   *types.FaaSConfig
}

func (s *Server) Start() {
	klog.Infof("Starting HTTP server on port %d", *s.BootstrapConfig.TCPPort)
	bootstrap.Serve(s.BootstrapHandlers, s.BootstrapConfig)
}
