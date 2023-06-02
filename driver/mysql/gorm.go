package mysql

import (
	"errors"
	"git.bianfeng.com/stars/wegame/wan/wanx/driver/redis"
	"reflect"
	"strings"
	"time"
)

type BaseModel struct {
	ID         int64     `json:"id" gorm:"primarykey"`
	CreateTime time.Time `json:"create_time" gorm:"autoCreateTime"`
	UpdateTime time.Time `json:"update_time" gorm:"autoUpdateTime"`
}

type DBModel[T any] struct {
	BaseModel
	DB *Mysql `json:"-" gorm:"-"`
}

type RDSModel[T any] struct {
	BaseModel
	RDS *redis.Redis `json:"-" gorm:"-"`
}

type DBAndRDSModel[T any] struct {
	DBModel[T]
	RDS *redis.Redis `json:"-" gorm:"-"`
}

func (r *DBModel[T]) Add(t T) (int64, error) {
	if err := r.DB.Create(t).Error; err != nil {
		return 0, err
	} else {
		val := reflect.ValueOf(t)
		id := reflect.Indirect(val).FieldByName("ID").Int()
		return id, nil
	}
}

func (r *DBModel[T]) Save(t T) (int64, error) {
	val := reflect.ValueOf(t)
	id := reflect.Indirect(val).FieldByName("ID").Int()
	if id > 0 {
		return r.UpdateById(id, t)
	} else {
		return r.Add(t)
	}
}

func (r *DBModel[T]) DeleteById(id int64) error {
	var t T
	return r.DB.Delete(&t, id).Error
}

func (r *DBModel[T]) Update(cond T, upt T) (int64, error) {
	tx := r.DB.Where(cond).Updates(upt)
	if err := tx.Error; err != nil {
		return 0, err
	}
	return tx.RowsAffected, nil
}

func (r *DBModel[T]) UpdateMap(cond T, upt map[string]interface{}) (int64, error) {
	tx := r.DB.Model(cond).Where(cond).Updates(upt)
	if err := tx.Error; err != nil {
		return 0, err
	}
	return tx.RowsAffected, nil
}

func (r *DBModel[T]) UpdateById(id int64, upt T) (int64, error) {
	if id <= 0 {
		return 0, errors.New("id can't be empty")
	}

	tx := r.DB.Where("id = ?", id).Updates(upt)
	if err := tx.Error; err != nil {
		return 0, err
	}
	return tx.RowsAffected, nil
}

func (r *DBModel[T]) UpdateMapById(id int64, upt map[string]interface{}) (int64, error) {
	var t T
	tx := r.DB.Model(&t).Where("id = ?", id).Updates(upt)
	if err := tx.Error; err != nil {
		return 0, err
	}
	return tx.RowsAffected, nil
}

func (r *DBModel[T]) FindById(id int64) (T, error) {
	var t T
	err := r.DB.First(&t, id).Error
	return t, err
}

func (r *DBModel[T]) FindOne(cond T) (T, error) {
	var t T
	err := r.DB.First(&t, cond).Error
	return t, err
}

func (r *DBModel[T]) List(cond T, orders ...string) ([]T, error) {
	var results []T
	ord := strings.Join(orders, ",")
	err := r.DB.Order(ord).Find(&results, cond).Error
	return results, err
}

func (r *DBModel[T]) Count(cond T) (int64, error) {
	var t T
	var total int64
	err := r.DB.Model(&t).Where(cond).Count(&total).Error
	return total, err
}

func (r *DBModel[T]) Exists(cond T) (bool, error) {
	cnt, err := r.Count(cond)
	if err != nil {
		return false, err
	} else {
		return cnt > 0, nil
	}
}

func (r *DBModel[T]) Page(cond T, offset, size int, orders ...string) ([]T, int64, error) {
	var results []T
	ord := strings.Join(orders, ",")
	total, err := r.Count(cond)
	if err != nil {
		return results, total, err
	} else if total == 0 {
		return results, 0, nil
	}

	if err := r.DB.Where(cond).Limit(size).Offset(offset).Order(ord).Find(&results).Error; err != nil {
		return results, 0, err
	} else {
		return results, total, nil
	}
}
