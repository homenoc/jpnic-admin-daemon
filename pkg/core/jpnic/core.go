package jpnic

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func (c *Config) GetResourceAll() (ResourceAllInfo, string, error) {
	var info ResourceAllInfo
	var html string
	client, menuURL, err := c.initAccess("資源管理者情報")
	if err != nil {
		return info, html, err
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
		return info, html, err
	}

	resBody, _, err := readShiftJIS(resp.Body)
	if err != nil {
		return info, html, err
	}

	html = resBody

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resBody))
	if err != nil {
		return info, html, err
	}

	re := regexp.MustCompile(`\(([^}]*)\)`)
	err = nil

	var title string
	cidrBlockSegment := false
	var cidrBlock ResourceCIDRBlock
	var rsAddressLists []ResourceAddressList
	var rsAddressList ResourceAddressList

	doc.Find("table").Children().Find("table").Children().Find("table").Children().Find("table").Children().Find("td").Each(func(_ int, tableHtml1 *goquery.Selection) {
		dataStr := strings.TrimSpace(tableHtml1.Text())
		index := tableHtml1.Index()

		switch index {
		case 0:
			cidrBlockSegment = false
			title = dataStr
			addressDetailURL, addressExists := tableHtml1.Find("a").Attr("href")
			if addressExists {
				cidrBlockSegment = strings.Contains(addressDetailURL, "entryinfo")
				splitAddress := strings.Split(dataStr, "(")
				tmpAddress := strings.Replace(splitAddress[0], "\n", "", 1)
				address := strings.Replace(tmpAddress, "	", "", 3)
				cidrBlock.Address = strings.TrimSpace(address)
				rsAddressList.Address = strings.TrimSpace(address)
				cidrBlock.URL = addressDetailURL
				rsAddressList.URL = addressDetailURL
			}
		case 1:
			switch title {
			case "資源管理者番号":
				info.ResourceManagerInfo.ResourceManagerNo = dataStr
			case "資源管理者略称":
				info.ResourceManagerInfo.Ryakusyo = dataStr
			case "管理組織名":
				info.ResourceManagerInfo.Org = dataStr
			case "Organization":
				info.ResourceManagerInfo.OrgEn = dataStr
			case "郵便番号":
				info.ResourceManagerInfo.ZipCode = dataStr
			case "住所":
				info.ResourceManagerInfo.Address = dataStr
			case "Address":
				info.ResourceManagerInfo.AddressEn = dataStr
			case "電話番号":
				info.ResourceManagerInfo.Tel = dataStr
			case "FAX番号":
				info.ResourceManagerInfo.Fax = dataStr
			case "資源管理責任者":
				info.ResourceManagerInfo.ResourceManagementManager = dataStr
			case "連絡担当窓口":
				info.ResourceManagerInfo.ContactPerson = dataStr
			case "一般問い合わせ窓口":
				info.ResourceManagerInfo.Inquiry = dataStr
			case "資源管理者通知アドレス":
				info.ResourceManagerInfo.NotifyMail = dataStr
			case "アサインメントウィンドウサイズ":
				info.ResourceManagerInfo.AssigmentWindowSize = dataStr
			case "管理開始日":
				info.ResourceManagerInfo.ManagementStartDate = dataStr
			case "管理終了日":
				info.ResourceManagerInfo.ManagementEndDate = dataStr
			case "最終更新日":
				info.ResourceManagerInfo.UpdateDate = dataStr
			default:
				if cidrBlockSegment {
					cidrBlock.AssignDate = dataStr
				}
			}
		case 2:
			switch title {
			case "総利用率":
				match := re.FindStringSubmatch(dataStr)
				if len(match) == 0 {
					err = fmt.Errorf("データが存在しません")
					break
				}
				splitAddress := strings.Split(match[1], "/")

				info.UsedAddress, err = strconv.ParseUint(splitAddress[0], 10, 32)
				if err != nil {
					break
				}
				info.AllAddress, err = strconv.ParseUint(splitAddress[1], 10, 32)
				if err != nil {
					break
				}

				info.UtilizationRatio, err = strconv.ParseFloat(dataStr[:strings.Index(dataStr, "%")], 16)
				if err != nil {
					break
				}
			case "ＡＤ　ｒａｔｉｏ":
				log.Println(strconv.Itoa(index) + ": " + dataStr)

				info.ADRatio, err = strconv.ParseFloat(dataStr, 16)
				if err != nil {
					break
				}
			default:
				if cidrBlockSegment {
					match := re.FindStringSubmatch(dataStr)
					if len(match) == 0 {
						err = fmt.Errorf("データが存在しません")
						break
					}
					splitAddress := strings.Split(match[1], "/")

					cidrBlock.UsedAddress, err = strconv.ParseUint(splitAddress[0], 10, 32)
					if err != nil {
						break
					}
					cidrBlock.AllAddress, err = strconv.ParseUint(splitAddress[1], 10, 32)
					if err != nil {
						break
					}

					cidrBlock.UtilizationRatio, err = strconv.ParseFloat(dataStr[:strings.Index(dataStr, "%")], 16)
					if err != nil {
						break
					}
				}
			}
		}
		if cidrBlockSegment && index == 2 {
			info.ResourceCIDRBlock = append(info.ResourceCIDRBlock, cidrBlock)
			rsAddressLists = append(rsAddressLists, rsAddressList)
		}
	})

	if err != nil {
		return info, html, err
	}

	var infos *[]InfoDetail
	infos = &[]InfoDetail{}

	_, err = Loop(&r, rsAddressLists, infos)
	if err != nil {
		return info, html, err
	}

	log.Println(infos)
	file, _ := json.MarshalIndent(infos, "", " ")
	_ = ioutil.WriteFile("./test1.json", file, 0644)

	return info, "", nil
}

func Loop(r *request, addressList []ResourceAddressList, infos *[]InfoDetail) (*ResourceAddressList, error) {
	for _, segment := range addressList {
		// 割振を除外している理由として、
		if segment.Status == "未割当" || segment.Status == "割振" {
			break
		}

		ok := true

		for ok {
			ok = false

			// sleep process
			time.Sleep(time.Second * 5)

			r.URL = baseURL + segment.URL
			resp, err := r.get()
			if err != nil {
				log.Println(err)
				if strings.Contains(err.Error(), "i/o timeout") {
					ok = true
					continue
				} else {
					return nil, err
				}
			}

			resBody, _, err := readShiftJIS(resp.Body)
			if err != nil {
				return nil, err
			}

			addressInfo, addressList, err := Search(resBody)
			if err != nil {
				return nil, err
			}

			log.Println(addressInfo)
			*infos = append(*infos, addressInfo)
			Loop(r, addressList, infos)
		}
	}

	return nil, nil
}

func Search(resBody string) (InfoDetail, []ResourceAddressList, error) {
	var addressInfo InfoDetail
	var rsAddressTmpLists []ResourceAddressList
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resBody))
	if err != nil {
		return addressInfo, nil, err
	}

	// アドレス一覧の取得
	var rsAddressTmpList ResourceAddressList

	dataType := 1 // 1:割振アドレスブロック, 2:SUBA登録ブロック, 3:アドレスリスト
	doc.Find("ul").Children().Find("table").Children().Find("td").Each(func(index int, tableHtml1 *goquery.Selection) {
		dataStr := strings.TrimSpace(tableHtml1.Text())
		tdIndex := tableHtml1.Index()
		//log.Println(isAddressList, addressListIndex)

		// 割振アドレスブロック判別処理

		// 重複処理
		isDuplicate := false
		tableHtml1.Find("table").Children().Each(func(index int, tableHtml2 *goquery.Selection) {
			dataStr2 := strings.TrimSpace(tableHtml2.Text())
			if dataStr == dataStr2 {
				isDuplicate = true
			}
		})
		if isDuplicate {
			return
		}

		// 割振アドレスブロックのタイトル除外処理
		if dataStr == "割振アドレスブロック" {
			return
		}

		// SUBA登録ブロックのタイトル除外処理
		if dataType == 1 && dataStr == "SUBA登録ブロック" {
			dataType = 2
			return
		}

		// アドレスリスト判別処理 & タイトルの除外処理
		if (dataType == 1 || dataType == 2) &&
			tdIndex == 3 && dataStr == "利用率" && strings.TrimSpace(tableHtml1.Prev().Text()) == "利用状況" {
			dataType = 3
			return
		} else if dataStr == "IPアドレス" || dataStr == "ホスト数" || dataStr == "利用状況" || dataStr == "利用率" {
			return
		}

		switch dataType {
		case 1:
		case 2:
		case 3:
			switch tdIndex {
			case 0:
				rsAddressTmpList = ResourceAddressList{}
				rsAddressTmpList.Address = dataStr
				rsAddressTmpList.URL, _ = tableHtml1.Find("a").Attr("href")
			case 1:
				rsAddressTmpList.Range = dataStr
			case 2:
				quantity, _ := strconv.Atoi(dataStr)
				rsAddressTmpList.Quantity = uint(quantity)
			case 3:
				rsAddressTmpList.Status = dataStr
			case 4:
				if dataStr == "" {
					rsAddressTmpLists = append(rsAddressTmpLists, rsAddressTmpList)
					break
				}
				re := regexp.MustCompile(`\(([^}]*)\)`)
				match := re.FindStringSubmatch(dataStr)
				splitAddress := strings.Split(match[1], "/")

				rsAddressTmpList.UsedAddress, err = strconv.ParseUint(splitAddress[0], 10, 32)
				if err != nil {
					break
				}
				rsAddressTmpList.AllAddress, err = strconv.ParseUint(splitAddress[1], 10, 32)
				if err != nil {
					break
				}

				rsAddressTmpList.UtilizationRatio, err = strconv.ParseFloat(dataStr[:strings.Index(dataStr, "%")], 16)
				if err != nil {
					break
				}
				rsAddressTmpLists = append(rsAddressTmpLists, rsAddressTmpList)
			}

		}
		//log.Println(dataType, index, tdIndex, dataStr)

	})

	//log.Println("============")

	// ネットワーク情報の取得
	doc.Find("table").Children().Find("table").Children().Find("table").Children().Find("table").Children().Find("td").Each(func(_ int, tableHtml1 *goquery.Selection) {
		dataStr := strings.TrimSpace(tableHtml1.Text())
		beforeDataStr := strings.TrimSpace(tableHtml1.Prev().Text())

		switch beforeDataStr {
		case "IPネットワークアドレス":
			addressInfo.IPAddress = dataStr
		case "資源管理者略称":
			addressInfo.Ryakusho = dataStr
		case "アドレス種別":
			addressInfo.Type = dataStr
		case "インフラ・ユーザ区分":
			addressInfo.InfraUserKind = dataStr
		case "ネットワーク名":
			addressInfo.NetworkName = dataStr
		case "組織名":
			addressInfo.Org = dataStr
		case "Organization":
			addressInfo.OrgEn = dataStr
		case "郵便番号":
			addressInfo.PostCode = dataStr
		case "住所":
			addressInfo.Address = dataStr
		case "Address":
			addressInfo.AddressEn = dataStr
		case "管理者連絡窓口":
			addressInfo.AdminJPNICHandleLink, _ = tableHtml1.Find("a").Attr("href")
			addressInfo.AdminJPNICHandle = dataStr
		case "技術連絡担当者":
			addressInfo.TechJPNICHandleLink, _ = tableHtml1.Find("a").Attr("href")
			addressInfo.TechJPNICHandle = dataStr
		case "Abuse":
			addressInfo.Abuse = dataStr
		case "通知アドレス":
			addressInfo.NotifyAddress = dataStr
		case "ネームサーバ":
			addressInfo.NameServer = dataStr
		case "DSレコード":
			addressInfo.DSRecord = dataStr
		case "審議番号":
			addressInfo.DeliNo = dataStr
		case "受付番号":
			addressInfo.RecepNo = dataStr
		case "割振年月日", "割当年月日":
			addressInfo.AssignDate = dataStr
		case "返却年月日":
			addressInfo.ReturnDate = dataStr
		case "最終更新":
			addressInfo.UpdateDate = dataStr
		default:
		}
		//if cidrBlockSegment && index == 2 {
		//	info.ResourceCIDRBlock = append(info.ResourceCIDRBlock, cidrBlock)
		//}
	})
	return addressInfo, rsAddressTmpLists, nil
}
