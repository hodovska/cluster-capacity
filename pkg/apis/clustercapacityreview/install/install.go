package install

import (
	"github.com/kubernetes-incubator/cluster-capacity/pkg/apis/clustercapacityreview"
	"github.com/kubernetes-incubator/cluster-capacity/pkg/apis/clustercapacityreview/v1alpha1"
	"k8s.io/kubernetes/pkg/apimachinery/announced"
)

func init() {
	if err := announced.NewGroupMetaFactory(
		&announced.GroupMetaFactoryArgs{
			GroupName: clustercapacityreview.GroupName,
			VersionPreferenceOrder: []string{v1alpha1.SchemeGroupVersion.Version},
			ImportPrefix: "github.com/kubernetes-incubator/cluster-capacity/pkg/apis/ccreview",
			AddInternalObjectsToScheme: clustercapacityreview.AddToScheme,
		},
		announced.VersionToSchemeFunc{
			v1alpha1.SchemeGroupVersion.Version: v1alpha1.AddToScheme,
		},
	).Announce().RegisterAndEnable(); err != nil {
		panic(err)
	}
}
