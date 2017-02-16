/*
Copyright 2017 The Kubernetes Authors.

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

package framework

// TODO: rename file to review.go

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/resource"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/conversion"
	"k8s.io/kubernetes/pkg/labels"

	ccreviewapi "github.com/kubernetes-incubator/cluster-capacity/pkg/apis/clustercapacityreview"
)

func getMainFailReason(message string) *ccreviewapi.ClusterCapacityReviewScheduleFailReason {
	slicedMessage := strings.Split(message, "\n")
	colon := strings.Index(slicedMessage[0], ":")

	fail := &ccreviewapi.ClusterCapacityReviewScheduleFailReason{
		FailType:    slicedMessage[0][:colon],
		FailMessage: strings.Trim(slicedMessage[0][colon+1:], " "),
	}
	return fail
}

func getResourceRequest(pod *api.Pod) *ccreviewapi.Resources {
	result := ccreviewapi.Resources{
		CPU:       resource.NewMilliQuantity(0, resource.DecimalSI),
		Memory:    resource.NewQuantity(0, resource.BinarySI),
		NvidiaGPU: resource.NewMilliQuantity(0, resource.DecimalSI),
	}
	for _, container := range pod.Spec.Containers {
		for rName, rQuantity := range container.Resources.Requests {
			switch rName {
			case api.ResourceMemory:
				result.Memory.Add(rQuantity)
			case api.ResourceCPU:
				result.CPU.Add(rQuantity)
			case api.ResourceNvidiaGPU:
				result.NvidiaGPU.Add(rQuantity)
			default:
				if api.IsOpaqueIntResourceName(rName) {
					// Lazily allocate this map only if required.
					if result.OpaqueIntResources == nil {
						result.OpaqueIntResources = map[api.ResourceName]int64{}
					}
					result.OpaqueIntResources[rName] += rQuantity.Value()
				}
			}
		}
	}
	return &result
}

func parsePodsReview(templatePods []*api.Pod, status Status) []*ccreviewapi.ClusterCapacityReviewResult {
	templatesCount := len(templatePods)
	result := make([]*ccreviewapi.ClusterCapacityReviewResult, 0)

	for i := 0; i < templatesCount; i++ {
		result = append(result, &ccreviewapi.ClusterCapacityReviewResult{
			ReplicasOnNodes: make(map[string]int),
			PodName:         templatePods[i].Name,
		})
	}

	for i, pod := range status.Pods {
		nodeName := pod.Spec.NodeName
		result[i%templatesCount].ReplicasOnNodes[nodeName]++
	}

	slicedMessage := strings.Split(status.StopReason, "\n")
	if len(slicedMessage) == 1 {
		return result
	}

	slicedMessage = strings.Split(slicedMessage[1][31:], `, `)
	allReasons := make(map[string]int)
	for _, nodeReason := range slicedMessage {
		leftParenthesis := strings.LastIndex(nodeReason, `(`)

		reason := nodeReason[:leftParenthesis-1]
		replicas, _ := strconv.Atoi(nodeReason[leftParenthesis+1 : len(nodeReason)-1])
		allReasons[reason] = replicas
	}

	result[(len(status.Pods)-1)%templatesCount].FailSummary = allReasons
	return result
}

func getPodsRequirements(pods []*api.Pod) []*ccreviewapi.Requirements {
	result := make([]*ccreviewapi.Requirements, 0)
	for _, pod := range pods {
		podRequirements := &ccreviewapi.Requirements{
			PodName:       pod.Name,
			Resources:     getResourceRequest(pod),
			NodeSelectors: pod.Spec.NodeSelector,
		}
		result = append(result, podRequirements)
	}
	return result
}

func deepCopyPods(in []*api.Pod, out []api.Pod) {
	cloner := conversion.NewCloner()
	for i, pod := range in {
		api.DeepCopy_api_Pod(pod, &out[i], cloner)
	}
}

func getReviewSpec(podTemplates []*api.Pod) ccreviewapi.ClusterCapacityReviewSpec {

	podCopies := make([]api.Pod, len(podTemplates))
	deepCopyPods(podTemplates, podCopies)
	return ccreviewapi.ClusterCapacityReviewSpec{
		Templates:       podCopies,
		PodRequirements: getPodsRequirements(podTemplates),
	}
}

func getReviewStatus(pods []*api.Pod, status Status) ccreviewapi.ClusterCapacityReviewStatus {
	return ccreviewapi.ClusterCapacityReviewStatus{
		CreationTimestamp: time.Now(),
		Replicas:          int32(len(status.Pods)),
		FailReason:        getMainFailReason(status.StopReason),
		Pods:              parsePodsReview(pods, status),
	}
}

func GetReport(pods []*api.Pod, status Status) *ccreviewapi.ClusterCapacityReview {
	internalReview := &ccreviewapi.ClusterCapacityReview{
		TypeMeta: unversioned.TypeMeta{
			Kind:       ccreviewapi.Kind("ClusterCapacityReview").Kind,
			APIVersion: ccreviewapi.SchemeGroupVersion.Version,
		},
		Spec:   getReviewSpec(pods),
		Status: getReviewStatus(pods, status),
	}

	// TODO: return versioned object?
	return internalReview
}

func instancesSum(replicasOnNodes map[string]int) int {
	result := 0
	for _, v := range replicasOnNodes {
		result += v
	}
	return result
}

func clusterCapacityReviewPrettyPrint(r *ccreviewapi.ClusterCapacityReview, verbose bool) {
	if verbose {
		for _, req := range r.Spec.PodRequirements {
			fmt.Printf("%v pod requirements:\n", req.PodName)
			fmt.Printf("\t- CPU: %v\n", req.Resources.CPU.String())
			fmt.Printf("\t- Memory: %v\n", req.Resources.Memory.String())
			if !req.Resources.NvidiaGPU.IsZero() {
				fmt.Printf("\t- NvidiaGPU: %v\n", req.Resources.NvidiaGPU.String())
			}
			if req.Resources.OpaqueIntResources != nil {
				fmt.Printf("\t- OpaqueIntResources: %v\n", req.Resources.OpaqueIntResources)
			}

			if req.NodeSelectors != nil {
				fmt.Printf("\t- NodeSelector: %v\n", labels.SelectorFromSet(labels.Set(req.NodeSelectors)).String())
			}
			fmt.Printf("\n")
		}
	}

	for _, pod := range r.Status.Pods {
		fmt.Printf("The cluster can schedule %v instance(s) of the pod %v.\n", instancesSum(pod.ReplicasOnNodes), pod.PodName)
	}
	fmt.Printf("\nTermination reason: %v: %v\n", r.Status.FailReason.FailType, r.Status.FailReason.FailMessage)

	if verbose && r.Status.Replicas > 0 {
		for _, pod := range r.Status.Pods {
			if pod.FailSummary != nil {
				fmt.Printf("fit failure summary on nodes: ")
				for reason, occurence := range pod.FailSummary {
					fmt.Printf("%v (%v), ", reason, occurence)
				}
				fmt.Printf("\n")
			}
		}
		fmt.Printf("\nPod distribution among nodes:\n")
		for _, pod := range r.Status.Pods {
			fmt.Printf("%v\n", pod.PodName)
			for node, replicas := range pod.ReplicasOnNodes {
				fmt.Printf("\t- %v: %v instance(s)\n", node, replicas)
			}
		}
	}
}

func clusterCapacityReviewPrintJson(r *ccreviewapi.ClusterCapacityReview) error {
	jsoned, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("Failed to create json: %v", err)
	}
	fmt.Println(string(jsoned))
	return nil
}

func clusterCapacityReviewPrintYaml(r *ccreviewapi.ClusterCapacityReview) error {
	yamled, err := yaml.Marshal(r)
	if err != nil {
		return fmt.Errorf("Failed to create yaml: %v", err)
	}
	fmt.Print(string(yamled))
	return nil
}

func ClusterCapacityReviewPrint(r *ccreviewapi.ClusterCapacityReview, verbose bool, format string) error {
	switch format {
	case "json":
		return clusterCapacityReviewPrintJson(r)
	case "yaml":
		return clusterCapacityReviewPrintYaml(r)
	case "":
		clusterCapacityReviewPrettyPrint(r, verbose)
		return nil
	default:
		return fmt.Errorf("output format %q not recognized", format)
	}
}
