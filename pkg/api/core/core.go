package core

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/homenoc/jpnic-admin-daemon/pkg/api/core/jpnic"
	"github.com/homenoc/jpnic-admin-daemon/pkg/api/core/tool/config"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net"
	"strconv"
	"time"
)

var dataCerts map[int]*JPNICCert

func Start() {
	dataCerts = map[int]*JPNICCert{}
	var getConfTimer uint = 5
	getConfTick := time.NewTicker(time.Duration(getConfTimer) * time.Second)

	go func() {
		var getInfoTimer uint = 5
		getInfoTick := time.NewTicker(time.Duration(getInfoTimer) * time.Second)

		log.Printf("start \n")
		for {
			select {
			case <-getInfoTick.C:
				log.Printf("start(getInfoTick) \n")
				now := time.Now()
				for id, dataValue := range dataCerts {
					if dataValue.RenewDate.Unix() <= now.Unix() {
						log.Println("", dataValue.RenewDate, "<", now)
						renewDate := now.Add(time.Minute * time.Duration(dataValue.CollectionInterval))
						dataCerts[id].RenewDate = renewDate
						if dataValue.IsADA {
							go GetJPNIC(*dataValue)
						}
					}
				}
			}
		}
	}()

	log.Printf("start \n")
	for {
		select {
		case <-getConfTick.C:
			config.GetCA()
			log.Println("get Info Tick")

			var db *sql.DB
			db, err := sql.Open(config.ConfDatabase.Driver, config.ConfDatabase.Option)
			if err != nil {
				log.Println(err)
			}

			defer db.Close()

			certRows, err := db.Query("SELECT * FROM jpnic_admin_jpnic")
			if err != nil {
				log.Println(err)
			}
			defer certRows.Close()

			var jpnicCert JPNICCert
			var jpnicCerts []JPNICCert

			for certRows.Next() {
				err = certRows.Scan(
					&jpnicCert.ID,
					&jpnicCert.Name,
					&jpnicCert.IsActive,
					&jpnicCert.IsIPv6,
					&jpnicCert.IsADA,
					&jpnicCert.CollectionInterval,
					&jpnicCert.ASN,
					&jpnicCert.P12Base64,
					&jpnicCert.P12Pass,
				)
				if err != nil {
					log.Println(err)
				}
				jpnicCerts = append(jpnicCerts, jpnicCert)
			}

			for _, jpn := range jpnicCerts {
				dataValue, isExists := dataCerts[jpn.ID]
				if isExists {
					// 書き換え
					if jpn.IsIPv6 != dataValue.IsIPv6 || jpn.IsADA != dataValue.IsADA ||
						jpn.CollectionInterval != dataValue.CollectionInterval ||
						jpn.ASN != dataValue.ASN || jpn.P12Base64 != dataValue.P12Base64 ||
						jpn.P12Pass != dataValue.P12Pass {
						log.Println("[replace] getting data: ", jpn.Name)
						dataCerts[jpn.ID] = &JPNICCert{
							ID:                 jpn.ID,
							Name:               jpn.Name,
							IsActive:           jpn.IsActive,
							IsIPv6:             jpn.IsIPv6,
							IsADA:              jpn.IsADA,
							CollectionInterval: jpn.CollectionInterval,
							ASN:                jpn.ASN,
							P12Base64:          jpn.P12Base64,
							P12Pass:            jpn.P12Pass,
							RenewDate:          dataValue.RenewDate.Add(time.Minute * time.Duration(jpn.CollectionInterval)),
						}
					}
				} else {
					now := time.Now()
					renewDate := now.Add(time.Minute * time.Duration(jpn.CollectionInterval))
					// Debug
					renewDate = now
					log.Println("[new] getting data: ", jpn.Name)
					dataCerts[jpn.ID] = &JPNICCert{
						ID:                 jpn.ID,
						Name:               jpn.Name,
						IsActive:           jpn.IsActive,
						IsIPv6:             jpn.IsIPv6,
						IsADA:              jpn.IsADA,
						CollectionInterval: jpn.CollectionInterval,
						ASN:                jpn.ASN,
						P12Base64:          jpn.P12Base64,
						P12Pass:            jpn.P12Pass,
						RenewDate:          renewDate,
					}
				}
			}

			for dataKey := range dataCerts {
				isJPNICCert := false
				for _, jpn := range jpnicCerts {
					if dataKey == jpn.ID {
						isJPNICCert = true
						break
					}
				}
				if !isJPNICCert {
					delete(dataCerts, dataKey)
				}
			}
		}
	}
}

func GetJPNIC(cert JPNICCert) {
	now := time.Now().UTC()
	timeDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	var db *sql.DB
	db, err := sql.Open(config.ConfDatabase.Driver, config.ConfDatabase.Option)
	if err != nil {
		log.Println(err)
		return
	}

	defer db.Close()

	jpnicConfig := jpnic.Config{
		URL:       "https://iphostmaster.nic.ad.jp/jpnic/certmemberlogin.do",
		CA:        config.CA,
		P12Base64: cert.P12Base64,
		P12Pass:   cert.P12Pass,
	}

	if cert.IsIPv6 {
		rows, err := db.Query("SELECT id, ip_address, address, address_en, recep_number FROM result_v6list WHERE get_date > $1 AND asn_id = $2", timeDate, cert.ASN)
		if err != nil {
			log.Println(err)
			return
		}

		var list ResultV6List
		for rows.Next() {
			err = rows.Scan(&list.ID, &list.IPAddress, &list.Address, &list.AddressEn, &list.RecepNumber)
			if err != nil {
				log.Println(err)
				return
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
			upd, err := db.Prepare("UPDATE result_v6list SET get_date = ?, address = ?, address_en = ?, recep_number = ?, asn_id = ? WHERE id = ?")
			if err != nil {
				log.Println(err)
				return
			}
			_, err = upd.Exec(time.Now().UTC(), "　", "　", "　", cert.ASN, list.ID)
			if err != nil {
				log.Println(err)
				return
			}
			upd.Close()

			return
		}
		log.Println("========================================================")
		log.Println(list)

		// 全体取得データがない場合
		if (list == ResultV6List{}) {
			// 1000件以上の場合も取得
			var infos []jpnic.InfoIPv6
			isOverList := true
			addressRange := ""

			for isOverList {
				filter := jpnic.SearchIPv6{
					IsDetail: false,
					Option1:  nil,
				}
				if addressRange != "" {
					filter.IPAddress = addressRange
				}
				data, err := jpnicConfig.SearchIPv6(filter)
				log.Println(err)
				if err != nil {
					log.Println(err)
					continue
				}

				isOverList = data.IsOverList

				// Todo: 後ほど実装(1000件を超える場合)
				//if isOverList {
				//	lastIPAddress, _, err := net.ParseCIDR(data.InfoIPv6[len(data.InfoIPv6)-1].IPAddress)
				//	if err != nil {
				//		log.Println(err)
				//	}
				//
				//	addressRange = lastIPAddress.String() + "-255.255.255.255"
				//}

				infos = append(infos, data.InfoIPv6...)
			}

			log.Println("==================-before range infos")

			for _, tmp := range infos {
				var id string
				layout := "2006/01/02"
				assignDate, _ := time.Parse(layout, tmp.AssignDate)

				ins, err := db.Prepare("INSERT INTO result_v6list (get_date, ip_address, network_name, assign_date, return_date, org, org_en, resource_admin_short, recep_number, deli_number, division, post_code, address, address_en, name_server, ds_record, notify_address, admin_jpnic_id, asn_id) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?) RETURNING id;")
				if err != nil {
					log.Println(err)
					return
				}
				defer ins.Close()

				ins.QueryRow(
					time.Now().UTC(),
					tmp.IPAddress,
					tmp.NetworkName,
					assignDate,
					tmp.ReturnDate,
					tmp.InfoDetail.Org,
					tmp.InfoDetail.OrgEn,
					tmp.InfoDetail.Ryakusho,
					tmp.RecepNo,
					tmp.DeliNo,
					"",
					tmp.InfoDetail.PostCode,
					tmp.InfoDetail.Address,
					tmp.InfoDetail.AddressEn,
					tmp.InfoDetail.NameServer,
					tmp.InfoDetail.DSRecord,
					tmp.InfoDetail.NotifyAddress,
					tmp.InfoDetail.AdminJPNICHandle,
					cert.ASN,
				).Scan(&id)
			}
		} else {
			// イレギュラー処理
			if list.RecepNumber == "" {
				fmt.Println("データがありません")
				return
			}

			// 同じ受付番号がないか確認
			var listIDs []string
			//log.Println(timeDate, jpnicCert.ASN, list.RecepNumber)
			rows, err = db.Query("SELECT id FROM result_v6list WHERE get_date > $1 AND asn_id = $2 AND address = '' AND address_en = '' AND recep_number = $3", timeDate, cert.ASN, list.RecepNumber)
			if err != nil {
				log.Println(err)
				return
			}

			for rows.Next() {
				var id string
				err = rows.Scan(&id)
				if err != nil {
					log.Println(err)
					return
				}
				listIDs = append(listIDs, id)
			}
			rows.Close()

			// JPNIC Handle探索
			rows, err = db.Query("SELECT id,jpnic_handle,get_date FROM result_jpnichandle WHERE get_date > $1 AND asn_id = $2 AND is_ipv6 = $3", timeDate, cert.ASN, true)
			if err != nil {
				log.Println(err)
				return
			}
			defer rows.Close()

			handles := make(map[string]int)
			var strHandles []string
			var handle JPNICHandle
			for rows.Next() {
				err = rows.Scan(&handle.ID, &handle.JPNICHandle, &handle.GetTime)
				if err != nil {
					log.Println(err)
					return
				}

				handles[handle.JPNICHandle] = handle.ID
				strHandles = append(strHandles, handle.JPNICHandle)

				fmt.Printf("ID: %d(%s),Handle: %s\n", handle.ID, handle.GetTime, handle.JPNICHandle)
			}

			data, err := jpnicConfig.SearchIPv6(jpnic.SearchIPv6{
				IsDetail:  true,
				Option1:   strHandles,
				IPAddress: list.IPAddress,
				RecepNo:   list.RecepNumber,
			})
			if err != nil {
				log.Println(err)
				return
			}

			// jpnic_handle DBに追加処理
			if data != nil && len(data.JPNICHandleDetail) != 0 {
				for _, jpnicHandle := range data.JPNICHandleDetail {
					var jpnicHandleID string

					layout := "2006/01/02 15:04"
					updateDate, _ := time.Parse(layout, jpnicHandle.UpdateDate)
					ins, err := db.Prepare("INSERT INTO result_jpnichandle (is_ipv6, get_date, jpnic_handle, name, name_en, email, org, org_en, division, division_en, tel, fax, update_date, asn_id) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?) RETURNING id;")
					if err != nil {
						log.Println(err)
						return
					}
					ins.QueryRow(
						true,
						time.Now().UTC(),
						jpnicHandle.JPNICHandle,
						jpnicHandle.Name,
						jpnicHandle.NameEn,
						jpnicHandle.Email,
						jpnicHandle.Org,
						jpnicHandle.OrgEn,
						jpnicHandle.Division,
						jpnicHandle.DivisionEn,
						jpnicHandle.Tel,
						jpnicHandle.Fax,
						updateDate,
						cert.ASN,
					).Scan(&jpnicHandleID)

					ins.Close()

					handles[jpnicHandle.JPNICHandle], _ = strconv.Atoi(jpnicHandleID)
				}
			}

			// result_v6list DBにUpdate処理
			for _, listID := range listIDs {
				upd, err := db.Prepare("UPDATE result_v6list SET get_date = ?, org = ?, org_en = ?, post_code = ?, address = ?, address_en = ?, name_server = ?, ds_record = ?, notify_address = ?, admin_jpnic_id = ?, asn_id = ? WHERE id = ?")
				if err != nil {
					log.Println(err)
					return
				}
				_, err = upd.Exec(
					time.Now().UTC(),
					data.InfoIPv6[0].InfoDetail.Org,
					data.InfoIPv6[0].InfoDetail.OrgEn,
					data.InfoIPv6[0].InfoDetail.PostCode,
					data.InfoIPv6[0].InfoDetail.Address,
					data.InfoIPv6[0].InfoDetail.AddressEn,
					data.InfoIPv6[0].InfoDetail.NameServer,
					data.InfoIPv6[0].InfoDetail.DSRecord,
					data.InfoIPv6[0].InfoDetail.NotifyAddress,
					handles[data.InfoIPv6[0].InfoDetail.AdminJPNICHandle],
					cert.ASN,
					listID,
				)
				if err != nil {
					log.Println(err)
					return
				}
				upd.Close()
				// JPNIC技術連絡先をDBに登録
				//for _, techHandle := range data[0].InfoDetail.TechJPNICHandle {

				ins, err := db.Prepare("INSERT INTO result_v6list_tech_jpnic (v6list_id, jpnichandle_id) VALUES(?,?)")
				if err != nil {
					log.Println(err)
					return
				}

				ins.Exec(listID, handles[data.InfoIPv6[0].InfoDetail.TechJPNICHandle])
				ins.Close()
			}
		}
	} else {
		// IPv4
		rows, err := db.Query("SELECT id, ip_address, address, address_en, recep_number FROM result_v4list WHERE get_date > $1 AND asn_id = $2", timeDate, cert.ASN)
		if err != nil {
			log.Println("Error", "query result_v4_list", err)
			return
		}

		var list ResultV4List
		for rows.Next() {
			err = rows.Scan(&list.ID, &list.IPAddress, &list.Address, &list.AddressEn, &list.RecepNumber)
			if err != nil {
				log.Println("Error", "scan result_v4_list", err)
				return
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
				log.Println("Error", "prepare result_v4_list", err)
				return
			}
			_, err = upd.Exec(time.Now().UTC(), "　", "　", "　", cert.ASN, list.ID)
			if err != nil {
				log.Println("Error", "update result_v4_list", err)
				return
			}
			upd.Close()

			return
		}

		// 全体取得データがない場合
		if (list == ResultV4List{}) {
			// 1000件以上の場合も取得
			var infos []jpnic.InfoIPv4
			isOverList := true
			addressRange := ""

			for isOverList {
				filter := jpnic.SearchIPv4{
					IsDetail: false,
					Option1:  nil,
				}
				if addressRange != "" {
					filter.IPAddress = addressRange
				}
				data, err := jpnicConfig.SearchIPv4(filter)
				if err != nil {
					log.Println("Error", "SearchIPv4", err)
					continue
				}

				isOverList = data.IsOverList

				if isOverList {
					lastIPAddress, _, err := net.ParseCIDR(data.InfoIPv4[len(data.InfoIPv4)-1].IPAddress)
					if err != nil {
						log.Println("Error", "parseCIDR", err)
						return
					}

					addressRange = lastIPAddress.String() + "-255.255.255.255"
				}

				infos = append(infos, data.InfoIPv4...)
			}

			for _, tmp := range infos {
				var id string
				layout := "2006/01/02"
				assignDate, _ := time.Parse(layout, tmp.AssignDate)

				ins, err := db.Prepare("INSERT INTO result_v4list (get_date, ip_address, size, network_name, assign_date, return_date, org, org_en, resource_admin_short, recep_number, deli_number, type, division, post_code, address, address_en, name_server, ds_record, notify_address, admin_jpnic_id, asn_id) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?) RETURNING id;")
				if err != nil {
					log.Println("Error", "prepare result_v4list", err)
					return
				}
				defer ins.Close()

				ins.QueryRow(
					time.Now().UTC(),
					tmp.IPAddress,
					tmp.Size,
					tmp.NetworkName,
					assignDate,
					tmp.ReturnDate,
					tmp.InfoDetail.Org,
					tmp.InfoDetail.OrgEn,
					tmp.InfoDetail.Ryakusho,
					tmp.RecepNo,
					tmp.DeliNo,
					tmp.Type,
					tmp.Division,
					tmp.InfoDetail.PostCode,
					tmp.InfoDetail.Address,
					tmp.InfoDetail.AddressEn,
					tmp.InfoDetail.NameServer,
					tmp.InfoDetail.DSRecord,
					tmp.InfoDetail.NotifyAddress,
					tmp.InfoDetail.AdminJPNICHandle,
					cert.ASN,
				).Scan(&id)
			}
		} else {
			// イレギュラー処理
			if list.RecepNumber == "" {
				fmt.Println("データがありません")
				return
			}

			// 同じ受付番号がないか確認
			var listIDs []string
			//log.Println(timeDate, jpnicCert.ASN, list.RecepNumber)
			rows, err = db.Query("SELECT id FROM result_v4list WHERE get_date > $1 AND asn_id = $2 AND address = '' AND address_en = '' AND recep_number = $3", timeDate, cert.ASN, list.RecepNumber)
			if err != nil {
				log.Println("Error", "query result_v4list", err)
				return
			}

			for rows.Next() {
				var id string
				err = rows.Scan(&id)
				if err != nil {
					log.Println("Error", "scan result_v4list", err)
					return
				}
				listIDs = append(listIDs, id)
			}
			rows.Close()

			// JPNIC Handle探索
			rows, err = db.Query("SELECT id,jpnic_handle,get_date FROM result_jpnichandle WHERE get_date > $1 AND asn_id = $2 AND is_ipv6 = $3", timeDate, cert.ASN, false)
			if err != nil {
				log.Println("Error", "query result_jpnichandle", err)
				return
			}
			defer rows.Close()

			handles := make(map[string]int)
			var strHandles []string
			var handle JPNICHandle
			for rows.Next() {
				err = rows.Scan(&handle.ID, &handle.JPNICHandle, &handle.GetTime)
				if err != nil {
					log.Println("Error", "scan result_jpnichandle", err)
					return
				}

				handles[handle.JPNICHandle] = handle.ID
				strHandles = append(strHandles, handle.JPNICHandle)

				fmt.Printf("ID: %d(%s),Handle: %s\n", handle.ID, handle.GetTime, handle.JPNICHandle)
			}

			data, err := jpnicConfig.SearchIPv4(jpnic.SearchIPv4{
				IsDetail:  true,
				Option1:   strHandles,
				IPAddress: list.IPAddress,
				RecepNo:   list.RecepNumber,
			})
			if err != nil {
				log.Println("Error", "searchIPv4(detail)", err)
			}

			// jpnic_handle DBに追加処理
			if data != nil && len(data.JPNICHandleDetail) != 0 {
				for _, jpnicHandle := range data.JPNICHandleDetail {
					var jpnicHandleID string

					layout := "2006/01/02 15:04"
					updateDate, _ := time.Parse(layout, jpnicHandle.UpdateDate)
					ins, err := db.Prepare("INSERT INTO result_jpnichandle (is_ipv6, get_date, jpnic_handle, name, name_en, email, org, org_en, division, division_en, tel, fax, update_date, asn_id) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?) RETURNING id;")
					if err != nil {
						log.Println("Error", "prepare(INSERT) result_jpnichandle", err)
						return
					}
					ins.QueryRow(
						false,
						time.Now().UTC(),
						jpnicHandle.JPNICHandle,
						jpnicHandle.Name,
						jpnicHandle.NameEn,
						jpnicHandle.Email,
						jpnicHandle.Org,
						jpnicHandle.OrgEn,
						jpnicHandle.Division,
						jpnicHandle.DivisionEn,
						jpnicHandle.Tel,
						jpnicHandle.Fax,
						updateDate,
						cert.ASN,
					).Scan(&jpnicHandleID)

					ins.Close()

					handles[jpnicHandle.JPNICHandle], _ = strconv.Atoi(jpnicHandleID)
				}
			}
			// result_v4list DBにUpdate処理
			for _, listID := range listIDs {
				upd, err := db.Prepare("UPDATE result_v4list SET get_date = ?, org = ?, org_en = ?, post_code = ?, address = ?, address_en = ?, name_server = ?, ds_record = ?, notify_address = ?, admin_jpnic_id = ?, asn_id = ? WHERE id = ?")
				if err != nil {
					log.Println("Error", "prepare(UPDATE) result_jpnichandle", err)
					return
				}
				_, err = upd.Exec(
					time.Now().UTC(),
					data.InfoIPv4[0].InfoDetail.Org,
					data.InfoIPv4[0].InfoDetail.OrgEn,
					data.InfoIPv4[0].InfoDetail.PostCode,
					data.InfoIPv4[0].InfoDetail.Address,
					data.InfoIPv4[0].InfoDetail.AddressEn,
					data.InfoIPv4[0].InfoDetail.NameServer,
					data.InfoIPv4[0].InfoDetail.DSRecord,
					data.InfoIPv4[0].InfoDetail.NotifyAddress,
					handles[data.InfoIPv4[0].InfoDetail.AdminJPNICHandle],
					cert.ASN,
					listID,
				)
				if err != nil {
					log.Println("Error", "update result_jpnichandle", err)
					return
				}
				upd.Close()

				// JPNIC技術連絡先をDBに登録
				//for _, techHandle := range data[0].InfoDetail.TechJPNICHandle {

				ins, err := db.Prepare("INSERT INTO result_v4list_tech_jpnic (v4list_id, jpnichandle_id) VALUES(?,?)")
				if err != nil {
					log.Println("Error", "prepare(INSERT) result_v4list_tech_jpnic", err)
					return
				}

				ins.Exec(listID, handles[data.InfoIPv4[0].InfoDetail.TechJPNICHandle])
				ins.Close()
			}
		}
	}
}
