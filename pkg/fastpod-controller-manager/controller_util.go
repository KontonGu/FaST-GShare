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

package fastpodcontrollermanager

import (
	"container/list"
	"fmt"
	"math"
	"path/filepath"
	"sort"
	"strconv"

	fastpodv1 "github.com/KontonGu/FaST-GShare/pkg/apis/fastgshare.caps.in.tum/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/klog/v2"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
	podv1 "k8s.io/kubernetes/pkg/api/v1/pod"
	"k8s.io/kubernetes/pkg/apis/core/helper"
	"k8s.io/kubernetes/pkg/features"
	"k8s.io/utils/integer"
)

const (
	LabelMinReplicas = "com.openfaas.scale.min"
)

var (
	//KeyFunc           = cache.DeletionHandlingMetaNamespaceKeyFunc
	podPhaseToOrdinal = map[corev1.PodPhase]int{corev1.PodPending: 0, corev1.PodUnknown: 1, corev1.PodRunning: 2}
)

// ------------------------- Pod Deletion Policy of a function (original version) -----------------------
// ActivePodsWithRanks is a sortable list of pods and a list of corresponding
// ranks which will be considered during sorting.  The two lists must have equal
// length.  After sorting, the pods will be ordered as follows, applying each
// rule in turn until one matches:
//
//  1. If only one of the pods is assigned to a node, the pod that is not
//     assigned comes before the pod that is.
//  2. If the pods' phases differ, a pending pod comes before a pod whose phase
//     is unknown, and a pod whose phase is unknown comes before a running pod.
//  3. If exactly one of the pods is ready, the pod that is not ready comes
//     before the ready pod.
//  4. If controller.kubernetes.io/pod-deletion-cost annotation is set, then
//     the pod with the lower value will come first.
//  5. If the pods' ranks differ, the pod with greater rank comes before the pod
//     with lower rank.
//  6. If both pods are ready but have not been ready for the same amount of
//     time, the pod that has been ready for a shorter amount of time comes
//     before the pod that has been ready for longer.
//  7. If one pod has a container that has restarted more than any container in
//     the other pod, the pod with the container with more restarts comes
//     before the other pod.
//  8. If the pods' creation times differ, the pod that was created more recently
//     comes before the older pod.
//
// In 6 and 8, times are compared in a logarithmic scale. This allows a level
// of randomness among equivalent Pods when sorting. If two pods have the same
// logarithmic rank, they are sorted by UUID to provide a pseudorandom order.
//
// If none of these rules matches, the second pod comes before the first pod.
//
// The intention of this ordering is to put pods that should be preferred for
// deletion first in the list.
type ActivePodsWithRanks struct {
	Pods []*corev1.Pod

	// Rank is a ranking of pods.  This ranking is used during sorting when
	// comparing two pods that are both scheduled, in the same phase, and
	// having the same ready status.
	Rank []int
	// Now is a reference timestamp for doing logarithmic timestamp comparisons.
	// If zero, comparison happens without scaling.
	Now metav1.Time
}

func (s ActivePodsWithRanks) Len() int {
	return len(s.Pods)
}

func (s ActivePodsWithRanks) Swap(i, j int) {
	s.Pods[i], s.Pods[j] = s.Pods[j], s.Pods[i]
	s.Rank[i], s.Rank[j] = s.Rank[j], s.Rank[i]
}

// Less compares two pods with corresponding ranks and returns true if the first
// one should be preferred for deletion.
func (s ActivePodsWithRanks) Less(i, j int) bool {
	// 1. Unassigned < assigned
	// If only one of the pods is unassigned, the unassigned one is smaller
	if s.Pods[i].Spec.NodeName != s.Pods[j].Spec.NodeName && (len(s.Pods[i].Spec.NodeName) == 0 || len(s.Pods[j].Spec.NodeName) == 0) {
		return len(s.Pods[i].Spec.NodeName) == 0
	}
	// 2. PodPending < PodUnknown < PodRunning
	if podPhaseToOrdinal[s.Pods[i].Status.Phase] != podPhaseToOrdinal[s.Pods[j].Status.Phase] {
		return podPhaseToOrdinal[s.Pods[i].Status.Phase] < podPhaseToOrdinal[s.Pods[j].Status.Phase]
	}
	// 3. Not ready < ready
	// If only one of the pods is not ready, the not ready one is smaller
	if podutil.IsPodReady(s.Pods[i]) != podutil.IsPodReady(s.Pods[j]) {
		return !podutil.IsPodReady(s.Pods[i])
	}

	// 4. lower pod-deletion-cost < higher pod-deletion cost
	if utilfeature.DefaultFeatureGate.Enabled(features.PodDeletionCost) {
		pi, _ := helper.GetDeletionCostFromPodAnnotations(s.Pods[i].Annotations)
		pj, _ := helper.GetDeletionCostFromPodAnnotations(s.Pods[j].Annotations)
		if pi != pj {
			return pi < pj
		}
	}

	// 5. Doubled up < not doubled up
	// If one of the two pods is on the same node as one or more additional
	// ready pods that belong to the same replicaset, whichever pod has more
	// colocated ready pods is less
	if s.Rank[i] != s.Rank[j] {
		return s.Rank[i] > s.Rank[j]
	}
	// TODO: take availability into account when we push minReadySeconds information from deployment into pods,
	//       see https://github.com/kubernetes/kubernetes/issues/22065
	// 6. Been ready for empty time < less time < more time
	// If both pods are ready, the latest ready one is smaller
	if podutil.IsPodReady(s.Pods[i]) && podutil.IsPodReady(s.Pods[j]) {
		readyTime1 := podReadyTime(s.Pods[i])
		readyTime2 := podReadyTime(s.Pods[j])
		if !readyTime1.Equal(readyTime2) {
			if !utilfeature.DefaultFeatureGate.Enabled(features.LogarithmicScaleDown) {
				return afterOrZero(readyTime1, readyTime2)
			} else {
				if s.Now.IsZero() || readyTime1.IsZero() || readyTime2.IsZero() {
					return afterOrZero(readyTime1, readyTime2)
				}
				rankDiff := logarithmicRankDiff(*readyTime1, *readyTime2, s.Now)
				if rankDiff == 0 {
					return s.Pods[i].UID < s.Pods[j].UID
				}
				return rankDiff < 0
			}
		}
	}
	// 7. Pods with containers with higher restart counts < lower restart counts
	if maxContainerRestarts(s.Pods[i]) != maxContainerRestarts(s.Pods[j]) {
		return maxContainerRestarts(s.Pods[i]) > maxContainerRestarts(s.Pods[j])
	}
	// 8. Empty creation time pods < newer pods < older pods
	if !s.Pods[i].CreationTimestamp.Equal(&s.Pods[j].CreationTimestamp) {
		if !utilfeature.DefaultFeatureGate.Enabled(features.LogarithmicScaleDown) {
			return afterOrZero(&s.Pods[i].CreationTimestamp, &s.Pods[j].CreationTimestamp)
		} else {
			if s.Now.IsZero() || s.Pods[i].CreationTimestamp.IsZero() || s.Pods[j].CreationTimestamp.IsZero() {
				return afterOrZero(&s.Pods[i].CreationTimestamp, &s.Pods[j].CreationTimestamp)
			}
			rankDiff := logarithmicRankDiff(s.Pods[i].CreationTimestamp, s.Pods[j].CreationTimestamp, s.Now)
			if rankDiff == 0 {
				return s.Pods[i].UID < s.Pods[j].UID
			}
			return rankDiff < 0
		}
	}
	return false
}

func afterOrZero(t1, t2 *metav1.Time) bool {
	if t1.Time.IsZero() || t2.Time.IsZero() {
		return t1.Time.IsZero()
	}
	return t1.After(t2.Time)
}

func podReadyTime(pod *corev1.Pod) *metav1.Time {
	if podutil.IsPodReady(pod) {
		for _, c := range pod.Status.Conditions {
			if c.Type == corev1.PodReady && c.Status == corev1.ConditionTrue {
				return &c.LastTransitionTime
			}
		}
	}
	return &metav1.Time{}
}

func logarithmicRankDiff(t1, t2, now metav1.Time) int64 {
	d1 := now.Sub(t1.Time)
	d2 := now.Sub(t2.Time)
	r1 := int64(-1)
	r2 := int64(-1)
	if d1 > 0 {
		r1 = int64(math.Log2(float64(d1)))
	}
	if d2 > 0 {
		r2 = int64(math.Log2(float64(d2)))
	}
	return r1 - r2
}

func maxContainerRestarts(pod *corev1.Pod) int {
	maxRestarts := 0
	for _, c := range pod.Status.ContainerStatuses {
		maxRestarts = integer.IntMax(maxRestarts, int(c.RestartCount))
	}
	return maxRestarts
}

// getReplicas returns the desired number of replicas for a function taking into account
// the min replicas label, HPA, the OF autoscaler and scaled to zero deployments
func getReplicas(fastpod *fastpodv1.FaSTPod) *int32 {
	var minReplicas *int32

	// extract min replicas from label if specified
	if fastpod != nil && fastpod.Labels != nil {
		lb := fastpod.Labels
		if value, exists := lb[LabelMinReplicas]; exists {
			r, err := strconv.Atoi(value)
			if err == nil && r > 0 {
				minReplicas = int32p(int32(r))
			}
		}
	}

	// extract current deployment replicas if specified
	fastpodReplicas := fastpod.Spec.Replicas

	// do not set replicas if min replicas is not set
	// and current deployment has no replicas count
	if minReplicas == nil && fastpodReplicas == nil {
		return nil
	}

	// set replicas to min if deployment has no replicas and min replicas exists
	if minReplicas != nil && fastpodReplicas == nil {
		return minReplicas
	}

	// do not override replicas when deployment is scaled to zero
	if fastpodReplicas != nil && *fastpodReplicas == 0 {
		return fastpodReplicas
	}

	// do not override replicas when min is not specified
	if minReplicas == nil && fastpodReplicas != nil {
		return fastpodReplicas
	}

	// do not override HPA or OF autoscaler replicas if the value is greater than min
	if minReplicas != nil && fastpodReplicas != nil {
		if *fastpodReplicas >= *minReplicas {
			return fastpodReplicas
		}
	}
	//minrep := *minReplicas

	//minrep = minrep - 1

	//minReplicas = &minrep
	return minReplicas
}

func int32p(i int32) *int32 {
	return &i
}

// ----------------------- The end of functions related to deletion Policy -----------------------

func getFaSTPodReplicaStatus(fastpod *fastpodv1.FaSTPod, pods []*corev1.Pod, manageReplicaErr error) fastpodv1.FaSTPodStatus {
	newStatus := fastpod.Status

	readyReplicasCnt := 0
	availableReplicaCnt := 0

	for _, pod := range pods {
		if podv1.IsPodReady(pod) {
			readyReplicasCnt++
			if podv1.IsPodAvailable(pod, 1, metav1.Now()) {
				availableReplicaCnt++
			}
		} else {
			klog.Infof("Waiting pod %s to be ready.", pod.Name)
		}
	}

	if manageReplicaErr != nil {
		var reason string
		if diff := len(pods) - int(*(fastpod.Spec.Replicas)); diff < 0 {
			reason = "FailedCreate"

		} else if diff > 0 {
			reason = "FailedDelete"
		}
		klog.Info(reason)
	}

	newStatus.Replicas = int32(len(pods))
	newStatus.AvailableReplicas = int32(availableReplicaCnt)
	newStatus.ReadyReplicas = int32(readyReplicasCnt)

	return newStatus
}

func filterInactivePods(pods []*corev1.Pod) []*corev1.Pod {
	var result []*corev1.Pod
	for _, p := range pods {
		if IsPodActive(p) {
			result = append(result, p)
		} else {
			klog.V(4).Infof("Ignoring inactive pod %v/%v in state %v, deletion time %v", p.Namespace, p.Name, p.Status.Phase, p.DeletionTimestamp)
		}
	}
	return result
}

func IsPodActive(p *corev1.Pod) bool {
	return corev1.PodSucceeded != p.Status.Phase && corev1.PodFailed != p.Status.Phase && p.DeletionTimestamp == nil
}

func IsPodHot(p *corev1.Pod) bool {
	return p.Annotations[FastGShareWarm] != "true"
}

func RandStr(n int) string {
	var letters = []rune("1234567890abcdefghijklmnopqrstuvwxyz")
	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

func FindInQueue(key string, pl *list.List) (*PodReq, bool) {
	for k := pl.Front(); k != nil; k = k.Next() {
		if k.Value.(*PodReq).Key == key {
			return k.Value.(*PodReq), true
		}
	}
	return nil, false
}

func makeLabels(fastpod *fastpodv1.FaSTPod) map[string]string {
	labels := map[string]string{
		"fast_function": fastpod.Name,
		"app":           fastpod.Name,
		"controller":    fastpod.Name,
	}

	if fastpod.Labels != nil {
		for k, v := range fastpod.Labels {
			labels[k] = v
		}
	}
	return labels
}

func getPodsRankedByRelatedPodsOnSameNode(podsToRank []*corev1.Pod) ActivePodsWithRanks {
	podsOnNode := make(map[string]int)
	for _, pod := range podsToRank {
		if IsPodActive(pod) {
			podsOnNode[pod.Spec.NodeName]++
		}
	}

	ranks := make([]int, len(podsToRank))

	for i, pod := range podsToRank {
		ranks[i] = podsOnNode[pod.Spec.NodeName]
	}
	return ActivePodsWithRanks{Pods: podsToRank, Rank: ranks, Now: metav1.Now()}
}

// Select pods from existedPods to delete based on the number of pods that should be deleted (diff)
func (ctr *Controller) getPodsToDelete(existedPods []*corev1.Pod, diff int) []*corev1.Pod {
	if diff < len(existedPods) {
		podsWithRanks := getPodsRankedByRelatedPodsOnSameNode(existedPods)
		sort.Sort(podsWithRanks)
		//
		return existedPods[:diff]
	} //else if diff > len(filteredPods)
	return nil
}

type FastpodProbes struct {
	Liveness  *corev1.Probe
	Readiness *corev1.Probe
}

func makeSimpleProbes() (*FastpodProbes, error) {
	var handler corev1.ProbeHandler
	path := filepath.Join("/home/app/probe/", "probe.sh")
	handler = corev1.ProbeHandler{
		Exec: &corev1.ExecAction{
			Command: []string{"/bin/bash", path},
		},
	}
	probes := FastpodProbes{}
	probes.Readiness = &corev1.Probe{
		ProbeHandler:        handler,
		InitialDelaySeconds: int32(2),
		TimeoutSeconds:      int32(1),
		PeriodSeconds:       int32(10),
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}

	probes.Liveness = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			Exec: &corev1.ExecAction{
				Command: []string{"cat", filepath.Join("/tmp/", ".lock")},
			},
		},
		InitialDelaySeconds: int32(2),
		TimeoutSeconds:      int32(1),
		PeriodSeconds:       int32(10),
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}

	return &probes, nil
}

// PodKey returns a key unique to the given pod within a cluster.
// It's used so we consistently use the same key scheme in this module.
// It does exactly what cache.MetaNamespaceKeyFunc would have done
// except there's not possibility for error since we know the exact type.
func PodKey(pod *corev1.Pod) string {
	return fmt.Sprintf("%v/%v", pod.Namespace, pod.Name)
}

func getPodKeys(pods []*corev1.Pod) []string {
	podKeys := make([]string, 0, len(pods))
	for _, pod := range pods {
		podKeys = append(podKeys, PodKey(pod))
	}
	return podKeys
}
