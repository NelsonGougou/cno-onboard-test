package v1alpha1

import "k8s.io/apimachinery/pkg/runtime"

// DeepCopyObject returns a generically typed copy of an object
func (in *Environment) DeepCopyObject() runtime.Object {
	out := Environment{}
	in.DeepCopyInto(&out)

	return &out
}

// DeepCopyObject returns a generically typed copy of an object
func (in *EnvironmentList) DeepCopyObject() runtime.Object {
	out := EnvironmentList{}
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta

	if in.Items != nil {
		out.Items = make([]Environment, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}

	return &out
}
