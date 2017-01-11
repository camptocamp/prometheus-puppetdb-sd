package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type Node struct {
	Certname  string `json:"certname"`
	Ipaddress string `json:"value"`
}

const (
	query = "facts{name='ipaddress' and nodes{resources{type='Class' and title='Collectd'}}}"
	sleep = 5 * time.Second
)

func main() {
	client := &http.Client{}

	for {
		form := strings.NewReader(fmt.Sprintf("{\"query\":\"%s\"}", query))
		req, err := http.NewRequest("POST", "http://localhost:8080/pdb/query/v4", form)
		if err != nil {
			fmt.Println(err)
			break
		}
		req.Header.Add("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			break
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			break
		}

		nodes := make([]Node, 0)
		err = json.Unmarshal(body, &nodes)
		if err != nil {
			fmt.Println(err)
			break
		}

		for _, node := range nodes {
			fmt.Printf("%s => %s\n", node.Certname, node.Ipaddress)
		}

		fmt.Printf("Sleeping for %v", sleep)
		time.Sleep(sleep)
	}
}
