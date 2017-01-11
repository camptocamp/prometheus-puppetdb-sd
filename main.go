package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type Node struct {
	Certname  string `json:"certname"`
	Ipaddress string `json:"value"`
}

func main() {
	form := strings.NewReader("{\"query\":\"facts{name='ipaddress' and nodes{resources{type='Class' and title='Collectd'}}}\"}")

	client := &http.Client{}
	req, _ := http.NewRequest("POST", "http://localhost:8080/pdb/query/v4", form)
	req.Header.Add("Content-Type", "application/json")

	resp, _ := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)

	nodes := make([]Node, 0)
	err := json.Unmarshal(body, &nodes)
	if err != nil {
		fmt.Println(err)
	}

	for _, node := range nodes {
		fmt.Printf("%s => %s\n", node.Certname, node.Ipaddress)
	}
}
