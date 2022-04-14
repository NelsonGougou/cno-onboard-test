package e2e

import (
	goctx "context"
	"fmt"
	"testing"
	"time"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	"gitlab.beopenit.com/cloud/onboarding-operator-kubernetes/pkg/apis"
	onboardingv1alpha1 "gitlab.beopenit.com/cloud/onboarding-operator-kubernetes/pkg/apis/onboarding/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

var (
	retryInterval        = time.Second * 5
	timeout              = time.Second * 60
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
	name                 = "environment"
	projectname          = "project1"
	storage              = "10Gi"
	uidUser              = "test-uid"
	labels               = map[string]string{
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

	exampleEnvironment = &onboardingv1alpha1.Environment{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Spec: onboardingv1alpha1.EnvironmentSpec{
			Name:    projectname,
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
)

func TestEnvironment(t *testing.T) {
	environmentList := &onboardingv1alpha1.EnvironmentList{}
	err := framework.AddToFrameworkScheme(apis.AddToScheme, environmentList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}
	// run subtests
	t.Run("environment-group", func(t *testing.T) {
		t.Run("Cluster", EnvironmentCluster)
	})
}

func environmentSetUpTest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) error {

	// use TestCtx's create helper to create the object and add a cleanup function for the new object
	err := f.Client.Create(goctx.TODO(), exampleEnvironment, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		return err
	}

	// Wait for Namespace
	err = waitForNamespace(t, f.KubeClient, exampleEnvironment.Spec.Name, retryInterval, timeout)
	if err != nil {
		fmt.Printf("err wait namespace: %v\n", err)
		return err
	}

	fmt.Printf("Created namespace is as expected\n")

	// Wait for Resource quota
	err = waitForResourceQuota(t, f.KubeClient, exampleEnvironment.Spec.Name, exampleEnvironment.Spec.Name, retryInterval, timeout)
	if err != nil {
		fmt.Printf("err wait quota: %v\n", err)
		return err
	}

	fmt.Printf("Created resourcequota is as expected\n")

	// Get Rolebinding
	err = waitForRoleBinding(t, f.KubeClient, exampleEnvironment.Spec.Name, exampleEnvironment.Spec.Name, retryInterval, timeout)
	if err != nil {
		fmt.Printf("err wait rolebinding: %v\n", err)
		return err
	}

	fmt.Printf("Created rolebinding is as expected\n")

	// test ok
	return nil
}

func EnvironmentCluster(t *testing.T) {
	t.Parallel()
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()
	err := ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}
	t.Log("Initialized cluster resources")
	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatal(err)
	}
	// get global framework variables
	f := framework.Global
	// wait for environment-onboardingv1alpha1 to be ready
	err = e2eutil.WaitForOperatorDeployment(t, f.KubeClient, namespace, "onboarding-operator-kubernetes", 1, retryInterval, timeout)
	if err != nil {
		fmt.Printf("Wait for operator error: %v \n", err)
		t.Fatal(err)
	}

	if err = environmentSetUpTest(t, f, ctx); err != nil {
		fmt.Printf("err test: %v\n", err)
		t.Fatal(err)
	}
}

func waitForNamespace(t *testing.T, kubeclient kubernetes.Interface, namespace string,
	retryInterval, timeout time.Duration) error {
	fmt.Println("============================= waitForNamespace ======================")
	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		ns, err := kubeclient.CoreV1().Namespaces().Get(goctx.TODO(), namespace, metav1.GetOptions{ResourceVersion: ""})
		if err != nil {
			if apierrors.IsNotFound(err) {
				fmt.Printf("Waiting for creation of Namespace: %s \n", ns.Name)
				return false, nil
			}
			fmt.Printf("err poll namespace: %v\n", err)
			return false, err
		}
		return true, nil
	})
	if err != nil {
		return err
	}
	fmt.Printf("Namespace %s available \n", name)
	return nil
}

func waitForResourceQuota(t *testing.T, kubeclient kubernetes.Interface, namespace, name string,
	retryInterval, timeout time.Duration) error {
	fmt.Println("============================= waitForResourceQuota ======================")
	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		rq, err := kubeclient.CoreV1().ResourceQuotas(namespace).Get(goctx.TODO(), name, metav1.GetOptions{ResourceVersion: ""})
		if err != nil {
			if apierrors.IsNotFound(err) {
				fmt.Printf("Waiting for creation of Resource quota: %s in Namespace: %s \n", rq.Name, namespace)
				return false, nil
			}
			fmt.Printf("err poll quota: %v\n", err)
			return false, err
		}
		return true, nil
	})
	if err != nil {
		return err
	}
	fmt.Printf("Resource quota %s in namespace %s is available \n", name, namespace)
	return nil
}

func waitForRoleBinding(t *testing.T, kubeclient kubernetes.Interface, namespace, name string,
	retryInterval, timeout time.Duration) error {
	fmt.Println("============================= waitForRoleBinding ======================")
	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		rb, err := kubeclient.RbacV1().RoleBindings(namespace).Get(goctx.TODO(), name, metav1.GetOptions{ResourceVersion: ""})
		if err != nil {
			if apierrors.IsNotFound(err) {
				fmt.Printf("Waiting for creation of Rolebinding: %s in Namespace: %s \n", rb.Name, namespace)
				return false, nil
			}
			fmt.Printf("err poll rolebinding: %v\n", err)
			return false, err
		}
		return true, nil
	})
	if err != nil {
		return err
	}
	fmt.Printf("Role binding %s in namespace %s is available \n", name, namespace)
	return nil
}
