package jpnic

import (
	"database/sql"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"strings"
	"time"
)

var userAgent = "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:91.0) Gecko/20100101 Firefox/91.0"
var contentType = "application/x-www-form-urlencoded"
var baseURL = "https://iphostmaster.nic.ad.jp"

type Config struct {
	URL       string
	CA        []byte
	P12Base64 string
	P12Pass   string
	DB        *sql.DB
}

func (c *Config) SearchIPv4(search SearchIPv4) (*ResultSearchIPv4, error) {
	client, menuURL, err := c.initAccess("登録情報検索(IPv4)")
	if err != nil {
		return nil, err
	}

	r := request{
		Client:      client,
		URL:         baseURL + "/jpnic/" + menuURL,
		Body:        "",
		UserAgent:   userAgent,
		ContentType: contentType,
	}

	resp, err := r.get()
	if err != nil {
		return nil, err
	}

	resBody, _, err := readShiftJIS(resp.Body)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resBody))
	if err != nil {
		return nil, err
	}

	submitURL, isExists := doc.Find("form").Attr("action")
	if !isExists {
		return nil, fmt.Errorf("submit URLが取得できませんでした")
	}
	submitID, isExists := doc.Find("form").Find("input").Attr("value")
	if !isExists {
		return nil, fmt.Errorf("inputフォームのIDが取得できませんでした")
	}

	var requestStr string
	// 管理者略称
	var resceAdmSnm string

	if search.Ryakusho == "" {
		// 自身のAS
		doc.Find("form").Find("ul").Find("table").Children().Find("table").Children().Find("input").Each(func(index int, s *goquery.Selection) {
			var name string
			name, isExists = s.Attr("name")
			if name == "resceAdmSnm" {
				resceAdmSnm, isExists = s.Attr("value")
			}
		})
		if !isExists {
			return nil, fmt.Errorf("資源管理者略称が見つかりませんでした")
		}
	} else {
		resceAdmSnm = search.Ryakusho
	}

	requestStr = "destdisp=" + submitID
	requestStr += "&ipaddr=" + search.IPAddress
	requestStr += "&sizeS=" + search.SizeStart
	requestStr += "&sizeE=" + search.SizeEnd
	requestStr += "&netwrkName=" + search.NetworkName
	requestStr += "&regDateS=" + search.RegStart
	requestStr += "&regDateE=" + search.RegEnd
	requestStr += "&rtnDateS=" + search.ReturnStart
	requestStr += "&rtnDateE=" + search.ReturnEnd
	requestStr += "&organizationName=" + search.Org
	requestStr += "&resceAdmSnm=" + resceAdmSnm
	requestStr += "&recepNo=" + search.RecepNo
	requestStr += "&deliNo=" + search.DeliNo
	requestStr += "&ipaddrKindPa=" + getSearchBoolean(search.IsPA)
	requestStr += "&regKindAllo=" + getSearchBoolean(search.IsAllocate)
	requestStr += "&regKindEvent=" + getSearchBoolean(search.IsAssignInfra)
	requestStr += "&regKindUser=" + getSearchBoolean(search.IsAssignUser)
	requestStr += "&regKindSubA=" + getSearchBoolean(search.IsSubAllocate)
	requestStr += "&ipaddrKindPiHistorical=" + getSearchBoolean(search.IsHistoricalPI)
	requestStr += "&ipaddrKindPiSpecial=" + getSearchBoolean(search.IsSpecialPI)
	requestStr += "&action=　検索　"

	// utf-8 => shift-jis
	reqBody, _, err := toShiftJIS(requestStr)
	if err != nil {
		return nil, err
	}

	r = request{
		Client:      client,
		URL:         baseURL + submitURL,
		Body:        reqBody,
		UserAgent:   userAgent,
		ContentType: contentType,
	}

	resp, err = r.post()
	if err != nil {
		return nil, err
	}

	resBody, _, err = readShiftJIS(resp.Body)
	if err != nil {
		return nil, err
	}

	doc, err = goquery.NewDocumentFromReader(strings.NewReader(resBody))
	if err != nil {
		return nil, err
	}

	var infos []InfoIPv4
	var info InfoIPv4
	var jpnicHandles []JPNICHandleDetail
	allCounter := 0
	index := 0
	isJPNICHandleExist := make(map[string]int)

	// option1 function
	for _, handle := range search.Option1 {
		isJPNICHandleExist[handle] = 0
	}

	doc.Find("table").Children().Find("td").Each(func(_ int, tableHtml *goquery.Selection) {
		className, _ := tableHtml.Attr("class")
		if className != "dataRow_mnt04" {
			return
		}
		dataStr := strings.TrimSpace(tableHtml.Text())
		switch index {
		case 0:
			info.IPAddress = dataStr
			info.DetailLink, _ = tableHtml.Find("a").Attr("href")
		case 1:
			info.Size = dataStr
		case 2:
			info.NetworkName = dataStr
		case 3:
			info.AssignDate = dataStr
		case 4:
			info.ReturnDate = dataStr
		case 5:
			info.OrgName = dataStr
		case 6:
			info.Ryakusho = dataStr
		case 7:
			info.RecepNo = dataStr
		case 8:
			info.DeliNo = dataStr
		case 9:
			info.Type = dataStr
		case 10:
			info.Division = dataStr
			// 詳細情報の取得
			if search.IsDetail && allCounter != 0 {
				//log.Println("==========")
				time.Sleep(1 * time.Second)
				//log.Println("req1")
				info.InfoDetail, err = getInfoDetail(client, info.DetailLink)
				if err != nil {

					return
				}
				// Admin JPNIC Handle
				if _, ok := isJPNICHandleExist[info.InfoDetail.TechJPNICHandle]; !ok {
					// 一定時間停止
					time.Sleep(1 * time.Second)
					//log.Println("req2")

					jpnic, err := getJPNICHandle(client, info.InfoDetail.AdminJPNICHandleLink)
					if err != nil {
						return
					}
					jpnicHandles = append(jpnicHandles, jpnic)
					isJPNICHandleExist[info.InfoDetail.TechJPNICHandle] = 0
				}
				// Tech JPNIC Handle
				if _, ok := isJPNICHandleExist[info.InfoDetail.AdminJPNICHandle]; !ok {
					//log.Println("req3")
					// 一定時間停止
					time.Sleep(1 * time.Second)

					jpnic, err := getJPNICHandle(client, info.InfoDetail.TechJPNICHandleLink)
					if err != nil {
						return
					}
					jpnicHandles = append(jpnicHandles, jpnic)
					isJPNICHandleExist[info.InfoDetail.AdminJPNICHandle] = 0
				}
				//log.Printf("count: %d\n", allCounter)
				//log.Println("==========")
			}
			index = -1
			if allCounter != 0 {
				infos = append(infos, info)
				info = InfoIPv4{}
			}
			allCounter++
		}
		index++
	})

	return &ResultSearchIPv4{
		IsOverList:        strings.Contains(resBody, "該当する情報が1000件を超えました (1000件まで表示します)"),
		InfoIPv4:          infos,
		JPNICHandleDetail: jpnicHandles,
	}, nil
}

func (c *Config) SearchIPv6(search SearchIPv6) (*ResultSearchIPv6, error) {
	client, menuURL, err := c.initAccess("登録情報検索(IPv6)")
	if err != nil {
		return nil, err
	}

	r := request{
		Client:      client,
		URL:         baseURL + "/jpnic/" + menuURL,
		Body:        "",
		UserAgent:   userAgent,
		ContentType: contentType,
	}

	resp, err := r.get()
	if err != nil {
		return nil, err
	}

	resBody, _, err := readShiftJIS(resp.Body)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resBody))
	if err != nil {
		return nil, err
	}

	submitURL, isExists := doc.Find("form").Attr("action")
	if !isExists {
		return nil, fmt.Errorf("submit URLが取得できませんでした")
	}
	submitID, isExists := doc.Find("form").Find("input").Attr("value")
	if !isExists {
		return nil, fmt.Errorf("inputフォームのIDが取得できませんでした")
	}

	var requestStr string
	// 管理者略称
	var resceAdmSnm string

	if search.Ryakusho == "" {
		// 自身のAS
		doc.Find("form").Find("ul").Find("table").Children().Find("table").Children().Find("input").Each(func(index int, s *goquery.Selection) {
			var name string
			name, isExists = s.Attr("name")
			if name == "resceAdmSnm" {
				resceAdmSnm, isExists = s.Attr("value")
			}
		})
		if !isExists {
			return nil, fmt.Errorf("資源管理者略称が見つかりませんでした")
		}
	} else {
		resceAdmSnm = search.Ryakusho
	}

	requestStr = "destdisp=" + submitID
	requestStr += "&ipaddr=" + search.IPAddress
	requestStr += "&sizeS=" + search.SizeStart
	requestStr += "&sizeE=" + search.SizeEnd
	requestStr += "&netwrkName=" + search.NetworkName
	requestStr += "&regDateS=" + search.RegStart
	requestStr += "&regDateE=" + search.RegEnd
	requestStr += "&rtnDateS=" + search.ReturnStart
	requestStr += "&rtnDateE=" + search.ReturnEnd
	requestStr += "&organizationName=" + search.Org
	requestStr += "&resceAdmSnm=" + resceAdmSnm
	requestStr += "&recepNo=" + search.RecepNo
	requestStr += "&deliNo=" + search.DeliNo
	requestStr += "&regKindAllo=" + getSearchBoolean(search.IsAllocate)
	requestStr += "&regKindEvent=" + getSearchBoolean(search.IsAssignInfra)
	requestStr += "&regKindUser=" + getSearchBoolean(search.IsAssignUser)
	requestStr += "&regKindSubA=" + getSearchBoolean(search.IsSubAllocate)
	requestStr += "&action=%81%40%8C%9F%8D%F5%81%40"

	// utf-8 => shift-jis
	reqBody, _, err := toShiftJIS(requestStr)
	if err != nil {
		return nil, err
	}

	r = request{
		Client:      client,
		URL:         baseURL + submitURL,
		Body:        reqBody,
		UserAgent:   userAgent,
		ContentType: contentType,
	}

	resp, err = r.post()
	if err != nil {
		return nil, err
	}

	resBody, _, err = readShiftJIS(resp.Body)
	if err != nil {
		return nil, err
	}

	doc, err = goquery.NewDocumentFromReader(strings.NewReader(resBody))
	if err != nil {
		return nil, err
	}

	var infos []InfoIPv6
	var info InfoIPv6
	var jpnicHandles []JPNICHandleDetail
	allCounter := 0
	index := 0
	isJPNICHandleExist := make(map[string]int)

	// option1 function
	for _, handle := range search.Option1 {
		isJPNICHandleExist[handle] = 0
	}

	doc.Find("table").Children().Find("td").Each(func(_ int, tableHtml *goquery.Selection) {
		className, _ := tableHtml.Attr("class")
		if className != "dataRow_mnt04" {
			return
		}
		dataStr := strings.TrimSpace(tableHtml.Text())
		switch index {
		case 0:
			info.IPAddress = dataStr
			info.DetailLink, _ = tableHtml.Find("a").Attr("href")
		case 1:
			info.NetworkName = dataStr
		case 2:
			info.AssignDate = dataStr
		case 3:
			info.ReturnDate = dataStr
		case 4:
			info.OrgName = dataStr
		case 5:
			info.Ryakusho = dataStr
		case 6:
			info.RecepNo = dataStr
		case 7:
			info.DeliNo = dataStr
		case 8:
			info.KindID = dataStr
			// 詳細情報の取得
			if search.IsDetail && allCounter != 0 {
				//log.Println("==========")
				time.Sleep(1 * time.Second)
				//log.Println("req1")
				info.InfoDetail, err = getInfoDetail(client, info.DetailLink)
				if err != nil {

					return
				}
				// Admin JPNIC Handle
				if _, ok := isJPNICHandleExist[info.InfoDetail.TechJPNICHandle]; !ok {
					// 一定時間停止
					time.Sleep(1 * time.Second)
					//log.Println("req2")

					jpnic, err := getJPNICHandle(client, info.InfoDetail.AdminJPNICHandleLink)
					if err != nil {
						return
					}
					jpnicHandles = append(jpnicHandles, jpnic)
					isJPNICHandleExist[info.InfoDetail.TechJPNICHandle] = 0
				}
				// Tech JPNIC Handle
				if _, ok := isJPNICHandleExist[info.InfoDetail.AdminJPNICHandle]; !ok {
					//log.Println("req3")
					// 一定時間停止
					time.Sleep(1 * time.Second)

					jpnic, err := getJPNICHandle(client, info.InfoDetail.TechJPNICHandleLink)
					if err != nil {
						return
					}
					jpnicHandles = append(jpnicHandles, jpnic)
					isJPNICHandleExist[info.InfoDetail.AdminJPNICHandle] = 0
				}
				//log.Printf("count: %d\n", allCounter)
				//log.Println("==========")
			}
			index = -1
			if allCounter != 0 {
				infos = append(infos, info)
				info = InfoIPv6{}
			}
			allCounter++
		}
		index++
	})

	return &ResultSearchIPv6{
		IsOverList:        strings.Contains(resBody, "該当する情報が1000件を超えました (1000件まで表示します)"),
		InfoIPv6:          infos,
		JPNICHandleDetail: jpnicHandles,
	}, nil
}
