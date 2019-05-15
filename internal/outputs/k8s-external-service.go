package outputs

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	promOperatorV1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	promOperatorClient "github.com/coreos/prometheus-operator/pkg/client/versioned"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/camptocamp/prometheus-puppetdb/internal/types"
)

// OutputK8SExternalService stores data needed to create Kubernetes external services
type OutputK8SExternalService struct {
	namespace    string
	objectLabels map[string]string

	k8sClient        kubernetes.Interface
	monitoringClient promOperatorClient.Interface
}

func setupOutputK8SExternalService(namespace string, objectLabels map[string]string) (*OutputK8SExternalService, error) {
	o := &OutputK8SExternalService{
		namespace:    namespace,
		objectLabels: objectLabels,
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	o.k8sClient = clientset

	if namespace == "" {
		kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			clientcmd.NewDefaultClientConfigLoadingRules(),
			&clientcmd.ConfigOverrides{},
		)
		namespace, _, err := kubeconfig.Namespace()
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve namespace: %s", err)
		}
		o.namespace = namespace
	}

	o.monitoringClient, err = promOperatorClient.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	log.Warning("Output `external-services` is still an experimental feature.")

	return o, nil
}

// WriteOutput creates external services in the Kubernetes cluster
func (o *OutputK8SExternalService) WriteOutput(staticConfigs []types.StaticConfig) (err error) {
	newTargets := []string{}
	smEndpoints := []promOperatorV1.Endpoint{}
	var relabelConfigs []*promOperatorV1.RelabelConfig

	for _, sc := range staticConfigs {
		relabelConfigs = []*promOperatorV1.RelabelConfig{}
		for labelName, labelValue := range sc.Labels {
			relabelConfigs = append(relabelConfigs, &promOperatorV1.RelabelConfig{
				TargetLabel: labelName,
				Replacement: labelValue,
			})
		}

		for _, target := range sc.Targets {
			splittedTarget := strings.Split(target, ":")
			address := splittedTarget[0]

			port := 80
			if sc.Labels["scheme"] == "https" {
				port = 443
			}
			if len(splittedTarget) > 1 {
				port, _ = strconv.Atoi(splittedTarget[1])
			}

			name := fmt.Sprintf("puppetdb-%s-%d", strings.Replace(address, ".", "-", -1), port)
			newTargets = append(newTargets, name)

			err = o.upsertEndpoint(name, o.objectLabels, address, port)
			if err != nil {
				log.Error(err)
				continue
			}

			err = o.upsertService(name, o.objectLabels, address, port)
			if err != nil {
				log.Error(err)
				continue
			}

			smEndpoints = append(smEndpoints, promOperatorV1.Endpoint{
				Port:                 name,
				Scheme:               sc.Labels["scheme"],
				Path:                 sc.Labels["metrics_path"],
				HonorLabels:          true,
				MetricRelabelConfigs: relabelConfigs,
			})
		}
	}

	// Cleanup endpoints and services
	endpoints, err := o.listEndpoints()
	if err != nil {
		log.Errorf("failed to list endpoints: %s", err)
	}
	services, err := o.listServices()
	if err != nil {
		log.Errorf("failed to list services: %s", err)
	}
	var keep bool
	for _, endpoint := range endpoints {
		keep = false
		for _, t := range newTargets {
			if t == endpoint {
				keep = true
			}
		}
		if !keep {
			err = o.k8sClient.CoreV1().Endpoints(o.namespace).Delete(endpoint, &metav1.DeleteOptions{})
			if err != nil {
				log.Errorf("failed to delete endpoint `%s': %s", endpoint, err)
				continue
			}
			log.Infof("endpoint `%s' deleted", endpoint)
		}
	}
	for _, service := range services {
		keep = false
		for _, t := range newTargets {
			if t == service {
				keep = true
			}
		}
		if !keep {
			err = o.k8sClient.CoreV1().Services(o.namespace).Delete(service, &metav1.DeleteOptions{})
			if err != nil {
				log.Errorf("failed to delete endpoint `%s': %s", service, err)
				continue
			}
			log.Infof("service `%s' deleted", service)
		}
	}

	err = o.updateServiceMonitor(smEndpoints)
	if err != nil {
		log.Error(err)
	}

	return
}

func (o *OutputK8SExternalService) updateServiceMonitor(endpoints []promOperatorV1.Endpoint) (err error) {
	sm, err := o.monitoringClient.MonitoringV1().ServiceMonitors(o.namespace).Get("prometheus-puppetdb", metav1.GetOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			_, err = o.monitoringClient.MonitoringV1().ServiceMonitors(o.namespace).Create(&promOperatorV1.ServiceMonitor{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ServiceMonitor",
					APIVersion: "monitoring.coreos.com/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:   "prometheus-puppetdb",
					Labels: o.objectLabels,
				},
				Spec: promOperatorV1.ServiceMonitorSpec{
					Selector: metav1.LabelSelector{
						MatchLabels: map[string]string{
							"prometheus-puppetdb": "true",
						},
					},
					Endpoints: endpoints,
				},
			})
		} else {
			err = fmt.Errorf("failed to retrieve service monitor: %s", err)
			return
		}
	} else {
		_, err = o.monitoringClient.MonitoringV1().ServiceMonitors(o.namespace).Update(&promOperatorV1.ServiceMonitor{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ServiceMonitor",
				APIVersion: "monitoring.coreos.com/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:            "prometheus-puppetdb",
				Labels:          o.objectLabels,
				ResourceVersion: sm.ObjectMeta.ResourceVersion,
			},
			Spec: promOperatorV1.ServiceMonitorSpec{
				Selector: metav1.LabelSelector{
					MatchLabels: map[string]string{
						"prometheus-puppetdb": "true",
					},
				},
				Endpoints: endpoints,
			},
		})
	}
	return
}

func (o *OutputK8SExternalService) upsertEndpoint(name string, labels map[string]string, address string, port int) (err error) {
	endpoint, err := o.k8sClient.CoreV1().Endpoints(o.namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			o.createEndpoint(name, labels, address, port)
			err = nil
		} else {
			err = fmt.Errorf("failed to retrieve endpoint `%s': %s", name, err)
			return
		}
	} else {
		o.updateEndpoint(endpoint, name, labels, address, port)
	}
	return
}

func (o *OutputK8SExternalService) upsertService(name string, labels map[string]string, address string, port int) (err error) {
	svc, err := o.k8sClient.CoreV1().Services(o.namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			o.createService(name, labels, address, port)
			err = nil
		} else {
			err = fmt.Errorf("failed to retrieve service `%s': %s", name, err)
			return
		}
	} else {
		o.updateService(svc, name, labels, address, port)
	}
	return
}

func (o *OutputK8SExternalService) listEndpoints() (endpoints []string, err error) {
	endpoints = []string{}

	list, err := o.k8sClient.CoreV1().Endpoints(o.namespace).List(metav1.ListOptions{
		LabelSelector: "prometheus-puppetdb",
	})
	if err != nil {
		return
	}
	for _, endpoint := range list.Items {
		endpoints = append(endpoints, endpoint.ObjectMeta.Name)
	}
	return
}

func (o *OutputK8SExternalService) listServices() (services []string, err error) {
	services = []string{}

	list, err := o.k8sClient.CoreV1().Services(o.namespace).List(metav1.ListOptions{
		LabelSelector: "prometheus-puppetdb",
	})
	if err != nil {
		return
	}
	for _, service := range list.Items {
		services = append(services, service.ObjectMeta.Name)
	}
	return
}

func (o *OutputK8SExternalService) createEndpoint(name string, labels map[string]string, address string, port int) (err error) {
	ip, err := o.getIPFromDomain(address)
	if err != nil {
		err = fmt.Errorf("failed to retrieve IPv4 from domain")
		return
	}

	_, err = o.k8sClient.CoreV1().Endpoints(o.namespace).Create(&v1.Endpoints{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Endpoints",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"prometheus-puppetdb": "true",
			},
		},
		Subsets: []v1.EndpointSubset{
			{
				Addresses: []v1.EndpointAddress{
					{
						IP: ip,
					},
				},
				Ports: []v1.EndpointPort{
					{
						Name:     name,
						Port:     int32(port),
						Protocol: "TCP",
					},
				},
			},
		},
	})
	if err != nil {
		log.Errorf("failed to create endpoint `%s': %s", name, err)
		return
	}
	log.Infof("endpoint `%s' created", name)
	return
}

func (o *OutputK8SExternalService) updateEndpoint(endpoint *v1.Endpoints, name string, labels map[string]string, address string, port int) (err error) {
	ip, err := o.getIPFromDomain(address)
	if err != nil {
		err = fmt.Errorf("failed to retrieve IPv4 from domain")
		return
	}

	// Check if the endpoint needs to be updated
	update := false
	if endpoint.Subsets[0].Addresses[0].IP != ip {
		update = true
	}
	if endpoint.Subsets[0].Ports[0].Port != int32(port) {
		update = true
	}
	if endpoint.Subsets[0].Ports[0].Name != name {
		update = true
	}

	if !update {
		return
	}

	_, err = o.k8sClient.CoreV1().Endpoints(o.namespace).Update(&v1.Endpoints{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Endpoints",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"prometheus-puppetdb": "true",
			},
		},
		Subsets: []v1.EndpointSubset{
			{
				Addresses: []v1.EndpointAddress{
					{
						IP: ip,
					},
				},
				Ports: []v1.EndpointPort{
					{
						Name:     name,
						Port:     int32(port),
						Protocol: v1.ProtocolTCP,
					},
				},
			},
		},
	})
	if err != nil {
		log.Errorf("failed to update endpoint `%s': %s", name, err)
		return
	}
	log.Infof("endpoint `%s' updated", name)
	return
}

func (o *OutputK8SExternalService) createService(name string, labels map[string]string, address string, port int) (err error) {
	_, err = o.k8sClient.CoreV1().Services(o.namespace).Create(&v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"prometheus-puppetdb": "true",
			},
		},
		Spec: v1.ServiceSpec{
			Type:         v1.ServiceTypeExternalName,
			ExternalName: address,
			ClusterIP:    "",
			Ports: []v1.ServicePort{
				{
					Name:     name,
					Port:     int32(port),
					Protocol: v1.ProtocolTCP,
				},
			},
		},
	})
	if err != nil {
		log.Errorf("failed to create service `%s': %s", name, err)
		return
	}
	log.Infof("service `%s' created", name)
	return
}

func (o *OutputK8SExternalService) updateService(svc *v1.Service, name string, labels map[string]string, address string, port int) (err error) {
	// Check if the service needs to be updated
	update := false
	if svc.Spec.ExternalName != address {
		update = true
	}
	if svc.Spec.Ports[0].Port != int32(port) {
		update = true
	}
	if svc.Spec.Ports[0].Name != name {
		update = true
	}

	if !update {
		return
	}

	_, err = o.k8sClient.CoreV1().Services(o.namespace).Update(&v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"prometheus-puppetdb": "true",
			},
		},
		Spec: v1.ServiceSpec{
			Type:         v1.ServiceTypeExternalName,
			ExternalName: address,
			ClusterIP:    "",
			Ports: []v1.ServicePort{
				{
					Name:     name,
					Port:     int32(port),
					Protocol: v1.ProtocolTCP,
				},
			},
		},
	})
	if err != nil {
		log.Errorf("failed to update service `%s': %s", name, err)
		return
	}
	log.Infof("service `%s' updated", name)
	return
}

func (o *OutputK8SExternalService) getIPFromDomain(address string) (ip string, err error) {
	// TODO: add IPv6 support
	if net.ParseIP(address).To4() != nil {
		ip = address
		return
	}

	// We assume the address is a domain name
	ipList, err := net.LookupIP(address)
	if err != nil {
		return
	}
	for _, rawIP := range ipList {
		if rawIP.To4() != nil {
			ip = rawIP.String()
			return
		}
	}
	err = fmt.Errorf("no IPv4 found")
	return
}
