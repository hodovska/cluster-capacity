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

import (
	"fmt"
	goruntime "runtime"
	"testing"

	"k8s.io/kubernetes/pkg/api"
	apitesting "k8s.io/kubernetes/pkg/api/testing"
	"k8s.io/kubernetes/pkg/apis/componentconfig"
	"k8s.io/kubernetes/pkg/version"
	soptions "k8s.io/kubernetes/plugin/cmd/kube-scheduler/app/options"
	"k8s.io/kubernetes/plugin/pkg/scheduler/factory"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kubernetes-incubator/cluster-capacity/pkg/framework/store"
)

func getGeneralNode(nodeName string) *api.Node {
	return &api.Node{
		ObjectMeta: api.ObjectMeta{Name: nodeName},
		Spec:       api.NodeSpec{},
		Status: api.NodeStatus{
			Conditions: []api.NodeCondition{
				{
					Type:               api.NodeOutOfDisk,
					Status:             api.ConditionFalse,
					Reason:             "KubeletHasSufficientDisk",
					Message:            fmt.Sprintf("kubelet has sufficient disk space available"),
					LastHeartbeatTime:  metav1.Time{},
					LastTransitionTime: metav1.Time{},
				},
				{
					Type:               api.NodeMemoryPressure,
					Status:             api.ConditionFalse,
					Reason:             "KubeletHasSufficientMemory",
					Message:            fmt.Sprintf("kubelet has sufficient memory available"),
					LastHeartbeatTime:  metav1.Time{},
					LastTransitionTime: metav1.Time{},
				},
				{
					Type:               api.NodeDiskPressure,
					Status:             api.ConditionFalse,
					Reason:             "KubeletHasNoDiskPressure",
					Message:            fmt.Sprintf("kubelet has no disk pressure"),
					LastHeartbeatTime:  metav1.Time{},
					LastTransitionTime: metav1.Time{},
				},
				{
					Type:               api.NodeReady,
					Status:             api.ConditionTrue,
					Reason:             "KubeletReady",
					Message:            fmt.Sprintf("kubelet is posting ready status"),
					LastHeartbeatTime:  metav1.Time{},
					LastTransitionTime: metav1.Time{},
				},
			},
			NodeInfo: api.NodeSystemInfo{
				MachineID:               "123",
				SystemUUID:              "abc",
				BootID:                  "1b3",
				KernelVersion:           "3.16.0-0.bpo.4-amd64",
				OSImage:                 "Debian GNU/Linux 7 (wheezy)",
				OperatingSystem:         goruntime.GOOS,
				Architecture:            goruntime.GOARCH,
				ContainerRuntimeVersion: "test://1.5.0",
				KubeletVersion:          version.Get().String(),
				KubeProxyVersion:        version.Get().String(),
			},
			Capacity: api.ResourceList{
				api.ResourceCPU:       *resource.NewMilliQuantity(1000, resource.DecimalSI),
				api.ResourceMemory:    *resource.NewQuantity(4E9, resource.BinarySI),
				api.ResourcePods:      *resource.NewQuantity(10, resource.DecimalSI),
				api.ResourceNvidiaGPU: *resource.NewQuantity(0, resource.DecimalSI),
			},
			Allocatable: api.ResourceList{
				api.ResourceCPU:       *resource.NewMilliQuantity(0, resource.DecimalSI),
				api.ResourceMemory:    *resource.NewQuantity(0, resource.BinarySI),
				api.ResourcePods:      *resource.NewQuantity(0, resource.DecimalSI),
				api.ResourceNvidiaGPU: *resource.NewQuantity(0, resource.DecimalSI),
			},
			Addresses: []api.NodeAddress{
				{Type: api.NodeLegacyHostIP, Address: "127.0.0.1"},
				{Type: api.NodeInternalIP, Address: "127.0.0.1"},
			},
			Images: []api.ContainerImage{},
		},
	}
}

func TestPrediction(t *testing.T) {
	// 1. create fake storage with initial data
	// - create three nodes, each node with different resources (cpu, memory)
	resourceStore := store.NewResourceStore()

	// create first node with 2 cpus and 4GB, with some resources already consumed
	node1 := getGeneralNode("test-node-1")
	node1.Status.Capacity = api.ResourceList{
		api.ResourceCPU:       *resource.NewMilliQuantity(2000, resource.DecimalSI),
		api.ResourceMemory:    *resource.NewQuantity(4E9, resource.BinarySI),
		api.ResourcePods:      *resource.NewQuantity(10, resource.DecimalSI),
		api.ResourceNvidiaGPU: *resource.NewQuantity(0, resource.DecimalSI),
	}
	node1.Status.Allocatable = api.ResourceList{
		api.ResourceCPU:       *resource.NewMilliQuantity(300, resource.DecimalSI),
		api.ResourceMemory:    *resource.NewQuantity(1E9, resource.BinarySI),
		api.ResourcePods:      *resource.NewQuantity(3, resource.DecimalSI),
		api.ResourceNvidiaGPU: *resource.NewQuantity(0, resource.DecimalSI),
	}

	if err := resourceStore.Add("nodes", runtime.Object(node1)); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// create second node with 2 cpus and 1GB, with some resources already consumed
	node2 := getGeneralNode("test-node-2")
	node2.Status.Capacity = api.ResourceList{
		api.ResourceCPU:       *resource.NewMilliQuantity(1000, resource.DecimalSI),
		api.ResourceMemory:    *resource.NewQuantity(4E9, resource.BinarySI),
		api.ResourcePods:      *resource.NewQuantity(10, resource.DecimalSI),
		api.ResourceNvidiaGPU: *resource.NewQuantity(0, resource.DecimalSI),
	}
	node2.Status.Allocatable = api.ResourceList{
		api.ResourceCPU:       *resource.NewMilliQuantity(400, resource.DecimalSI),
		api.ResourceMemory:    *resource.NewQuantity(2E9, resource.BinarySI),
		api.ResourcePods:      *resource.NewQuantity(3, resource.DecimalSI),
		api.ResourceNvidiaGPU: *resource.NewQuantity(0, resource.DecimalSI),
	}
	if err := resourceStore.Add("nodes", runtime.Object(node2)); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// create third node with 2 cpus and 4GB, with some resources already consumed
	node3 := getGeneralNode("test-node-3")
	node3.Status.Capacity = api.ResourceList{
		api.ResourceCPU:       *resource.NewMilliQuantity(2000, resource.DecimalSI),
		api.ResourceMemory:    *resource.NewQuantity(4E9, resource.BinarySI),
		api.ResourcePods:      *resource.NewQuantity(10, resource.DecimalSI),
		api.ResourceNvidiaGPU: *resource.NewQuantity(0, resource.DecimalSI),
	}
	node3.Status.Allocatable = api.ResourceList{
		api.ResourceCPU:       *resource.NewMilliQuantity(1200, resource.DecimalSI),
		api.ResourceMemory:    *resource.NewQuantity(1E9, resource.BinarySI),
		api.ResourcePods:      *resource.NewQuantity(3, resource.DecimalSI),
		api.ResourceNvidiaGPU: *resource.NewQuantity(0, resource.DecimalSI),
	}
	if err := resourceStore.Add("nodes", runtime.Object(node3)); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	namespace := &api.Namespace{ObjectMeta: api.ObjectMeta{Name: "test-node-3"}}
	if err := resourceStore.Add("namespaces", runtime.Object(namespace)); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	simulatedPod := &api.Pod{
		ObjectMeta: api.ObjectMeta{Name: "simulated-pod", Namespace: "test-node-3", ResourceVersion: "10"},
		Spec:       apitesting.DeepEqualSafePodSpec(),
	}

	limitResourceList := make(map[api.ResourceName]resource.Quantity)
	requestsResourceList := make(map[api.ResourceName]resource.Quantity)

	limitResourceList[api.ResourceCPU] = *resource.NewMilliQuantity(100, resource.DecimalSI)
	limitResourceList[api.ResourceMemory] = *resource.NewQuantity(5E6, resource.BinarySI)
	limitResourceList[api.ResourceNvidiaGPU] = *resource.NewQuantity(0, resource.DecimalSI)
	requestsResourceList[api.ResourceCPU] = *resource.NewMilliQuantity(100, resource.DecimalSI)
	requestsResourceList[api.ResourceMemory] = *resource.NewQuantity(5E6, resource.BinarySI)
	requestsResourceList[api.ResourceNvidiaGPU] = *resource.NewQuantity(0, resource.DecimalSI)

	// set pod's resource consumption
	simulatedPod.Spec.Containers = []api.Container{
		{
			Resources: api.ResourceRequirements{
				Limits:   limitResourceList,
				Requests: requestsResourceList,
			},
		},
	}

	// 2. create predictor
	// - create simple configuration file for scheduler (use the default values or from systemd env file if reasonable)
	cc, err := New(&soptions.SchedulerServer{
		KubeSchedulerConfiguration: componentconfig.KubeSchedulerConfiguration{
			SchedulerName:                  api.DefaultSchedulerName,
			HardPodAffinitySymmetricWeight: api.DefaultHardPodAffinitySymmetricWeight,
			FailureDomains:                 api.DefaultFailureDomains,
			AlgorithmProvider:              factory.DefaultProvider,
		},
		Master:     "http://localhost:8080",
		Kubeconfig: "/etc/kubernetes/kubeconfig",
	},
		simulatedPod,
		6,
		ResourceSpaceFull,
		&ApiServerOptions{},
	)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// 3. run predictor
	if err := cc.SyncWithStore(resourceStore); err != nil {
		t.Errorf("Unable to sync resources: %v", err)
	}
	if err := cc.Run(); err != nil {
		t.Errorf("Unable to run analysis: %v", err)
	}

	// TODO: modify when sequence is implemented
	for reason, replicas := range cc.Report().Status.Pods[0].ReplicasOnNodes {
		t.Logf("Reason: %v, instances: %v\n", reason, replicas)
	}

	t.Logf("Stop reason: %v\n", cc.Report().Status.FailReason)

	//4. check expected number of pods is scheduled and reflected in the resource storage
	if cc.Report().Status.FailReason.FailType != "LimitReached" {
		t.Errorf("Unexpected stop reason occured: %v, expecting: LimitReached", cc.Report().Status.FailReason.FailType)
	}
}
