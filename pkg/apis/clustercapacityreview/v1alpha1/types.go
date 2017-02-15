package v1alpha1

import (
	"time"

	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/resource"
)

type ClusterCapacityReview struct {
	unversioned.TypeMeta `json:",inline"`
	Spec   ClusterCapacityReviewSpec `json:"spec"`
	Status ClusterCapacityReviewStatus `json:"status"`
}

type ClusterCapacityReviewSpec struct {
	// the pod desired for scheduling
	Templates []api.Pod `json:"templates"`

	// desired number of replicas that should be scheduled
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	PodRequirements []*Requirements `json:"podRequirements"`
}

type ClusterCapacityReviewStatus struct {
	CreationTimestamp time.Time `json:"creationTimestamp"`
	// actual number of replicas that could schedule
	Replicas int32 `json:"replicas"`

	FailReason *ClusterCapacityReviewScheduleFailReason `json:"failReason"`

	// per node information about the scheduling simulation
	Pods []*ClusterCapacityReviewResult `json:"pods"`
}

type ClusterCapacityReviewResult struct {
	PodName string `json:"podName"`
	// numbers of replicas on nodes
	ReplicasOnNodes map[string]int `json:"replicasOnNodes"`
	// reason why no more pods could schedule (if any on this node)
	// [reason]num of nodes with that reason
	FailSummary map[string]int `json:"failSummary,omitempty"`
}

type Resources struct {
	CPU                *resource.Quantity `json:"cpu,omitempty"`
	Memory             *resource.Quantity `json:"memory,omitempty"`
	NvidiaGPU          *resource.Quantity `json:"nvidiaGPU,omitempty"`
	OpaqueIntResources map[api.ResourceName]int64 `json:"opaqueIntResources,omitempty"`
}

type Requirements struct {
	PodName       string `json:"podName"`
	Resources     *Resources `json:"resources"`
	NodeSelectors map[string]string `json:"nodeSelectors,omitempty"`
}

type ClusterCapacityReviewScheduleFailReason struct {
	FailType    string `json:"failType"`
	FailMessage string `json:"failMessage"`
}
