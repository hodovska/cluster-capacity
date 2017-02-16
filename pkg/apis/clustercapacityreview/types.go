package clustercapacityreview

import (
	"time"

	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/resource"
)

type ClusterCapacityReview struct {
	unversioned.TypeMeta
	Spec   ClusterCapacityReviewSpec
	Status ClusterCapacityReviewStatus
}

type ClusterCapacityReviewSpec struct {
	// the pod desired for scheduling
	Templates []api.Pod

	// desired number of replicas that should be scheduled
	// +optional
	Replicas int32

	PodRequirements []*Requirements
}

type ClusterCapacityReviewStatus struct {
	CreationTimestamp time.Time
	// actual number of replicas that could schedule
	Replicas int32

	FailReason *ClusterCapacityReviewScheduleFailReason

	// per node information about the scheduling simulation
	Pods []*ClusterCapacityReviewResult
}

type ClusterCapacityReviewResult struct {
	PodName string
	// numbers of replicas on nodes
	ReplicasOnNodes map[string]int
	// reason why no more pods could schedule (if any on this node)
	// [reason]num of nodes with that reason
	FailSummary map[string]int
}

type Resources struct {
	CPU                *resource.Quantity
	Memory             *resource.Quantity
	NvidiaGPU          *resource.Quantity
	OpaqueIntResources map[api.ResourceName]int64
}

type Requirements struct {
	PodName       string
	Resources     *Resources
	NodeSelectors map[string]string
}

type ClusterCapacityReviewScheduleFailReason struct {
	FailType    string
	FailMessage string
}
