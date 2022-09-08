package database

import (
	"errors"
	"github.com/go-sql-driver/mysql"
)

func (b *Base) CreateResultV4ListTechJPNIC(handle []V4TechJPNICLists) ([]V4TechJPNICLists, error) {
	result := b.DB.Table("result_v4list_tech_jpnic").Create(&handle)
	if result.Error != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(result.Error, &mysqlErr) && mysqlErr.Number == 1062 {
			//log.Println("This is not error!! :duplicate")
			return handle, nil
		}
		return handle, result.Error
	}

	return handle, nil
}
