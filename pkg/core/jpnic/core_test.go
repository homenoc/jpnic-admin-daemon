package jpnic

import (
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

func TestGetResourceAll(t *testing.T) {
	con := Config{
		PfxFilePath: pfxFilePathV4,
		PfxPass:     pfxPass,
		CAFilePath:  caFilePath,
	}

	data, _, err := con.GetResourceAll()
	if err != nil {
		t.Fatal(err)
	}

	for _, tmp := range data.ResourceCIDRBlocks {
		t.Log(tmp)
	}

	//t.Log("--------------HTML--------------")

	//t.Log(html)
}
