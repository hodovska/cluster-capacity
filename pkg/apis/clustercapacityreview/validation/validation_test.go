package validation

import (
	"fmt"
	"github.com/kubernetes-incubator/cluster-capacity/pkg/apis/clustercapacityreview"
	"github.com/kubernetes-incubator/cluster-capacity/pkg/test"
	"k8s.io/kubernetes/pkg/api"
	"testing"
	"time"
)

func TestValidateClusterCapacityReview(t *testing.T) {
	validPodTemplates := make([]api.Pod, 0)
	for i := 0; i < 2; i++ {
		pod := test.PodExample(fmt.Sprintf("test-pod-%d", i))
		pod.Spec.Containers = []api.Container{{Name: "ctr", Image: "image", ImagePullPolicy: "IfNotPresent"}}

		validPodTemplates = append(validPodTemplates, pod)
	}

	validRequirements := make([]*clustercapacityreview.Requirements, 0)
	for i := 0; i < 2; i++ {
		requirement := &clustercapacityreview.Requirements{
			PodName:       fmt.Sprintf("test-pod-%d", i),
			NodeSelectors: map[string]string{"a": "b"},
		}
		validRequirements = append(validRequirements, requirement)
	}
	validPodResults := make([]*clustercapacityreview.ClusterCapacityReviewResult, 0)
	for i := 0; i < 2; i++ {
		result := &clustercapacityreview.ClusterCapacityReviewResult{
			PodName:         fmt.Sprintf("test-pod-%d", i),
			ReplicasOnNodes: map[string]int{"node-1": 1},
		}
		validPodResults = append(validPodResults, result)
	}

	scenarios := map[string]struct {
		isExpectedFailure bool
		review            *clustercapacityreview.ClusterCapacityReview
	}{
		"good-review": {
			isExpectedFailure: false,
			review: &clustercapacityreview.ClusterCapacityReview{
				Spec: clustercapacityreview.ClusterCapacityReviewSpec{
					Templates:       validPodTemplates,
					PodRequirements: validRequirements,
				},
				Status: clustercapacityreview.ClusterCapacityReviewStatus{
					CreationTimestamp: time.Now(),
					Replicas:          3,
					FailReason: &clustercapacityreview.ClusterCapacityReviewScheduleFailReason{
						FailType:    "foo",
						FailMessage: "bar",
					},
					Pods: validPodResults,
				},
			},
		},
		"no pod name in requirement": {
			isExpectedFailure: true,
			review: &clustercapacityreview.ClusterCapacityReview{
				Spec: clustercapacityreview.ClusterCapacityReviewSpec{
					Templates: validPodTemplates,
					PodRequirements: []*clustercapacityreview.Requirements{
						&clustercapacityreview.Requirements{
							PodName: "test-pod-1",
						},
						&clustercapacityreview.Requirements{
							NodeSelectors: map[string]string{"a": "b"},
						},
					},
				},
				Status: clustercapacityreview.ClusterCapacityReviewStatus{
					CreationTimestamp: time.Now(),
					Pods:              validPodResults,
				},
			},
		},
		"no pod name in result": {
			isExpectedFailure: true,
			review: &clustercapacityreview.ClusterCapacityReview{
				Spec: clustercapacityreview.ClusterCapacityReviewSpec{
					Templates:       validPodTemplates,
					PodRequirements: validRequirements,
				},
				Status: clustercapacityreview.ClusterCapacityReviewStatus{
					CreationTimestamp: time.Now(),
					Pods: []*clustercapacityreview.ClusterCapacityReviewResult{
						&clustercapacityreview.ClusterCapacityReviewResult{
							ReplicasOnNodes: map[string]int{"node-1": 1},
						},
						&clustercapacityreview.ClusterCapacityReviewResult{
							PodName: "test-pod",
						},
					},
				},
			},
		},
		"missing creation timestamp": {
			isExpectedFailure: true,
			review: &clustercapacityreview.ClusterCapacityReview{
				Spec: clustercapacityreview.ClusterCapacityReviewSpec{
					Templates:       validPodTemplates,
					PodRequirements: validRequirements,
				},
				Status: clustercapacityreview.ClusterCapacityReviewStatus{
					Replicas: 3,
					FailReason: &clustercapacityreview.ClusterCapacityReviewScheduleFailReason{
						FailType:    "foo",
						FailMessage: "bar",
					},
					Pods: validPodResults,
				},
			},
		},
	}

	for name, scenario := range scenarios {
		errs := ValidateClusterCapacityReview(scenario.review)
		if len(errs) == 0 && scenario.isExpectedFailure {
			t.Errorf("Unexpected success for scenario: %s", name)
		}
		if len(errs) > 0 && !scenario.isExpectedFailure {
			t.Errorf("Unexpected failure for scenario: %s - %+v", name, errs)
		}
	}
}
