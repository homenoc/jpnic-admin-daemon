package main

import (
	"database/sql"
	"fmt"
	"github.com/homenoc/jpnic-gui-daemon/pkg/core"
	"github.com/homenoc/jpnic-gui-daemon/pkg/core/jpnic"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"time"
)

func main() {
	config := core.Config{
		NextTime: 10,
		DB: struct {
			Type string `yaml:"type"`
			Path string `yaml:"path"`
		}{},
	}
	var getConfTimer uint = 5
	getConfTick := time.NewTicker(time.Duration(getConfTimer) * time.Second)
	getInfoTick := time.NewTicker(time.Duration(config.NextTime) * time.Second)

	log.Printf("start \n")
	for {
		select {
		case <-getConfTick.C:
			beforeNextTime := config.NextTime
			b, _ := ioutil.ReadFile("./config.yaml")
			err := yaml.Unmarshal(b, &config)
			if err != nil {
				log.Println(err)
			}
			log.Printf("config timer: %d\n", config.NextTime)
			if config.NextTime != beforeNextTime {
				getInfoTick = time.NewTicker(time.Duration(config.NextTime) * time.Second)
				log.Printf("New NextTimer: %d\n", config.NextTime)
			}
		case <-getInfoTick.C:
			log.Println("get Info Tick")
			var sqliteOption = "file:" + config.DB.Path + "?cache=shared&mode=rwc&_journal_mode=WAL"
			now := time.Now().UTC()
			timeDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

			var db *sql.DB
			db, err := sql.Open(config.DB.Type, sqliteOption)
			if err != nil {
				log.Fatal(err)
			}

			defer db.Close()

			certRows, err := db.Query("SELECT * FROM jpnic_gui_jpnic")
			if err != nil {
				log.Fatal(err)
			}
			defer certRows.Close()

			var jpnicCert core.JPNICCert

			for certRows.Next() {
				err = certRows.Scan(
					&jpnicCert.ID,
					&jpnicCert.Name,
					&jpnicCert.IsActive,
					&jpnicCert.IsIPv6,
					&jpnicCert.ASN,
					&jpnicCert.CA,
					&jpnicCert.P12Base64,
					&jpnicCert.P12Pass,
				)
				if err != nil {
					log.Fatal(err)
				}

				jpnicConfig := jpnic.Config{
					URL:       "https://iphostmaster.nic.ad.jp/jpnic/certmemberlogin.do",
					CA:        jpnicCert.CA,
					P12Base64: jpnicCert.P12Base64,
					P12Pass:   jpnicCert.P12Pass,
					//PfxFilePath: jpnicCert.P12Path,
					//PfxPass:     jpnicCert.P12Pass,
					//CAFilePath:  jpnicCert.CAPath,
				}

				rows, err := db.Query("SELECT id, ip_address, address, address_en, recep_number FROM result_v4list WHERE get_date > $1 AND asn_id = $2", timeDate, jpnicCert.ASN)
				if err != nil {
					log.Fatal(err)
				}

				var list core.ResultV4List
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
					_, err = upd.Exec(time.Now().UTC(), "　", "　", "　", jpnicCert.ASN, list.ID)
					if err != nil {
						log.Fatal(err)
					}
					upd.Close()

					return
				}

				// 全体取得データがない場合
				if (list == core.ResultV4List{}) {
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
							log.Println(err)
							continue
						}

						isOverList = data.IsOverList

						if isOverList {
							lastIPAddress, _, err := net.ParseCIDR(data.InfoIPv4[len(data.InfoIPv4)-1].IPAddress)
							if err != nil {
								log.Fatal(err)
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
							log.Fatal(err)
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
							jpnicCert.ASN,
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
					rows, err = db.Query("SELECT id FROM result_v4list WHERE get_date > $1 AND asn_id = $2 AND address = '' AND address_en = '' AND recep_number = $3", timeDate, jpnicCert.ASN, list.RecepNumber)
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
					rows, err = db.Query("SELECT id,jpnic_handle,get_date FROM result_jpnichandle WHERE get_date > $1 AND asn_id = $2 AND is_ipv6 = $3", timeDate, jpnicCert.ASN, false)
					if err != nil {
						log.Fatal(err)
					}
					defer rows.Close()

					handles := make(map[string]int)
					var strHandles []string
					var handle core.JPNICHandle
					for rows.Next() {
						err = rows.Scan(&handle.ID, &handle.JPNICHandle, &handle.GetTime)
						if err != nil {
							log.Fatal(err)
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
						log.Println(err)
						continue
					}

					// jpnic_handle DBに追加処理
					if len(data.JPNICHandleDetail) != 0 {
						for _, jpnicHandle := range data.JPNICHandleDetail {

							var jpnicHandleID string

							layout := "2006/01/02 15:04"
							updateDate, _ := time.Parse(layout, jpnicHandle.UpdateDate)
							ins, err := db.Prepare("INSERT INTO result_jpnichandle (is_ipv6, get_date, jpnic_handle, name, name_en, email, org, org_en, division, division_en, tel, fax, update_date, asn_id) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?) RETURNING id;")
							if err != nil {
								log.Fatal(err)
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
								jpnicCert.ASN,
							).Scan(&jpnicHandleID)

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
							jpnicCert.ASN,
							listID,
						)
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

						ins.Exec(listID, handles[data.InfoIPv4[0].InfoDetail.TechJPNICHandle])

						ins.Close()
						//}
					}
				}
			}
			certRows.Close()
		}
	}
	getInfoTick.Stop()
}
