package apiserver

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/registry/generic"
)

type apiBindingAwareCRDRESTOptionsGetter struct {
	delegate generic.RESTOptionsGetter
	crd      *apiextensionsv1.CustomResourceDefinition
}

func (t apiBindingAwareCRDRESTOptionsGetter) GetRESTOptions(resource schema.GroupResource) (generic.RESTOptions, error) {
	ret, err := t.delegate.GetRESTOptions(resource)
	if err != nil {
		return ret, err
	}

	if _, bound := t.crd.Annotations["apis.kcp.dev/bound-crd"]; bound {
		ret.ResourcePrefix = t.crd.Name
	}

	return ret, err
}
