package outputs

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	promOperatorFakeClient "github.com/coreos/prometheus-operator/pkg/client/versioned/fake"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testclient "k8s.io/client-go/kubernetes/fake"

	"github.com/camptocamp/prometheus-puppetdb/internal/types"
)

func TestK8SExternalServiceWriteOutputSuccess(t *testing.T) {
	data1 := []types.StaticConfig{
		types.StaticConfig{
			Targets: []string{
				"127.0.0.1:9103",
			},
			Labels: map[string]string{
				"foo":   "bar",
				"lorem": "ipsum",
			},
		},
		types.StaticConfig{
			Targets: []string{
				"10.0.0.1:9103",
				"10.0.0.1:9104",
			},
			Labels: map[string]string{},
		},
	}
	data2 := []types.StaticConfig{
		types.StaticConfig{
			Targets: []string{
				"127.0.0.1:9103",
			},
			Labels: map[string]string{
				"foo": "bar",
			},
		},
		types.StaticConfig{
			Targets: []string{
				"10.0.0.1:9101",
				"10.0.0.4:9104",
				"10.0.0.3:9103",
			},
			Labels: map[string]string{},
		},
	}

	o := &OutputK8SExternalService{
		namespace:        "foo",
		k8sClient:        testclient.NewSimpleClientset(),
		monitoringClient: promOperatorFakeClient.NewSimpleClientset(),
	}

	// First round
	err := o.WriteOutput(data1)
	assert.Nil(t, err)

	// Check ServiceMonitor
	serviceMonitor, err := o.monitoringClient.MonitoringV1().ServiceMonitors(o.namespace).Get("prometheus-puppetdb", metav1.GetOptions{})
	if err != nil {
		assert.FailNow(t, "failed to retrieve service monitor", err.Error())
	}

	for _, sc := range data1 {
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

			// Check service
			svc, err := o.k8sClient.CoreV1().Services(o.namespace).Get(name, metav1.GetOptions{})
			if err != nil {
				assert.FailNow(t, "failed to retrieve service", err.Error())
			}
			assert.Equal(t, address, svc.Spec.ExternalName)
			assert.Equal(t, int32(port), svc.Spec.Ports[0].Port)
			assert.Equal(t, name, svc.Spec.Ports[0].Name)

			// Check Prometheus endpoint
			endpoint, err := o.k8sClient.CoreV1().Endpoints(o.namespace).Get(name, metav1.GetOptions{})
			if err != nil {
				assert.FailNow(t, "failed to retrieve endpoint", err.Error())
			}
			assert.Equal(t, address, endpoint.Subsets[0].Addresses[0].IP)
			assert.Equal(t, int32(port), endpoint.Subsets[0].Ports[0].Port)
			assert.Equal(t, name, endpoint.Subsets[0].Ports[0].Name)

			// Check Service Monitor endpoint
			found := false
			for _, e := range serviceMonitor.Spec.Endpoints {
				if name == e.Port {
					found = true
				}
			}
			assert.True(t, found)
		}
	}

	// Second round
	err = o.WriteOutput(data2)
	assert.Nil(t, err)

	// Check ServiceMonitor
	serviceMonitor, err = o.monitoringClient.MonitoringV1().ServiceMonitors(o.namespace).Get("prometheus-puppetdb", metav1.GetOptions{})
	if err != nil {
		assert.FailNow(t, "failed to retrieve service monitor", err.Error())
	}

	for _, sc := range data2 {
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

			// Check service
			svc, err := o.k8sClient.CoreV1().Services(o.namespace).Get(name, metav1.GetOptions{})
			if err != nil {
				assert.FailNow(t, "failed to retrieve service", err.Error())
			}
			assert.Equal(t, address, svc.Spec.ExternalName)
			assert.Equal(t, int32(port), svc.Spec.Ports[0].Port)
			assert.Equal(t, name, svc.Spec.Ports[0].Name)

			// Check Prometheus endpoint
			endpoint, err := o.k8sClient.CoreV1().Endpoints(o.namespace).Get(name, metav1.GetOptions{})
			if err != nil {
				assert.FailNow(t, "failed to retrieve endpoint", err.Error())
			}
			assert.Equal(t, address, endpoint.Subsets[0].Addresses[0].IP)
			assert.Equal(t, int32(port), endpoint.Subsets[0].Ports[0].Port)
			assert.Equal(t, name, endpoint.Subsets[0].Ports[0].Name)

			// Check Service Monitor endpoint
			found := false
			for _, e := range serviceMonitor.Spec.Endpoints {
				if name == e.Port {
					found = true
				}
			}
			assert.True(t, found)
		}
	}

	// Targets cleanup
	_, err = o.k8sClient.CoreV1().Services(o.namespace).Get("puppetdb-10-0-0-1-9103", metav1.GetOptions{})
	if err == nil {
		assert.FailNow(t, "service `puppetdb-10-0-0-1-9103' should have been removed")
	}
	assert.Equal(t, "services \"puppetdb-10-0-0-1-9103\" not found", err.Error())

	_, err = o.k8sClient.CoreV1().Endpoints(o.namespace).Get("puppetdb-10-0-0-1-9103", metav1.GetOptions{})
	if err == nil {
		assert.FailNow(t, "endpoint `puppetdb-10-0-0-1-9103' should have been removed")
	}
	assert.Equal(t, "endpoints \"puppetdb-10-0-0-1-9103\" not found", err.Error())

	for _, e := range serviceMonitor.Spec.Endpoints {
		if e.Port == "puppetdb-10-0-0-1-9103" {
			assert.FailNow(t, "service monitor endpoint `puppetdb-10-0-0-1-9103' should have been removed")
		}
	}

}

/* updateService() tests */

func TestK8SExternalServiceUpdateServiceNotRequired(t *testing.T) {
	expectedAddress := "127.0.0.1"
	expectedPort := 9103
	expectedName := "puppetdb-127-0-0-1-9103"

	o := &OutputK8SExternalService{
		namespace:        "foo",
		k8sClient:        testclient.NewSimpleClientset(),
		monitoringClient: promOperatorFakeClient.NewSimpleClientset(),
	}

	svc, err := o.k8sClient.CoreV1().Services(o.namespace).Create(&v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: expectedName,
			Labels: map[string]string{
				"prometheus-puppetdb": "true",
			},
		},
		Spec: v1.ServiceSpec{
			Type:         v1.ServiceTypeExternalName,
			ExternalName: expectedAddress,
			ClusterIP:    "",
			Ports: []v1.ServicePort{
				{
					Name:     expectedName,
					Port:     int32(expectedPort),
					Protocol: v1.ProtocolTCP,
				},
			},
		},
	})
	if err != nil {
		assert.FailNow(t, "failed to create fake service", err)
	}

	err = o.updateService(svc, expectedName, map[string]string{}, expectedAddress, expectedPort)
	assert.Nil(t, err)

	svc, err = o.k8sClient.CoreV1().Services(o.namespace).Get(expectedName, metav1.GetOptions{})
	assert.Nil(t, err)
	assert.Equal(t, svc.Spec.ExternalName, expectedAddress)
	assert.Equal(t, svc.Spec.Ports[0].Port, int32(expectedPort))
	assert.Equal(t, svc.Spec.Ports[0].Name, expectedName)
}

func TestK8SExternalServiceUpdateServiceRequired(t *testing.T) {
	expectedAddress := "127.0.0.1"
	expectedPort := 9104
	expectedName := "foo"

	o := &OutputK8SExternalService{
		namespace:        "foo",
		k8sClient:        testclient.NewSimpleClientset(),
		monitoringClient: promOperatorFakeClient.NewSimpleClientset(),
	}

	svc, err := o.k8sClient.CoreV1().Services(o.namespace).Create(&v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: expectedName,
			Labels: map[string]string{
				"prometheus-puppetdb": "true",
			},
		},
		Spec: v1.ServiceSpec{
			Type:         v1.ServiceTypeExternalName,
			ExternalName: expectedAddress,
			ClusterIP:    "",
			Ports: []v1.ServicePort{
				{
					Name:     expectedName,
					Port:     int32(9103),
					Protocol: v1.ProtocolTCP,
				},
			},
		},
	})
	if err != nil {
		assert.FailNow(t, "failed to create fake service", err)
	}

	err = o.updateService(svc, expectedName, map[string]string{}, expectedAddress, expectedPort)
	assert.Nil(t, err)

	svc, err = o.k8sClient.CoreV1().Services(o.namespace).Get(expectedName, metav1.GetOptions{})
	assert.Nil(t, err)
	assert.Equal(t, svc.Spec.ExternalName, expectedAddress)
	assert.Equal(t, svc.Spec.Ports[0].Port, int32(expectedPort))
	assert.Equal(t, svc.Spec.Ports[0].Name, expectedName)
}

/* updateEndpoint() */

func TestK8SExternalServiceUpdateEndpointNotRequired(t *testing.T) {
	expectedAddress := "127.0.0.1"
	expectedPort := 9103
	expectedName := "puppetdb-127-0-0-1-9103"

	o := &OutputK8SExternalService{
		namespace:        "foo",
		k8sClient:        testclient.NewSimpleClientset(),
		monitoringClient: promOperatorFakeClient.NewSimpleClientset(),
	}

	// Create fake endpoint
	endpoint, err := o.k8sClient.CoreV1().Endpoints(o.namespace).Create(&v1.Endpoints{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Endpoint",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: expectedName,
			Labels: map[string]string{
				"prometheus-puppetdb": "true",
			},
		},
		Subsets: []v1.EndpointSubset{
			{
				Addresses: []v1.EndpointAddress{
					{
						IP: expectedAddress,
					},
				},
				Ports: []v1.EndpointPort{
					{
						Name:     expectedName,
						Port:     int32(expectedPort),
						Protocol: "TCP",
					},
				},
			},
		},
	})
	if err != nil {
		assert.FailNow(t, "failed to create fake endpoint", err)
	}

	err = o.updateEndpoint(endpoint, expectedName, map[string]string{}, expectedAddress, expectedPort)
	assert.Nil(t, err)

	endpoint, err = o.k8sClient.CoreV1().Endpoints(o.namespace).Get(expectedName, metav1.GetOptions{})
	assert.Nil(t, err)
	assert.Equal(t, endpoint.Subsets[0].Addresses[0].IP, expectedAddress)
	assert.Equal(t, endpoint.Subsets[0].Ports[0].Port, int32(expectedPort))
	assert.Equal(t, endpoint.Subsets[0].Ports[0].Name, expectedName)
}

func TestK8SExternalServiceUpdateEndpointRequired(t *testing.T) {
	expectedAddress := "127.0.0.1"
	expectedPort := 9104
	expectedName := "foo"

	o := &OutputK8SExternalService{
		namespace:        "foo",
		k8sClient:        testclient.NewSimpleClientset(),
		monitoringClient: promOperatorFakeClient.NewSimpleClientset(),
	}

	endpoint, err := o.k8sClient.CoreV1().Endpoints(o.namespace).Create(&v1.Endpoints{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Endpoints",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: expectedName,
			Labels: map[string]string{
				"prometheus-puppetdb": "true",
			},
		},
		Subsets: []v1.EndpointSubset{
			{
				Addresses: []v1.EndpointAddress{
					{
						IP: expectedAddress,
					},
				},
				Ports: []v1.EndpointPort{
					{
						Name:     expectedName,
						Port:     int32(9103),
						Protocol: v1.ProtocolTCP,
					},
				},
			},
		},
	})
	if err != nil {
		assert.FailNow(t, "failed to create fake endpoint", err)
	}

	err = o.updateEndpoint(endpoint, expectedName, map[string]string{}, expectedAddress, expectedPort)
	assert.Nil(t, err)

	endpoint, err = o.k8sClient.CoreV1().Endpoints(o.namespace).Get(expectedName, metav1.GetOptions{})
	assert.Nil(t, err)
	assert.Equal(t, endpoint.Subsets[0].Addresses[0].IP, expectedAddress)
	assert.Equal(t, endpoint.Subsets[0].Ports[0].Port, int32(expectedPort))
	assert.Equal(t, endpoint.Subsets[0].Ports[0].Name, expectedName)
}

func TestCreateEndpointDomainName(t *testing.T) {
	o := &OutputK8SExternalService{
		namespace:        "foo",
		k8sClient:        testclient.NewSimpleClientset(),
		monitoringClient: promOperatorFakeClient.NewSimpleClientset(),
	}

	err := o.createEndpoint("foo", map[string]string{}, "localhost", 9103)
	assert.Nil(t, err)
}

func TestK8SExternalServiceGetIPFromDomainValidIPv4(t *testing.T) {
	o := &OutputK8SExternalService{
		namespace:        "foo",
		k8sClient:        testclient.NewSimpleClientset(),
		monitoringClient: promOperatorFakeClient.NewSimpleClientset(),
	}
	ip, err := o.getIPFromDomain("127.0.0.1")
	assert.Nil(t, err)
	assert.Equal(t, ip, "127.0.0.1")
}

// IPv6 not yet supported
func TestK8SExternalServiceGetIPFromDomainValidIPv6(t *testing.T) {
	o := &OutputK8SExternalService{
		namespace:        "foo",
		k8sClient:        testclient.NewSimpleClientset(),
		monitoringClient: promOperatorFakeClient.NewSimpleClientset(),
	}
	_, err := o.getIPFromDomain("2001:0db8:85a3:0000:0000:8a2e:0370:7334")
	assert.NotNil(t, err)
}

func TestK8SExternalServiceGetIPFromDomainValidDomain(t *testing.T) {
	o := &OutputK8SExternalService{
		namespace:        "foo",
		k8sClient:        testclient.NewSimpleClientset(),
		monitoringClient: promOperatorFakeClient.NewSimpleClientset(),
	}
	ip, err := o.getIPFromDomain("localhost")
	assert.Nil(t, err)
	assert.Equal(t, ip, "127.0.0.1")
}
