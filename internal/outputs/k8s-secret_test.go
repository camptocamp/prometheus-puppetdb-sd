package outputs

import (
	"strings"
	"testing"

	"github.com/camptocamp/prometheus-puppetdb-sd/internal/config"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testclient "k8s.io/client-go/kubernetes/fake"
)

func (o *K8sSecretOutput) testK8sSecretWriteOutput(t *testing.T) {
	o.k8sClient = testclient.NewSimpleClientset()

	o.secretName = "prometheus-puppetdb-sd-out"
	o.namespace = "monitoring"
	o.objectLabels = map[string]string{
		"app.kubernetes.io/name": "prometheus-puppetdb-sd",
	}
	o.secretKey = "puppetdb-sd.yml"
	o.secretKeyPattern = "*.yml"

	oldKeys := map[string]bool{}

	for i := range scrapeConfigs {
		err := o.WriteOutput(scrapeConfigs[i])
		if err != nil {
			assert.FailNow(t, "Failed to write output", err.Error())
		}

		secret, err := o.k8sClient.CoreV1().Secrets(o.namespace).Get(o.secretName, metav1.GetOptions{})
		if err != nil {
			assert.FailNow(t, "Failed to retrieve secret", err.Error())
		}

		switch o.format {
		case config.ScrapeConfigs, config.MergedStaticConfigs:
			output := string(secret.Data[o.secretKey])
			expectedOutput := expectedOutputs[i][o.format].(string)

			assert.Equal(t, strings.TrimSpace(expectedOutput), strings.TrimSpace(output))
		case config.StaticConfigs:
			keys := map[string]bool{}

			for _, scrapeConfig := range scrapeConfigs[i] {
				jobName := scrapeConfig.JobName

				key := strings.Replace(o.secretKeyPattern, "*", jobName, 1)

				output := string(secret.Data[key])
				expectedOutput := expectedOutputs[i][o.format].(map[string]string)[jobName]

				assert.Equal(t, strings.TrimSpace(expectedOutput), strings.TrimSpace(output))

				keys[key] = true
				delete(oldKeys, key)
			}

			for key := range oldKeys {
				if _, ok := secret.Data[key]; ok {
					assert.Fail(t, "Unexpected key in output secret", "key: %s", key)
				}
			}

			oldKeys = keys
		}
	}
}

func TestK8sSecretWriteOutputScrapeConfigsSuccess(t *testing.T) {
	o := K8sSecretOutput{
		format: config.ScrapeConfigs,
	}

	o.testK8sSecretWriteOutput(t)
}

func TestK8sSecretWriteOutputStaticConfigsSuccess(t *testing.T) {
	o := K8sSecretOutput{
		format: config.StaticConfigs,
	}

	o.testK8sSecretWriteOutput(t)
}

func TestK8sSecretWriteOutputMergedStaticConfigsSuccess(t *testing.T) {
	o := K8sSecretOutput{
		format: config.MergedStaticConfigs,
	}

	o.testK8sSecretWriteOutput(t)
}
