/*
Copyright 2024 FaST-GShare Authors, KontonGu (Jianfeng Gu), et. al.
@Techinical University of Munich, CAPS Cloud Team

The code structure is modified based on k8s/proxy.go in the OpenFaaS project:
https://github.com/openfaas/faas-netes/blob/master/pkg/k8s/proxy.go

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

package k8s

import (
	"sync"
	"time"

	fastlisters "github.com/KontonGu/FaST-GShare/pkg/client/listers/fastgshare.caps.in.tum/v1"
	"github.com/gocelery/gocelery"
	"github.com/gomodule/redigo/redis"
	gcache "github.com/patrickmn/go-cache"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corelister "k8s.io/client-go/listers/core/v1"
	"k8s.io/klog/v2"
)

// watchdogPort for the OpenFaaS function watchdog
const watchdogPort = 8080

// target request per pod
const target = "com.openfaas.scale.target"

type PodsWithInfos struct {
	//Pods     []PodInfo
	Pods     []*corev1.Pod
	podInfos *gcache.Cache

	Now metav1.Time
}

func (s PodsWithInfos) Len() int {
	return len(s.Pods)
}

func (s PodsWithInfos) Swap(i, j int) {
	s.Pods[i], s.Pods[j] = s.Pods[j], s.Pods[i]
}

func (s PodsWithInfos) Less(i, j int) bool {
	name_i := s.Pods[i].Name
	name_j := s.Pods[j].Name

	//if a pod is unsigned, then the unsigned one is smaller
	info1, ok := s.podInfos.Get(name_i)
	info2, ok2 := s.podInfos.Get(name_j)

	if !ok || !ok2 {
		return ok
	}

	if info1.(*PodInfo).Timeout || info2.(*PodInfo).Timeout {
		return info1.(*PodInfo).Timeout
	}

	if info1.(*PodInfo).PossiTimeout || info2.(*PodInfo).PossiTimeout {
		return info1.(*PodInfo).PossiTimeout
	}

	//rate smaller < larger rate
	/*
		if s.podInfos[name_i].RateChange >= s.podInfos[name_j].RateChange {
			return s.podInfos[name_i].Rate < s.podInfos[name_j].Rate
		}
	*/
	if info1.(*PodInfo).Rate == 0 || info2.(*PodInfo).Rate == 0 {
		return (info1.(*PodInfo).Rate == 0)
	}

	/*
		if s.podInfos[name_i].Rate == s.podInfos[name_j].Rate{
			return s.podInfos[name_i].LastInvoke < s.podInfos[name_j].LastInvoke
		}

	*/

	return info1.(*PodInfo).Rate < info2.(*PodInfo).Rate

	//return s.podInfos[name_i].RateChange < s.podInfos[name_j].RateChange
}

func NewFunctionLookup(ns string, podLister corelister.PodLister, fastpodLister fastlisters.FaSTPodLister, db *gcache.Cache) *FunctionLookup {
	klog.Infof("Connecting to Redis...")
	redisPool := &redis.Pool{
		MaxActive:       0,
		MaxIdle:         10,
		IdleTimeout:     0,
		MaxConnLifetime: 0,

		Dial: func() (redis.Conn, error) {
			c, err := redis.DialURL("redis://redis.redis.svc.cluster.local:6379")
			if err != nil {
				klog.Infof("Error connecting to Redis...")
				return nil, err
			}
			klog.Infof("Connected to redis...")
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
	// initialize celery client
	cli, _ := gocelery.NewCeleryClient(
		gocelery.NewRedisBroker(redisPool),
		&gocelery.RedisCeleryBackend{Pool: redisPool},
		5,
	)

	klog.Infof("Initialize celery client...")
	return &FunctionLookup{
		DefaultNamespace: ns,
		//EndpointLister:   lister,
		FastpodLister: fastpodLister,
		PodLister:     podLister,
		//Listers:    map[string]fastLister{},
		//FastInfos: fastpodInfos,
		//lock:     sync.RWMutex{},
		//DB: db,
		Database:     db,
		CeleryClient: cli,
	}

}

type FunctionLookup struct {
	DefaultNamespace string
	//EndpointLister   corelister.EndpointsLister
	//endpoint lister may not needed for custom version
	FastpodLister fastlisters.FaSTPodLister
	PodLister     corelister.PodLister
	//Listers    map[string]fastLister

	RateRep bool

	//Service bool

	//emitter goka.Emitter
	lock sync.RWMutex

	Database *gcache.Cache
	//DB *buntdb.DB
	//redispool redis.Pool
	CeleryClient *gocelery.CeleryClient
}

func (l *FunctionLookup) AddFunc(funcname string) {
	//TODO
	fstpcache := gcache.New(5*time.Minute, 10*time.Minute)
	klog.Infof("DEBUG: initializing, FaSTPod info %s", funcname)
	l.Database.Set(funcname, fstpcache, gcache.NoExpiration)

}

func (l *FunctionLookup) DeleteFunction(name string) {
	//l.lock
	l.Database.Delete(name)
}

func (l *FunctionLookup) DeletePodInfo(funcName string, podName string) {
	if fastcache, found := l.Database.Get(funcName); found {
		klog.Infof("Deleting pod %s info of fastpod = %s", podName, funcName)
		fastcache.(gcache.Cache).Delete(podName)
		klog.Infof("Deleting finished pod %s info of fastpod = %s", podName, funcName)
	}
}
