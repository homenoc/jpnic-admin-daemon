package main

import (
	"database/sql"
	"fmt"
	"github.com/homenoc/jpnic-go"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"strconv"
	"time"
)

var caFilePath = "/Users/y-yoneda/github/homenoc/jpnic-go/cert/rootcacert_r3.cer"

//HomeNOC
var pfxPass = "homenoc"
var pfxFilePathV4 = "/Users/y-yoneda/github/homenoc/jpnic-go/cert/v4-openssl.p12"
var pfxFilePathV6 = "/Users/y-yoneda/github/homenoc/jpnic-go/cert/v6-openssl.p12"

var sqlitePath = "/Users/y-yoneda/github/homenoc/jpnic_gui/db.sqlite3"
var asNumber = 59105

func main() {
	con := jpnic.Config{
		URL:         "https://iphostmaster.nic.ad.jp/jpnic/certmemberlogin.do",
		PfxFilePath: pfxFilePathV4,
		PfxPass:     pfxPass,
		CAFilePath:  caFilePath,
	}

	var sqliteOption = "file:" + sqlitePath + "?cache=shared&mode=rwc&_journal_mode=WAL"
	now := time.Now()
	timeDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	db, err := sql.Open("sqlite3", sqliteOption)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	rows, err := db.Query("SELECT id, ip_address, address, address_en, recep_number FROM result_v4list WHERE get_date > $1 AND asn_id = $2", timeDate, asNumber)
	if err != nil {
		log.Fatal(err)
	}

	var list ResultV4List
	for rows.Next() {
		err = rows.Scan(&list.ID, &list.IPAddress, &list.Address, &list.AddressEn, &list.RecepNumber)
		if err != nil {
			log.Fatal(err)
		}

		if list.Address == "" && list.AddressEn == "" {
			fmt.Printf("ID: %d, IPAddress: %s Address: %s(%s),受付番号: %s\n", list.ID, list.IPAddress, list.Address, list.AddressEn, list.RecepNumber)
			break
		}
	}
	rows.Close()

	// イレギュラー処理
	// この場合は、住所/住所(English)情報,受付番号に空白文字を突っ込む
	if list.ID > 0 && list.RecepNumber == "" {
		fmt.Println("イレギュラー処理")
		upd, err := db.Prepare("UPDATE result_v4list SET get_date = ?, address = ?, address_en = ?, recep_number = ?, asn_id = ? WHERE id = ?")
		if err != nil {
			log.Fatal(err)
		}
		_, err = upd.Exec(time.Now(), "　", "　", "　", asNumber, list.ID)
		if err != nil {
			log.Fatal(err)
		}
		upd.Close()

		return
	}

	// イレギュラー処理
	if list.RecepNumber == "" {
		fmt.Println("データがありません")
		return
	}

	// 全体取得データがない場合
	if (list == ResultV4List{}) {
		data, _, err := con.SearchIPv4(jpnic.SearchIPv4{Myself: true, IsDetail: false})
		if err != nil {
			log.Fatal(err)
		}

		for _, tmp := range data {
			var id string
			layout := "2006/01/02"
			assignDate, _ := time.Parse(layout, tmp.AssignDate)

			ins, err := db.Prepare("INSERT INTO result_v4list (get_date, ip_address, size, network_name, assign_date, return_date, org, org_en, resource_admin_short, recep_number, deli_number, type, division, post_code, address, address_en, name_server, ds_record, notify_address, admin_jpnic_id, asn_id) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?) RETURNING id;")
			if err != nil {
				log.Fatal(err)
			}
			defer ins.Close()

			ins.QueryRow(time.Now(), tmp.IPAddress, tmp.Size, tmp.NetworkName, assignDate, tmp.ReturnDate, tmp.InfoDetail.Org, tmp.InfoDetail.OrgEn, tmp.InfoDetail.Ryakusho, tmp.RecepNo, tmp.DeliNo, "", "", tmp.InfoDetail.PostCode, tmp.InfoDetail.Address, tmp.InfoDetail.AddressEn, tmp.InfoDetail.NameServer, tmp.InfoDetail.DSRecord, tmp.InfoDetail.NotifyAddress, tmp.InfoDetail.AdminJPNICHandle, "59105").Scan(&id)
		}
	} else {
		// 同じ受付番号がないか確認
		var listIDs []string
		rows, err = db.Query("SELECT id FROM result_v4list WHERE get_date > $1 AND asn_id = $2 AND address = '' AND address_en = '' AND recep_number = $3", timeDate, asNumber, list.RecepNumber)
		if err != nil {
			log.Fatal(err)
		}

		for rows.Next() {
			var id string
			err = rows.Scan(&id)
			if err != nil {
				log.Fatal(err)
			}
			listIDs = append(listIDs, id)
		}
		rows.Close()

		// JPNIC Handle探索
		rows, err = db.Query("SELECT id,jpnic_handle,get_date FROM result_jpnichandle WHERE get_date > $1 AND asn_id = $2 AND is_ipv6 = $3", timeDate, asNumber, false)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		handles := make(map[string]int)
		var strHandles []string
		var handle JPNICHandle
		for rows.Next() {
			err = rows.Scan(&handle.ID, &handle.JPNICHandle, &handle.GetTime)
			if err != nil {
				log.Fatal(err)
			}

			handles[handle.JPNICHandle] = handle.ID
			strHandles = append(strHandles, handle.JPNICHandle)

			fmt.Printf("ID: %d(%s),Handle: %s\n", handle.ID, handle.GetTime, handle.JPNICHandle)
		}

		data, jpnicHandles, err := con.SearchIPv4(jpnic.SearchIPv4{
			Myself:    true,
			IsDetail:  true,
			Option1:   strHandles,
			IPAddress: list.IPAddress,
			RecepNo:   list.RecepNumber,
		})
		if err != nil {
			log.Fatal(err)
		}

		// jpnic_handle DBに追加処理
		if len(jpnicHandles) != 0 {
			for _, jpnicHandle := range jpnicHandles {

				var jpnicHandleID string

				layout := "2006/01/02 15:04"
				updateDate, _ := time.Parse(layout, jpnicHandle.UpdateDate)
				ins, err := db.Prepare("INSERT INTO result_jpnichandle (is_ipv6, get_date, jpnic_handle, name, name_en, email, org, org_en, division, division_en, tel, fax, update_date, asn_id) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?) RETURNING id;")
				if err != nil {
					log.Fatal(err)
				}
				ins.QueryRow(false, time.Now(), jpnicHandle.JPNICHandle, jpnicHandle.Name, jpnicHandle.NameEn, jpnicHandle.Email, jpnicHandle.Org, jpnicHandle.OrgEn, jpnicHandle.Division, jpnicHandle.DivisionEn, jpnicHandle.Tel, jpnicHandle.Fax, updateDate, asNumber).Scan(&jpnicHandleID)

				ins.Close()

				handles[jpnicHandle.JPNICHandle], _ = strconv.Atoi(jpnicHandleID)
			}
		}

		// result_v4list DBにUpdate処理
		for _, listID := range listIDs {
			upd, err := db.Prepare("UPDATE result_v4list SET get_date = ?, org = ?, org_en = ?, post_code = ?, address = ?, address_en = ?, name_server = ?, ds_record = ?, notify_address = ?, admin_jpnic_id = ?, asn_id = ? WHERE id = ?")
			if err != nil {
				log.Fatal(err)
			}
			_, err = upd.Exec(time.Now(), data[0].InfoDetail.Org, data[0].InfoDetail.OrgEn, data[0].InfoDetail.PostCode, data[0].InfoDetail.Address, data[0].InfoDetail.AddressEn, data[0].InfoDetail.NameServer, data[0].InfoDetail.DSRecord, data[0].InfoDetail.NotifyAddress, handles[data[0].InfoDetail.AdminJPNICHandle], asNumber, listID)
			if err != nil {
				log.Fatal(err)
			}
			upd.Close()

			// JPNIC技術連絡先をDBに登録
			//for _, techHandle := range data[0].InfoDetail.TechJPNICHandle {

			ins, err := db.Prepare("INSERT INTO result_v4list_tech_jpnic (v4list_id, jpnichandle_id) VALUES(?,?)")
			if err != nil {
				log.Fatal(err)
			}

			ins.Exec(listID, handles[data[0].InfoDetail.TechJPNICHandle])

			ins.Close()
			//}
		}
	}
}
