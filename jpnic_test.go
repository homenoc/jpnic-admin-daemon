package main

import (
	"github.com/homenoc/jpnic-gui-daemon/pkg/core/jpnic"
	"log"
	"net"
	"testing"
)

var caFilePath = "/Users/y-yoneda/Documents/homenoc-cert/rootcacert_r3.cer"

// Sakura
//var pfxPass = "sakura"
//var pfxFilePathV4 = "/Users/y-yoneda/Documents/sakura-cert/AS5970.p12"
//var pfxFilePathV6 = "/Users/y-yoneda/github/homenoc/jpnic-go/cert/v6-openssl.p12"

// HomeNOC
var pfxPass = "homenoc"
var pfxFilePathV4 = "/Users/y-yoneda/Documents/homenoc-cert/v4.p12"

var ryakusho = "HOMENOC"

func TestSearchIPv4(t *testing.T) {
	var infos []jpnic.InfoIPv4

	jpnicConfig := jpnic.Config{
		URL:         "https://iphostmaster.nic.ad.jp/jpnic/certmemberlogin.do",
		PfxFilePath: pfxFilePathV4,
		PfxPass:     pfxPass,
		CAFilePath:  caFilePath,
	}

	isOverList := true
	addressRange := ""

	for isOverList {
		filter := jpnic.SearchIPv4{
			IsDetail: false,
			Option1:  nil,
			Ryakusho: ryakusho,
		}
		if addressRange != "" {
			filter.IPAddress = addressRange
		}
		data, err := jpnicConfig.SearchIPv4(filter)
		if err != nil {
			t.Fatal(err)
		}

		isOverList = data.IsOverList

		if isOverList {
			lastIPAddress, _, err := net.ParseCIDR(data.InfoIPv4[len(data.InfoIPv4)-1].IPAddress)
			if err != nil {
				t.Fatal(err)
			}
			log.Println(lastIPAddress)

			addressRange = lastIPAddress.String() + "-255.255.255.255"
		}

		infos = append(infos, data.InfoIPv4...)
	}
	log.Println(infos)
	log.Println(len(infos))
}
