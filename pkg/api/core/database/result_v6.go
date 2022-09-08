package database

import (
	"errors"
	"gorm.io/gorm"
)

func (b *Base) CreateResultV6List(handle V6List) (V6List, error) {
	result := b.DB.Table("result_v6list").Create(&handle)
	if result.Error != nil {
		return handle, result.Error
	}

	return handle, nil
}

func (b *Base) GetV6JPNICDataReceivedCount(getStartDate, getEndDate string, AsnID uint) (int64, error) {
	var counter int64
	result := b.DB.Table("result_v6list").Select("id", "ip_address", "recep_number").
		Where("get_start_date >= ? AND get_start_date < ? AND asn_id = ? AND is_get = ? AND is_disabled = ?",
			getStartDate, getEndDate, AsnID, false, false).Count(&counter)
	if result.Error != nil {
		return counter, result.Error
	}

	return counter, result.Error
}

func (b *Base) GetV6JPNICDataNotReceivedCount(getStartDate, getEndDate string, AsnID uint) (int64, error) {
	var counter int64
	result := b.DB.Table("result_v6list").Select("id", "ip_address", "recep_number").
		Where("get_start_date >= ? AND get_start_date < ? AND asn_id = ? AND is_get = ? AND is_disabled = ?",
			getStartDate, getEndDate, AsnID, true, false).Count(&counter)
	if result.Error != nil {
		return counter, result.Error
	}

	return counter, result.Error
}

func (b *Base) GetRangeV6List(getStartDate, getEndDate string, AsnID uint) (V6List, error) {
	var jpnicHandleList V6List
	result := b.DB.Table("result_v6list").Select("id", "ip_address", "recep_number").
		Where("get_start_date >= ? AND get_start_date < ? AND is_get = ? AND is_disabled = ? AND asn_id = ?",
			getStartDate, getEndDate, 1, 0, AsnID).
		First(&jpnicHandleList)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return jpnicHandleList, nil
		}
		return jpnicHandleList, result.Error
	}

	return jpnicHandleList, nil
}

func (b *Base) GetRangeV6ListByRecepNumber(getStartDate, getEndDate, recepNumber string, AsnID uint) ([]V6List, error) {
	var jpnicHandleList []V6List
	result := b.DB.Table("result_v6list").Select("id", "ip_address", "recep_number").
		Where("get_start_date >= ? AND get_start_date < ? AND is_get = ? AND is_disabled = ? AND asn_id = ? AND recep_number = ?",
			getStartDate, getEndDate, true, false, AsnID, recepNumber).
		First(&jpnicHandleList)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return jpnicHandleList, nil
		}
		return jpnicHandleList, result.Error
	}

	return jpnicHandleList, nil
}

func (b *Base) UpdateV6List(list V6List) error {
	result := b.DB.Table("result_v6list").Model(&V6List{ID: list.ID}).
		Updates(map[string]interface{}{
			"get_date":       list.GetDate,
			"is_disabled":    list.IsDisabled,
			"is_get":         list.IsGet,
			"org":            list.Org,
			"org_en":         list.OrgEn,
			"post_code":      list.PostCode,
			"address":        list.Address,
			"address_en":     list.AddressEn,
			"name_server":    list.NameServer,
			"ds_record":      list.DsRecord,
			"notify_address": list.NotifyAddress,
			"admin_jpnic_id": list.AdminJpnicId,
			"asn_id":         list.AsnId,
		})
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (b *Base) UpdateV6ListIrregular(id uint, getStartDate string, AsnId uint) error {
	result := b.DB.Table("result_v6list").Model(&V6List{ID: id}).
		Updates(map[string]interface{}{
			"get_date":    getStartDate,
			"is_disabled": false,
			"is_get":      false,
			"asn_id":      AsnId,
		})
	if result.Error != nil {
		return result.Error
	}

	return nil
}
