package outputs

import (
	"fmt"

	yaml "gopkg.in/yaml.v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// OutputK8SConfigMap stores data needed to fill a Kubernetes ConfigMap
type OutputK8SConfigMap struct {
	namespace     string
	configMapName string

	k8sClient kubernetes.Interface
}

func setupOutputK8SConfigMap(namespace, configMapName string) (*OutputK8SConfigMap, error) {
	o := &OutputK8SConfigMap{
		namespace:     namespace,
		configMapName: configMapName,
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
	return o, nil
}

// WriteOutput writes data to a Kubernetes ConfigMap
func (o *OutputK8SConfigMap) WriteOutput(data interface{}) (err error) {
	configMap, err := o.k8sClient.CoreV1().ConfigMaps(o.namespace).Get(o.configMapName, metav1.GetOptions{})
	if err != nil {
		configMap = &v1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ConfigMap",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: o.configMapName,
			},
			Data: map[string]string{
				"targets.yml": "",
			},
		}
		configMap, err = o.k8sClient.CoreV1().ConfigMaps(o.namespace).Create(configMap)
		if err != nil {
			return fmt.Errorf("failed to create configmap: %s", err)
		}
	}

	c, err := yaml.Marshal(&data)
	if err != nil {
		return
	}

	configMap = &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: o.configMapName,
		},
		Data: map[string]string{
			"targets.yml": string(c),
		},
	}
	configMap, err = o.k8sClient.CoreV1().ConfigMaps(o.namespace).Update(configMap)
	if err != nil {
		return fmt.Errorf("failed to update configmap: %s", err)
	}
	return
}
