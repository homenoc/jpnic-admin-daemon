package database

func (b *Base) CreateResultV4ListTechJPNIC(handle V4TechJPNICLists) (V4TechJPNICLists, error) {
	result := b.DB.Table("result_v4list_tech_jpnic").Create(&handle)
	if result.Error != nil {
		return handle, result.Error
	}

	return handle, nil
}
