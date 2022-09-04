package config

import (
	"io/ioutil"
	"net/http"
)

var CA []byte

func GetCA() {
	req, _ := http.NewRequest("GET", Conf.CaURL, nil)

	client := new(http.Client)
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	CA, _ = ioutil.ReadAll(resp.Body)
}
