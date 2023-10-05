package source

import (
	"context"
	"time"

	"github.com/hyperkineticnerd/k8s-firewall/provider/juniper"
	"github.com/sirupsen/logrus"

	// "k8s-firewall/core"

	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	// appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/labels"
	kubeinformers "k8s.io/client-go/informers"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type serviceSource struct {
	client           kubernetes.Interface
	namespace        string
	annotationFilter string

	serviceInformer coreinformers.ServiceInformer
	podInformer     coreinformers.PodInformer
	nodeInformer    coreinformers.NodeInformer
	labelSelector   labels.Selector
}

// type Service interface {
// 	NewServiceSource(ctx context.Context, kubeClient kubernetes.Interface) (*Source, error)
// }

func NewServiceSource(ctx context.Context, kubeClient kubernetes.Interface, namespace string) (Source, error) {
	informerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, 10*time.Minute)
	serviceInformer := informerFactory.Core().V1().Services()
	podInformer := informerFactory.Core().V1().Pods()
	nodeInformer := informerFactory.Core().V1().Nodes()

	serviceInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
			},
		},
	)
	podInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
			},
		},
	)
	nodeInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
			},
		},
	)

	// informerFactory.Start(ctx.Done())
	stopCh := make(chan struct{})
	informerFactory.Start(stopCh) // runs in background
	informerFactory.WaitForCacheSync(stopCh)

	return &serviceSource{
		client:           kubeClient,
		namespace:        namespace,
		annotationFilter: "",
		serviceInformer:  serviceInformer,
		nodeInformer:     nodeInformer,
		podInformer:      podInformer,
	}, nil
}

func (s *serviceSource) getProtoFromService(svc *corev1.Service, port int) string {
	for _, port := range svc.Spec.Ports {
		logrus.Debugf("Proto: %s", port.Protocol)
		switch port.Protocol {
		case "TCP":
			return "tcp"
		case "UDP":
			return "udp"
		default:
			return "error"
		}
	}
	// return string(svc.Spec.Ports[port].Protocol)
	return ""
}

func (s *serviceSource) getPortFromService(svc *corev1.Service, port int) int32 {
	for _, port := range svc.Spec.Ports {
		logrus.Debugf("Port: %d", port.NodePort)
		return port.NodePort
	}
	// return svc.Spec.Ports[port].Port
	return 0
}

func (s *serviceSource) getNodeAddresses(nodeName string) string {
	// Fetch Pod Nodes
	node, _ := s.nodeInformer.Lister().Get(nodeName)
	nodeAddresses := node.Status.Addresses
	for _, addr := range nodeAddresses {
		switch addr.Type {
		case corev1.NodeInternalIP:
			return addr.Address
		case corev1.NodeExternalIP:
			return addr.Address
		}
	}
	return ""
}

func (s *serviceSource) getTarget(svc *corev1.Service) string {
	// Get pod selectors from Service
	podSelectors := labels.SelectorFromSet(svc.Spec.Selector)
	// Fetch Pods
	pods, _ := s.podInformer.Lister().Pods(svc.Namespace).List(podSelectors)
	for _, pod := range pods {
		// Return Pod HostIP
		logrus.Debugf("NodeName: %s", pod.Spec.NodeName)
		return pod.Spec.NodeName
	}
	// Fallback to Service Name
	return svc.Name
}

func (s *serviceSource) getIpAddr(svc *corev1.Service) string {
	// Get pod selectors from Service
	podSelectors := labels.SelectorFromSet(svc.Spec.Selector)
	// Fetch Pods
	pods, _ := s.podInformer.Lister().Pods(svc.Namespace).List(podSelectors)
	for _, pod := range pods {
		// Return Pod HostIP
		// sc.getNodeAddresses(ctx, pod.Spec.NodeName)
		return pod.Status.HostIP
	}
	// Fallback to Service ClusterIP
	return svc.Spec.ClusterIP
}

func (s *serviceSource) parseNodePort(ctx context.Context, svc *corev1.Service) *juniper.PortForward {
	return &juniper.PortForward{
		Name:         getNameFromAnnotation(svc.Annotations),
		Target:       s.getTarget(svc),
		IpAddr:       s.getIpAddr(svc),
		Port:         s.getPortFromService(svc, 0),
		Proto:        s.getProtoFromService(svc, 0),
		Policy:       getPolicyFromAnnotation(svc.Annotations),
		ExternalZone: getExternalZoneFromAnnotation(svc.Annotations),
		InternalZone: getInternalZoneFromAnnotation(svc.Annotations),
		RuleSet:      getRuleSetFromAnnotation(svc.Annotations),
	}
}

func (s *serviceSource) PortForwards(ctx context.Context) ([]*juniper.PortForward, error) {
	portforwards := []*juniper.PortForward{}
	services, err := s.serviceInformer.Lister().Services("").List(labels.Everything())
	if err != nil {
		return nil, err
	}
	for _, svc := range services {
		switch svc.Spec.Type {
		case corev1.ServiceTypeNodePort:
			target := getNameFromAnnotation(svc.Annotations)
			if target == "" {
				logrus.Debugf("Skipping service due to missing Annotations")
				continue
			}
			portforwards = append(portforwards, s.parseNodePort(ctx, svc))
		}
	}
	return portforwards, nil
}
