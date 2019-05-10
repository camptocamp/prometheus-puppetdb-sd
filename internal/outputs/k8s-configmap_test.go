package outputs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testclient "k8s.io/client-go/kubernetes/fake"

	"github.com/camptocamp/prometheus-puppetdb/internal/types"
)

func TestK8SConfigMapWriteOutputSuccess(t *testing.T) {
	data := []types.StaticConfig{
		{
			Targets: []string{
				"127.0.0.1:9103",
			},
			Labels: map[string]string{
				"foo": "bar",
			},
		},
	}

	o := &OutputK8SConfigMap{
		namespace:     "foo",
		configMapName: "bar",
		k8sClient:     testclient.NewSimpleClientset(),
	}

	err := o.WriteOutput(data)
	assert.Nil(t, err)

	cm, err := o.k8sClient.CoreV1().ConfigMaps(o.namespace).Get(o.configMapName, metav1.GetOptions{})
	if err != nil {
		assert.FailNow(t, "failed to retrieve config map", err.Error())
	}
	assert.Equal(
		t,
		map[string]string{
			"targets.yml": "- targets:\n  - 127.0.0.1:9103\n  labels:\n    foo: bar\n",
		},
		cm.Data,
	)

}
