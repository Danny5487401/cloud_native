<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [clientset](#clientset)
  - [源码](#%E6%BA%90%E7%A0%81)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# clientset
clientset 代表了kubernetes中所有的资源类型，这里不包含CRD的资源 .如：

- core
- extensions
- batch
- ...

## 源码

```go
// /Users/python/go/pkg/mod/k8s.io/client-go@v0.24.0/kubernetes/clientset.go

// Clientset contains the clients for groups. Each group has exactly one
// version included in a Clientset.
type Clientset struct {
	*discovery.DiscoveryClient
	admissionregistrationV1      *admissionregistrationv1.AdmissionregistrationV1Client
	admissionregistrationV1beta1 *admissionregistrationv1beta1.AdmissionregistrationV1beta1Client
	internalV1alpha1             *internalv1alpha1.InternalV1alpha1Client
	appsV1                       *appsv1.AppsV1Client
	appsV1beta1                  *appsv1beta1.AppsV1beta1Client
	appsV1beta2                  *appsv1beta2.AppsV1beta2Client
	authenticationV1             *authenticationv1.AuthenticationV1Client
	authenticationV1beta1        *authenticationv1beta1.AuthenticationV1beta1Client
	authorizationV1              *authorizationv1.AuthorizationV1Client
	authorizationV1beta1         *authorizationv1beta1.AuthorizationV1beta1Client
	autoscalingV1                *autoscalingv1.AutoscalingV1Client
	autoscalingV2                *autoscalingv2.AutoscalingV2Client
	autoscalingV2beta1           *autoscalingv2beta1.AutoscalingV2beta1Client
	autoscalingV2beta2           *autoscalingv2beta2.AutoscalingV2beta2Client
	batchV1                      *batchv1.BatchV1Client
	batchV1beta1                 *batchv1beta1.BatchV1beta1Client
	certificatesV1               *certificatesv1.CertificatesV1Client
	certificatesV1beta1          *certificatesv1beta1.CertificatesV1beta1Client
	coordinationV1beta1          *coordinationv1beta1.CoordinationV1beta1Client
	coordinationV1               *coordinationv1.CoordinationV1Client
	coreV1                       *corev1.CoreV1Client
	discoveryV1                  *discoveryv1.DiscoveryV1Client
	discoveryV1beta1             *discoveryv1beta1.DiscoveryV1beta1Client
	eventsV1                     *eventsv1.EventsV1Client
	eventsV1beta1                *eventsv1beta1.EventsV1beta1Client
	extensionsV1beta1            *extensionsv1beta1.ExtensionsV1beta1Client
	flowcontrolV1alpha1          *flowcontrolv1alpha1.FlowcontrolV1alpha1Client
	flowcontrolV1beta1           *flowcontrolv1beta1.FlowcontrolV1beta1Client
	flowcontrolV1beta2           *flowcontrolv1beta2.FlowcontrolV1beta2Client
	networkingV1                 *networkingv1.NetworkingV1Client
	networkingV1beta1            *networkingv1beta1.NetworkingV1beta1Client
	nodeV1                       *nodev1.NodeV1Client
	nodeV1alpha1                 *nodev1alpha1.NodeV1alpha1Client
	nodeV1beta1                  *nodev1beta1.NodeV1beta1Client
	policyV1                     *policyv1.PolicyV1Client
	policyV1beta1                *policyv1beta1.PolicyV1beta1Client
	rbacV1                       *rbacv1.RbacV1Client
	rbacV1beta1                  *rbacv1beta1.RbacV1beta1Client
	rbacV1alpha1                 *rbacv1alpha1.RbacV1alpha1Client
	schedulingV1alpha1           *schedulingv1alpha1.SchedulingV1alpha1Client
	schedulingV1beta1            *schedulingv1beta1.SchedulingV1beta1Client
	schedulingV1                 *schedulingv1.SchedulingV1Client
	storageV1beta1               *storagev1beta1.StorageV1beta1Client
	storageV1                    *storagev1.StorageV1Client
	storageV1alpha1              *storagev1alpha1.StorageV1alpha1Client
}
```
