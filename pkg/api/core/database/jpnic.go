package database

func (b *Base) GetAllJPNIC() (*[]JPNIC, error) {
	var jpnicLists []JPNIC
	result := b.DB.Table("jpnic_admin_jpnic").Find(&jpnicLists)
	if result.Error != nil {
		return nil, result.Error
	}

	return &jpnicLists, nil
}
