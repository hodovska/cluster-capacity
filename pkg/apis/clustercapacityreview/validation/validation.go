package validation

import (
	"github.com/kubernetes-incubator/cluster-capacity/pkg/apis/clustercapacityreview"
	"k8s.io/kubernetes/pkg/util/validation/field"
	"k8s.io/kubernetes/pkg/api/validation"
)

func validateClusterCapacityReviewSpec(spec *clustercapacityreview.ClusterCapacityReviewSpec) field.ErrorList {
	allErrs := field.ErrorList{}

	for _, pod := range spec.Templates {
		allErrs = append(allErrs, validation.ValidatePod(&pod)...)
	}

	reqFldPath := field.NewPath("spec", "podRequirements")
	for i, requirement := range spec.PodRequirements {
		idxPath := reqFldPath.Index(i)
		if len(requirement.PodName) == 0 {
			allErrs = append(allErrs, field.Required(idxPath.Child("podName"), "Pod requirements must be associated with pod name"))
		}
	}
	return allErrs
}

func validateClusterCapacityReviewStatus(status *clustercapacityreview.ClusterCapacityReviewStatus) field.ErrorList {
	allErrs := field.ErrorList{}
	fldPath := field.NewPath("status")

	if status.CreationTimestamp.IsZero() {
		allErrs = append(allErrs, field.Required(fldPath.Child("creationTimestamp"), "Missing creation timestamp"))
	}

	for i, result := range status.Pods {
		idxPath := fldPath.Child("pods").Index(i)
		allErrs = append(allErrs, validateClusterCapacityReviewResult(result, idxPath)...)
	}
	return allErrs
}

func validateClusterCapacityReviewResult(result *clustercapacityreview.ClusterCapacityReviewResult, fldpath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if len(result.PodName) == 0 {
		allErrs = append(allErrs, field.Required(fldpath.Child("podName"), "Pod name must be specified for review result"))
	}
	return allErrs
}

func ValidateClusterCapacityReview(review *clustercapacityreview.ClusterCapacityReview) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs, validateClusterCapacityReviewSpec(&review.Spec)...)
	allErrs = append(allErrs, validateClusterCapacityReviewStatus(&review.Status)...)

	return allErrs
}
