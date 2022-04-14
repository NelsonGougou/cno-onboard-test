package environment

import (
	"context"
	onboardingv1alpha1 "gitlab.beopenit.com/cloud/onboarding-operator-kubernetes/pkg/apis/onboarding/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_environment")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Environment Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileEnvironment{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("environment-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Environment
	err = c.Watch(&source.Kind{Type: &onboardingv1alpha1.Environment{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Namespace and requeue the owner Environment
	err = c.Watch(&source.Kind{Type: &corev1.Namespace{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &onboardingv1alpha1.Environment{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource ResourceQuota and requeue the owner Environment
	err = c.Watch(&source.Kind{Type: &corev1.ResourceQuota{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &onboardingv1alpha1.Environment{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource RoleBinding and requeue the owner Environment
	err = c.Watch(&source.Kind{Type: &v1.RoleBinding{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &onboardingv1alpha1.Environment{},
	})
	if err != nil {
		return err
	}
	return nil
}

// blank assignment to verify that ReconcileEnvironment implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileEnvironment{}

// ReconcileEnvironment reconciles a Environment object
type ReconcileEnvironment struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for an Environment object and makes changes based on the state read
// and what is in the Environment.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileEnvironment) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Environment Name", request.Name)
	reqLogger.Info("Reconciling Environment")

	// Fetch the Environment instance
	instance := &onboardingv1alpha1.Environment{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Define a new Namespace object
	namespace := newNamespaceForCR(instance)

	// Set Environment instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, namespace, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Namespace already exists
	foundNs := &corev1.Namespace{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: namespace.Name}, foundNs)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Namespace", "Namespace.Name", namespace.Name)
		err = r.client.Create(context.TODO(), namespace)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}
	reqLogger.Info("Namespace reconciled", "Namespace.Name", foundNs.Name)



	// Define a new resource quota object
	rq := newResourceQuotaForCR(instance)
	// Set Environment instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, rq, r.scheme); err != nil {
		return reconcile.Result{}, err
	}
	// Check if this ResourceQuota already exists
	foundResourceQuota := &corev1.ResourceQuota{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: rq.Name, Namespace: rq.Namespace}, foundResourceQuota)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new ResourceQuota", "ResourceQuota.Namespace", rq.Namespace, "ResourceQuota.Name", rq.Name)
		err = r.client.Create(context.TODO(), rq)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	} else {
		//resourceQuota update
		rq.ResourceVersion = foundResourceQuota.ResourceVersion
		err = r.client.Update(context.TODO(), rq)
		if err!=nil{
			return reconcile.Result{}, err
		}
	}
	reqLogger.Info("ResourceQuota reconciled", "ResourceQuota.Namespace", rq.Namespace, "ResourceQuota.Name", rq.Name)

	// Define a new resource limitRange object
	limitRange := getLimiteRange(instance)
	// Set Environment instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, limitRange, r.scheme); err != nil {
		return reconcile.Result{}, err
	}
	// Check if this LimitRange already exists
	foundLimitRange := &corev1.LimitRange{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: limitRange.Name, Namespace: limitRange.Namespace}, foundLimitRange)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new LimitRange", "LimitRange.Namespace", limitRange.Namespace, "LimitRange.Name", limitRange.Name)
		err = r.client.Create(context.TODO(), limitRange)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Define a new rolebinding object
	adminRolebinding := newRoleBindingForCR(instance)
	for _, rolebinding := range adminRolebinding {
		// Set Environment instance as the owner and controller
		if err := controllerutil.SetControllerReference(instance, rolebinding, r.scheme); err != nil {
			return reconcile.Result{}, err
		}
		// Check if this RoleBinding already exists
		foundRb := &v1.RoleBinding{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: rolebinding.Name, Namespace: rolebinding.Namespace}, foundRb)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating a new RoleBinding", "RoleBinding.Namespace", rolebinding.Namespace, "RoleBinding.Name", rolebinding.Name)
			err = r.client.Create(context.TODO(), rolebinding)
			if err != nil {
				return reconcile.Result{}, err
			}
		} else if err != nil {
			return reconcile.Result{}, err
		}else {
			//roleBinding update
			rolebinding.ResourceVersion = foundRb.ResourceVersion
			err = r.client.Update(context.TODO(), rolebinding)
			if err!=nil{
				return reconcile.Result{}, err
			}
		}
		reqLogger.Info("RoleBinding reconciled", "RoleBinding.Namespace", rolebinding.Namespace, "RoleBinding.Name", rolebinding.Name)
	}

	return reconcile.Result{}, nil
}

// newNamespaceForCR returns a namespace with the name and labels defined in the cr spec
func newNamespaceForCR(cr *onboardingv1alpha1.Environment) *corev1.Namespace {
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   cr.Spec.Name,
			Labels: cr.Labels,
		},
	}
	return namespace
}

// newResourceQuotaForCR returns a resourcequota with the name and labels defined in the cr spec
func newResourceQuotaForCR(cr *onboardingv1alpha1.Environment) *corev1.ResourceQuota {
	resourceQuota := &corev1.ResourceQuota{
		TypeMeta: metav1.TypeMeta{
			Kind: "ResourceQuota",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cno-resource-quota",
			Namespace: cr.Spec.Name,
			Labels:    cr.Labels,
		},
		Spec: corev1.ResourceQuotaSpec{
			Hard: corev1.ResourceList{
				"requests.cpu":               resource.MustParse(cr.Spec.Resources.ResourceRequests.CPU),
				"requests.memory":            resource.MustParse(cr.Spec.Resources.ResourceRequests.Memory),
				"requests.ephemeral-storage": resource.MustParse(cr.Spec.Resources.ResourceRequests.EphemeralStorage),
				"limits.cpu":                 resource.MustParse(cr.Spec.Resources.ResourceLimits.CPU),
				"limits.memory":              resource.MustParse(cr.Spec.Resources.ResourceLimits.Memory),
				"limits.ephemeral-storage":   resource.MustParse(cr.Spec.Resources.ResourceLimits.EphemeralStorage),
				"requests.storage":           resource.MustParse(cr.Spec.Storage),
			},
		},
	}
	return resourceQuota
}

func getLimiteRange(cr *onboardingv1alpha1.Environment) *corev1.LimitRange {
	return &corev1.LimitRange{
		TypeMeta: metav1.TypeMeta{
			Kind: "LimitRange",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cno-limit-range",
			Namespace: cr.Spec.Name,
			Labels:    cr.Labels,
		},
		Spec: corev1.LimitRangeSpec{
			Limits: []corev1.LimitRangeItem {
				{
					Type: "Container",
					Default: corev1.ResourceList{
						"cpu":    resource.MustParse("500m"),
						"memory": resource.MustParse("512Mi"),
					},
					DefaultRequest: corev1.ResourceList{
						"cpu":    resource.MustParse("100m"),
						"memory": resource.MustParse("256Mi"),
					},
				},
			},
		},
	}
}

// newRolebindingForCR returns a rolebinding with the name and labels defined in the cr spec
func newRoleBindingForCR(cr *onboardingv1alpha1.Environment) []*v1.RoleBinding {
	var result []*v1.RoleBinding
	var admins, viewers []v1.Subject
	for _, user := range cr.Spec.Users {
		if user.Role == "admin" {
			admins = append(admins, v1.Subject{
				Kind: "User",
				Name: user.Username,
			})
		} else if user.Role == "dev" && !cr.Spec.IsProd{
			admins = append(admins, v1.Subject{
				Kind: "User",
				Name: user.Username,
			})
		} else if user.Role == "viewer" {
			viewers = append(viewers, v1.Subject{
				Kind: "User",
				Name: user.Username,
			})
		}
	}
	adminRoleBinding := &v1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cno-admin-role-binding",
			Namespace: cr.Spec.Name,
			Labels:    cr.Labels,
		},
		Subjects: admins,
		RoleRef: v1.RoleRef{
			Name:     "cno-admin-cluster-role",
			Kind:     "ClusterRole",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
	result = append(result, adminRoleBinding)
	viewerRoleBinding := &v1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cno-viewer-role-binding",
			Namespace: cr.Spec.Name,
			Labels:    cr.Labels,
		},
		Subjects: viewers,
		RoleRef: v1.RoleRef{
			Name:     "cno-viewer-cluster-role",
			Kind:     "ClusterRole",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
	result = append(result, viewerRoleBinding)
	return result
}
