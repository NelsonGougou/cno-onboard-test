package environment

import (
	"context"
	"reflect"
	"testing"

	onboardingv1alpha1 "gitlab.beopenit.com/cloud/onboarding-operator-kubernetes/pkg/apis/onboarding/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var (
	name        = "environment"
	projectname = "project1"
	storage     = "10Gi"
	uidUser     = "test-uid"
	labels      = map[string]string{
		"uid":  "test-uid",
		"name": "environment",
	}
	requestCPU       = "1000m"
	requestMemory    = "100Mi"
	requestEphemeral = "1Gi"
	limitCPU         = "2000m"
	limitMemory      = "500Mi"
	limitEphemeral   = "2Gi"
	users            = []onboardingv1alpha1.User{{Username: "user1", Role: "admin"}, {Username: "user2", Role: "viewer"}}

	environment = &onboardingv1alpha1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Spec: onboardingv1alpha1.EnvironmentSpec{
			Name: projectname,
			Resources: onboardingv1alpha1.Resources{
				ResourceRequests: onboardingv1alpha1.ResourceDescription{
					CPU:              requestCPU,
					Memory:           requestMemory,
					EphemeralStorage: requestEphemeral,
				},
				ResourceLimits: onboardingv1alpha1.ResourceDescription{
					CPU:              limitCPU,
					Memory:           limitMemory,
					EphemeralStorage: limitEphemeral,
				},
			},
			Storage: storage,
			Users:   users,
		},
	}

	namespace = &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   environment.Spec.Name,
			Labels: environment.Labels,
		},
	}

	resourceQuota = &corev1.ResourceQuota{
		TypeMeta: metav1.TypeMeta{
			Kind: "ResourceQuota",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cno-resource-quota", //environment.Spec.Name,
			Namespace: environment.Spec.Name,
			Labels:    environment.Labels,
		},
		Spec: corev1.ResourceQuotaSpec{
			Hard: corev1.ResourceList{
				"requests.cpu":               resource.MustParse(environment.Spec.Resources.ResourceRequests.CPU),
				"requests.memory":            resource.MustParse(environment.Spec.Resources.ResourceRequests.Memory),
				"requests.ephemeral-storage": resource.MustParse(environment.Spec.Resources.ResourceRequests.EphemeralStorage),
				"limits.cpu":                 resource.MustParse(environment.Spec.Resources.ResourceLimits.CPU),
				"limits.memory":              resource.MustParse(environment.Spec.Resources.ResourceLimits.Memory),
				"limits.ephemeral-storage":   resource.MustParse(environment.Spec.Resources.ResourceLimits.EphemeralStorage),
				"requests.storage":           resource.MustParse(environment.Spec.Storage),
			},
		},
	}

	subjects = []v1.Subject{
		{
			Kind: "User",
			Name: environment.Spec.Users[0].Username,
		},
		{
			Kind: "User",
			Name: environment.Spec.Users[1].Username,
		},
	}

	adminRoleBinding = &v1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cno-admin-role-binding",
			Namespace: environment.Spec.Name,
			Labels:    environment.Labels,
		},
		Subjects: []v1.Subject{subjects[0]},
		RoleRef: v1.RoleRef{
			Name:     "cno-admin-cluster-role",
			Kind:     "ClusterRole",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}

	viewerRoleBinding = &v1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cno-viewer-role-binding",
			Namespace: environment.Spec.Name,
			Labels:    environment.Labels,
		},
		Subjects: []v1.Subject{subjects[1]},
		RoleRef: v1.RoleRef{
			Name:     "cno-viewer-cluster-role",
			Kind:     "ClusterRole",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
)

func TestEnvironmentController(t *testing.T) {
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(logf.ZapLogger(true))
	var log = logf.Log.WithName("controller_environment")

	// Objects to track in the fake client.
	objs := []runtime.Object{
		environment,
	}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(onboardingv1alpha1.SchemeGroupVersion, environment)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClient(objs...)

	// Create a ReconcileEnvironment object with the scheme and fake client.
	r := &ReconcileEnvironment{client: cl, scheme: s}

	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource .
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: "",
		},
	}
	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Logf("reconcile did not requeue request as expected")
	}

	// Check if a namespace has been created.
	ns := &corev1.Namespace{}
	log.WithValues("Environment Name", name).Info("Ns check ")
	err = cl.Get(context.TODO(), types.NamespacedName{Name: projectname}, ns)
	if err != nil {
		t.Fatalf("get namespace: (%v)", err)
	}

	res, err = r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if res.Requeue {
		t.Error("reconcile requeue which is not expected")
	}

	log.WithValues("Environment Name", name).Info("RQ check ")
	// Check if a resourcequota has been created
	foundRq := &corev1.ResourceQuota{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: "cno-resource-quota", Namespace: projectname}, foundRq)
	if err != nil {
		t.Fatalf("get resourcequota: (%v)", err)
	}

	res, err = r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if res.Requeue {
		t.Error("reconcile requeue which is not expected")
	}

	log.WithValues("Environment Name", name).Info("admin RB check ")
	// Check if an admin rolebinding has been created
	foundAdminRb := &v1.RoleBinding{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: "cno-admin-role-binding", Namespace: projectname}, foundAdminRb)
	if err != nil {
		t.Fatalf("get rolebinding: (%v)", err)
	}

	log.WithValues("Environment Name", name).Info("viewer RB check ")
	// Check if a viewer rolebinding has been created
	foundViewerRb := &v1.RoleBinding{}
	err = cl.Get(context.TODO(), types.NamespacedName{Name: "cno-viewer-role-binding", Namespace: projectname}, foundViewerRb)
	if err != nil {
		t.Fatalf("get rolebinding: (%v)", err)
	}

	t.Logf("Environment Controller Handled the environment: %v", name)
}

func TestNewNamespaceForCR(t *testing.T) {
	ns := newNamespaceForCR(environment)
	if !reflect.DeepEqual(namespace, ns) {
		t.Errorf("newNamespaceForCR didn't produce the expected output")
	} else {
		t.Logf("newNamespaceForCR produced the expected namespace")
	}
}

func TestNewResourceQuotaForCR(t *testing.T) {
	rq := newResourceQuotaForCR(environment)
	if !reflect.DeepEqual(resourceQuota, rq) {
		t.Errorf("newResourceQuotaForCR didn't produce the expected output")
	} else {
		t.Logf("newResourceQuotaForCR produced the expected resourcequota")
	}
}
func TestNewRoleBindingForCR(t *testing.T) {
	rb := newRoleBindingForCR(environment)
	// Check the admin rolebinding
	if !reflect.DeepEqual(adminRoleBinding, rb[0]) {
		t.Errorf("newRoleBindingForCR didn't produce the expected admin RoleBinding")
	} else {
		t.Logf("newRoleBindingForCR produced the expected rolebinding")
	}
	// Check the viewer rolebinding
	if !reflect.DeepEqual(viewerRoleBinding, rb[1]) {
		t.Errorf("newRoleBindingForCR didn't produce the expected viewer RoleBinding")
	} else {
		t.Logf("newRoleBindingForCR produced the expected rolebinding")
	}
}
