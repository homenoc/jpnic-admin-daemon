package config

import (
	"io/ioutil"
	"net/http"
)

var CA []byte

func GetCA() error {
	req, err := http.NewRequest("GET", Conf.CaURL, nil)
	if err != nil {
		return err
	}
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	CA, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return nil
}
