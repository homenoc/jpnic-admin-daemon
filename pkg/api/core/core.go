package core

import (
	"fmt"
	"github.com/homenoc/jpnic-admin-daemon/pkg/api/core/database"
	"github.com/homenoc/jpnic-admin-daemon/pkg/api/core/jpnic"
	"github.com/homenoc/jpnic-admin-daemon/pkg/api/core/tool/config"
	"github.com/homenoc/jpnic-admin-daemon/pkg/api/core/tool/etc"
	"log"
	"reflect"
	"time"
)

var dataCerts map[uint]*JPNICCert

func Start() {
	dataCerts = map[uint]*JPNICCert{}
	var getConfTimer uint = 60
	getConfTick := time.NewTicker(time.Duration(getConfTimer) * time.Second)

	go func() {
		var getInfoTimer uint = 10
		getInfoTick := time.NewTicker(time.Duration(getInfoTimer) * time.Second)

		log.Printf("start \n")
		for {
			select {
			case <-getInfoTick.C:
				log.Printf("start(getInfoTick) \n")
				now := time.Now()
				for id, dataValue := range dataCerts {
					if dataValue.RenewDate.Unix() <= now.Unix() {
						log.Println("getInfo", dataValue.RenewDate, "<", now)
						renewDate := now.Add(time.Minute * time.Duration(dataValue.Base.CollectionInterval))
						dataCerts[id].RenewDate = renewDate
						if dataValue.Base.Ada {
							go GetInitProcess(*dataValue)
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
			go config.GetCA()
			log.Println("get Info Tick")

			go func() {
				db, err := database.Connect()
				if err != nil {
					log.Println("get config", "Database connection error", err)
					return
				}

				jpnicLists, err := db.GetAllJPNIC()
				if err != nil {
					log.Println(err)
				}

				if jpnicLists == nil {
					log.Println("getting error jpnic list database...")
					return
				}

				for _, jpnicList := range *jpnicLists {
					now := time.Now()
					dataValue, isExists := dataCerts[jpnicList.ID]
					if isExists {
						if !reflect.DeepEqual(&jpnicList, &dataValue.Base) {
							log.Println("[replace] getting data: ", jpnicList.Name)
							dataCerts[jpnicList.ID] = &JPNICCert{
								Base:      jpnicList,
								RenewDate: now.Add(time.Minute * time.Duration(jpnicList.CollectionInterval)),
							}
						}
					} else {
						renewDate := now.Add(time.Minute * time.Duration(jpnicList.CollectionInterval))
						// Debug
						renewDate = now
						log.Println("[new] getting data: ", jpnicList.Name)
						dataCerts[jpnicList.ID] = &JPNICCert{
							Base:      jpnicList,
							RenewDate: renewDate,
						}
					}
				}

				for dataKey := range dataCerts {
					isJPNICCert := false
					for _, jpn := range *jpnicLists {
						if dataKey == jpn.ID {
							isJPNICCert = true
							break
						}
					}
					if !isJPNICCert {
						delete(dataCerts, dataKey)
					}
				}

			}()
		}
	}
}

type base struct {
	todayStartTime string
	todayEndTime   string
	jpnicConfig    jpnic.Config
	cert           JPNICCert
	db             *database.Base
}

func GetInitProcess(cert JPNICCert) {
	startTime := etc.GetTodayStartDateTime()
	endTime := etc.GetTodayEndDateTime(true)
	jpnicConfig := jpnic.Config{
		URL:       "https://iphostmaster.nic.ad.jp/jpnic/certmemberlogin.do",
		CA:        config.CA,
		P12Base64: cert.Base.P12Base64,
		P12Pass:   cert.Base.P12Pass,
	}

	db, err := database.Connect()
	if err != nil {
		log.Println("get config", "Database connection error", err)
	}

	b := base{
		todayStartTime: startTime,
		todayEndTime:   endTime,
		jpnicConfig:    jpnicConfig,
		cert:           cert,
		db:             db,
	}

	if !cert.Base.IsIPv6 {
		// IPv4
		err = b.GetJPNICProcess(&ipv4{
			base: b,
		})
		if err != nil {
			log.Println(err)
		}
	} else {
		//IPv6
		err = b.GetJPNICProcess(&ipv6{
			base: b,
		})
		if err != nil {
			log.Println(err)
		}
	}
}

func (b *base) GetJPNICProcess(p Process) error {
	log.Println("GetJPNICProcess")
	err := p.getBaseIPList()
	if err != nil {
		return err
	}
	// イレギュラー処理
	// この場合は、住所/住所(English)情報,受付番号に空白文字を突っ込む
	result, err := p.irregularProcess(b.cert.Base.ASN)
	if result {
		return nil
	}
	if err != nil {
		return err
	}

	// 本日分取得後に再度取得しないようにする
	isExists, err := checkJPNICDataExists(p)
	if err != nil {
		return err
	}
	if isExists {
		return fmt.Errorf("This is not error!!: (no need to get data.) ")
	}

	// 全体取得データがない場合
	if p.emptyCheck() {
		err = p.getBaseData()
		if err != nil {
			return err
		}
	} else {
		// イレギュラー処理
		if err = p.irregularCheck(); err != nil {
			return err
		}

		// 同じ受付番号がないか確認
		ids, err := p.getSameRecepNumID()
		if err != nil {
			log.Println(err)
			return err
		}

		// JPNIC Handle探索
		jpnicHandles, err := b.db.GetRangeJPNICHandle(b.todayStartTime, b.cert.Base.ASN, p.getIPv6())
		if err != nil {
			log.Println(err)
			return err
		}
		log.Println(jpnicHandles)

		handles := make(map[string]uint)
		var strHandles []string
		for _, jpnicHandle := range *jpnicHandles {
			handles[jpnicHandle.JPNICHandle] = jpnicHandle.ID
			strHandles = append(strHandles, jpnicHandle.JPNICHandle)
			//fmt.Printf("ID: %d(%s),Handle: %s\n", jpnicHandle.ID, jpnicHandle.GetStartDate, jpnicHandle.JPNICHandle)
		}

		err = p.getDetail(strHandles, handles, ids)
		if err != nil {
			return err
		}
	}
	return nil
}
