package outputs

import (
	"fmt"
	"strings"

	yaml "gopkg.in/yaml.v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/camptocamp/prometheus-puppetdb-sd/internal/config"
	"github.com/camptocamp/prometheus-puppetdb-sd/internal/types"
)

// K8sSecretOutput stores data needed to fill a Kubernetes Secret
type K8sSecretOutput struct {
	k8sClient kubernetes.Interface

	secretName       string
	namespace        string
	objectLabels     map[string]string
	secretKey        string
	secretKeyPattern string
	extraSecretName  string
	extraSecretKey   string

	format config.OutputFormat
}

func setupK8sSecretOutput(cfg *config.OutputConfig) (*K8sSecretOutput, error) {
	o := &K8sSecretOutput{
		secretName:       cfg.K8sSecret.SecretName,
		namespace:        cfg.K8sSecret.Namespace,
		objectLabels:     cfg.K8sSecret.ObjectLabels,
		secretKey:        cfg.K8sSecret.SecretKey,
		secretKeyPattern: cfg.K8sSecret.SecretKeyPattern,
		extraSecretName:  cfg.K8sSecret.ExtraConfigSecretName,
		extraSecretKey:   cfg.K8sSecret.ExtraConfigSecretKey,

		format: cfg.Format,
	}

	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)

	config, err := kubeconfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	o.k8sClient = clientset

	if o.namespace == "" {
		namespace, _, err := kubeconfig.Namespace()
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve namespace: %s", err)
		}
		o.namespace = namespace
	}

	return o, nil
}

// WriteOutput writes Prometheus configuration to a Kubernetes Secret
func (o *K8sSecretOutput) WriteOutput(scrapeConfigs []*types.ScrapeConfig) (err error) {
	secret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:   o.secretName,
			Labels: o.objectLabels,
		},
	}

	// Output Secret
	_, err = o.k8sClient.CoreV1().Secrets(o.namespace).Get(o.secretName, metav1.GetOptions{})
	if err != nil {
		_, err = o.k8sClient.CoreV1().Secrets(o.namespace).Create(&secret)
		if err != nil {
			return fmt.Errorf("failed to create secret (%s)", err)
		}
	}

	// Extra Secret
	extraContent, err := o.getExtraConfigContent()

	secret.Data = map[string][]byte{}

	var c []byte
	var mc []byte

	switch o.format {
	case config.ScrapeConfigs:
		c, err = yaml.Marshal(scrapeConfigs)
		if err != nil {
			return
		}

		secret.Data[o.secretKey] = append(c, extraContent...)
	case config.StaticConfigs, config.MergedStaticConfigs:
		for _, scrapeConfig := range scrapeConfigs {
			c, err = yaml.Marshal(scrapeConfig.StaticConfigs)
			if err != nil {
				return
			}

			if o.format == config.MergedStaticConfigs {
				mc = append(mc, c...)
			} else {
				secret.Data[strings.Replace(o.secretKeyPattern, "*", scrapeConfig.JobName, 1)] = c
			}
		}

		if o.format == config.MergedStaticConfigs {
			secret.Data[o.secretKey] = append(mc, extraContent...)
		}
	default:
		err = fmt.Errorf("unexpected output format '%s'", o.format)

		return
	}

	_, err = o.k8sClient.CoreV1().Secrets(o.namespace).Update(&secret)
	if err != nil {
		return fmt.Errorf("failed to update secret (%s)", err)
	}

	return
}

// getExtraConfigContent returns the content of the extra config secret
func (o *K8sSecretOutput) getExtraConfigContent() (content []byte, err error) {
	var extraContent string

	if o.extraSecretName == "" || o.extraSecretKey == "" {
		return
	}

	extraSecret, err := o.k8sClient.CoreV1().Secrets(o.namespace).Get(o.extraSecretName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve extra secret (%s)", err)
	}

	if val, ok := extraSecret.Data[o.extraSecretKey]; ok {
		extraContent = string(val)
	} else {
		return nil, fmt.Errorf("failed to retrieve extra secret content (%s)", err)
	}
	content = []byte(extraContent + "\n")
	return
}
