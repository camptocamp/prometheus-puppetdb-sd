package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"

	"gopkg.in/yaml.v2"
)

var version = "undefined"

type Config struct {
	Version     bool   `short:"V" long:"version" description:"Display version."`
	PuppetDBURL string `short:"u" long:"puppetdb-url" description:"PuppetDB base URL." env:"PROMETHEUS_PUPPETDB_URL" default:"http://puppetdb:8080"`
	Query       string `short:"q" long:"puppetdb-query" description:"PuppetDB query." env:"PROMETHEUS_PUPPETDB_QUERY" default:"facts { name='ipaddress' and nodes { deactivated is null and ! (facts { name = 'osfamily' and value = 'RedHat' } and facts { name='operatingsystemmajrelease' and value = '5' }) and facts { name='collectd_version' and value ~ '^5\\\\.7' } and resources { type='Class' and title='Collectd' } } }"`
	Port        int    `short:"p" long:"collectd-port" description:"Collectd port." env:"PROMETHEUS_PUPPETDB_COLLECTD_PORT" default:"9103"`
	File        string `short:"c" long:"config-file" description:"Prometheus target file." env:"PROMETHEUS_PUPPETDB_FILE" default:"/etc/prometheus-targets/prometheus-targets.yml"`
	Sleep       string `short:"s" long:"sleep" description:"Sleep time between queries." env:"PROMETHEUS_PUPPETDB_SLEEP" default:"5s"`
	Manpage     bool   `short:"m" long:"manpage" description:"Output manpage."`
}

type Node struct {
	Certname  string `json:"certname"`
	Ipaddress string `json:"value"`
}

type Override struct {
	Certname string                 `json:"certname"`
	Override map[string]interface{} `json:"value"`
}

type Targets struct {
	Targets []string          `yaml:"targets"`
	Labels  map[string]string `yaml:"labels"`
}

func main() {
	cfg, err := loadConfig(version)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	client := &http.Client{}

	for {
		nodes, err := getNodes(client, cfg.PuppetDBURL, cfg.Query)
		if err != nil {
			fmt.Println(err)
			break
		}

		overrides, err := getOverrides(client, cfg.PuppetDBURL)
		if err != nil {
			fmt.Println(err)
			break
		}

		err = writeNodes(nodes, overrides, cfg.Port, cfg.File)
		if err != nil {
			fmt.Println(err)
			break
		}

		sleep, err := time.ParseDuration(cfg.Sleep)
		if err != nil {
			fmt.Println(err)
			break
		}
		fmt.Printf("Sleeping for %v\n", sleep)
		time.Sleep(sleep)
	}
}

func loadConfig(version string) (c Config, err error) {
	parser := flags.NewParser(&c, flags.Default)
	_, err = parser.Parse()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if c.Version {
		fmt.Printf("Prometheus-puppetdb v%v\n", version)
		os.Exit(0)
	}

	if c.Manpage {
		var buf bytes.Buffer
		parser.ShortDescription = "Prometheus scrape lists based on PuppetDB"
		parser.WriteManPage(&buf)
		fmt.Printf(buf.String())
		os.Exit(0)
	}
	return
}

func getNodes(client *http.Client, puppetdb string, query string) (nodes []Node, err error) {
	form := strings.NewReader(fmt.Sprintf("{\"query\":\"%s\"}", query))
	puppetdbURL := fmt.Sprintf("%s/pdb/query/v4", puppetdb)
	req, err := http.NewRequest("POST", puppetdbURL, form)
	if err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(body, &nodes)
	return
}

func getOverrides(client *http.Client, puppetdb string) (overrides map[string]map[string]interface{}, err error) {
	form := strings.NewReader(fmt.Sprintf("{\"query\":\"%s\"}", "facts[certname, value] { name='prometheus_target_conf' }"))
	puppetdbURL := fmt.Sprintf("%s/pdb/query/v4", puppetdb)
	req, err := http.NewRequest("POST", puppetdbURL, form)
	if err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var nodes []Override
	err = json.Unmarshal(body, &nodes)

	overrides = make(map[string]map[string]interface{})
	for _, node := range nodes {
		overrides[node.Certname] = node.Override
	}

	return
}

func writeNodes(nodes []Node, overrides map[string]map[string]interface{}, port int, file string) (err error) {
	allTargets := []Targets{}

	for _, node := range nodes {
		var targets = Targets{}
		var hostname = node.Ipaddress
		var zeport = port
		if o, ok := overrides[node.Certname]; ok {
			if h, ok := o["hostname"]; ok {
				hostname = h.(string)
			}
			if p, ok := o["port"]; ok {
				zeport = int(p.(float64))
			}
		}
		var target = fmt.Sprintf("%s:%v", hostname, zeport)
		targets.Targets = append(targets.Targets, target)
		targets.Labels = map[string]string{
			"job":      "collectd",
			"certname": node.Certname,
		}
		allTargets = append(allTargets, targets)
	}

	d, err := yaml.Marshal(&allTargets)

	err = ioutil.WriteFile(file, d, 0644)
	return nil
}
