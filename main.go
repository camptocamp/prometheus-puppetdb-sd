package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type Node struct {
	Certname  string `json:"certname"`
	Ipaddress string `json:"value"`
}

type Targets struct {
	Targets []string `yaml:"targets"`
}

const (
	query = "facts { name='ipaddress' and nodes { facts { name='collectd_version' and value ~ '^5\\\\.7' } and resources { type='Class' and title='Collectd' } } }"
	port  = "9103"
	file  = "/etc/prometheus-config/prometheus-targets.yml"
	sleep = 5 * time.Second
)

func main() {
	client := &http.Client{}

	for {
		nodes, err := getNodes(client)
		if err != nil {
			fmt.Println(err)
			break
		}

		err = writeNodes(nodes)
		if err != nil {
			fmt.Println(err)
			break
		}

		fmt.Printf("Sleeping for %v\n", sleep)
		time.Sleep(sleep)
	}
}

func getNodes(client *http.Client) (nodes []Node, err error) {
	form := strings.NewReader(fmt.Sprintf("{\"query\":\"%s\"}", query))
	req, err := http.NewRequest("POST", "http://puppetdb:8080/pdb/query/v4", form)
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

func writeNodes(nodes []Node) (err error) {
	targets := Targets{}

	for _, node := range nodes {
		target := fmt.Sprintf("%s:%s", node.Ipaddress, port)
		targets.Targets = append(targets.Targets, target)
	}

	d, err := yaml.Marshal(&targets)

	err = ioutil.WriteFile(file, d, 0644)
	return nil
}
