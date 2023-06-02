package mysql

import (
	"fmt"
	"testing"

	"git.bianfeng.com/stars/wegame/wan/wanx/di/box"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gotest.tools/v3/assert"
)

var dbx *Mysql

func init() {
	opt := &Option{
		Host:      "192.168.191.185",
		Port:      "3306",
		Username:  "root",
		Password:  "6TcYi9GuDLiKGXYm",
		Databases: "fishing",
		Charset:   "utf8mb4",
	}
	dbx, _ = NewMysql2(nil, opt)
	dbx.AutoMigrate(&User{}, &UserDetail{}, &UserGroup{})
}

func NewMysql2(_ box.Context, opt *Option) (*Mysql, error) {
	m := &Mysql{Option: opt}
	/*
		dsn := fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
			opt.Username, opt.Password, opt.Host, opt.Port, opt.Databases, opt.Charset)

		db, err := gorm.Open(mysql.Open(dsn))*/
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		// SkipDefaultTransaction: true,
		// PrepareStmt:            true,
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	m.DB = db
	if opt.Debug {
		m.DB = m.DB.Debug()
	}
	return m, nil
}

func Test_User(t *testing.T) {
	u := NewUser(dbx)
	g := NewUserGroup(dbx)
	var groupType int8 = 1
	openId := "123123"
	groupId, _ := g.Add(&UserGroup{GroupType: groupType, OpenId: openId})
	assert.Assert(t, groupId > 0)

	group, _ := g.FindById(groupId)
	assert.Assert(t, group != nil)

	group, _ = g.FindOne(groupType, openId)
	assert.Assert(t, group != nil)

	user := &User{
		GroupId:  groupId,
		Status:   0,
		Nickname: "hhh",
	}
	detail := &UserDetail{
		BindPhone:     "1508866xxxx",
		Ip:            0,
		Imei:          "sdfs233sd",
		RegCpsId:      1,
		RegMethod:     1,
		RegVersion:    1100,
		RegPhoneModel: "mi10",
	}
	uid, _ := u.Add(user, detail)
	assert.Assert(t, uid > 0)

	fu, fd, _ := u.FindById(uid)
	assert.Assert(t, fu != nil, fd != nil)

	ul, _ := u.ListByGroupId(groupId)
	assert.Assert(t, len(ul) > 0)

	err := u.DeleteById(uid)
	assert.Assert(t, err == nil)
}

// ///////////////////////////////////////////////////////////////////////////////////
func Test_Repository(t *testing.T) {
	r := NewUser(dbx)
	user := &User{
		GroupId:  1,
		Status:   0,
		Nickname: "kkk",
	}
	detail := &UserDetail{
		BindPhone:     "1508866xxxx",
		Ip:            0,
		Imei:          "sdfs233sd",
		RegCpsId:      1,
		RegMethod:     1,
		RegVersion:    1100,
		RegPhoneModel: "mi10",
	}
	uid, _ := r.Add(user, detail)
	assert.Assert(t, uid > 0)

	_, _, err := r.FindById(uid)
	assert.Assert(t, err == nil)

	_, err = r.FindOne(&User{GroupId: 1})
	assert.Assert(t, err == nil)

	results, err := r.List(&User{Nickname: "kkk"})
	assert.Assert(t, err == nil)
	for _, result := range results {
		fmt.Println("result:", result)
	}

	results, cnt, err := r.Page(&User{Nickname: "kkk"}, 0, 5)
	assert.Assert(t, err == nil)

	cond := &User{Nickname: "kkk"}
	cond.ID = uid
	upt := &User{Status: 0}
	cnt, err = r.Update(cond, upt)
	assert.Assert(t, err == nil && cnt > 0)

	cond = &User{Nickname: "kkk"}
	m := map[string]interface{}{"status": 0}
	cnt, err = r.UpdateMap(cond, m)
	assert.Assert(t, err == nil && cnt > 0)

	cnt, err = r.UpdateById(uid, &User{Status: 1})
	assert.Assert(t, err == nil && cnt > 0)

	m = map[string]interface{}{"status": 0}
	cnt, err = r.UpdateMapById(uid, m)
	assert.Assert(t, err == nil && cnt > 0)

	err = r.DeleteById(uid)
	assert.Assert(t, err == nil)
}

func Test_Save(t *testing.T) {
	user := &User{
		GroupId:  1,
		Status:   0,
		Nickname: "kkk",
	}
	user.ID = 1
	r := NewUser(dbx)
	id, err := r.Save(user)
	fmt.Println(id, err)
}

// UserGroup 用户组, 一个组可以包含多个用户账号
type UserGroup struct {
	DBModel[*UserGroup]
	GroupType  int8   `gorm:"column:group_type;not null;default:0;index:idx_user_group_type" json:"group_type"`
	OpenId     string `gorm:"column:open_id;size:50;not null;default:'';index:idx_user_group_open_id" json:"open_id"`
	Nickname   string `gorm:"column:nickname;size:50;not null;default:''" json:"nickname"`
	UnionId    string `gorm:"column:union_id;size:50;not null;default:''" json:"union_id"`
	Sex        int8   `gorm:"column:sex;not null;default:1" json:"sex"`
	Country    string `gorm:"column:country;size:20;not null;default:''" json:"country"`
	Province   string `gorm:"column:province;size:20;not null;default:''" json:"province"`
	City       string `gorm:"column:city;size:20;not null;default:''" json:"city"`
	HeadImgUrl string `gorm:"column:head_img_url;size:500;not null;default:''" json:"head_img_url"`
	Email      string `gorm:"column:email;size:50;not null;default:''" json:"email"`
}

func NewUserGroup(dbx *Mysql) *UserGroup {
	m := &UserGroup{}
	m.DB = dbx
	return m
}

// FindOne 重载父类方法
func (g *UserGroup) FindOne(groupType int8, openId string) (*UserGroup, error) {
	cond := UserGroup{GroupType: groupType, OpenId: openId}
	return g.DBModel.FindOne(&cond)
}

func (g *UserGroup) TableName() string {
	return "user_group"
}

// User 用户账号表
type User struct {
	DBModel[*User]
	GroupId  int64  `gorm:"column:group_id;not null;default:0;index:idx_user_group_id" json:"group_id"`
	Status   int8   `gorm:"column:status;not null;default:0" json:"status"`
	Nickname string `gorm:"column:nickname;size:50;not null;default:''" json:"nickname"`
}

func NewUser(dbx *Mysql) *User {
	m := &User{}
	m.DB = dbx
	return m
}

// Add 重载父类方法
func (u *User) Add(user *User, detail *UserDetail) (int64, error) {
	err := u.DB.Transaction(func(tx *gorm.DB) (err error) {
		if err = tx.Create(user).Error; err != nil {
			return err
		}

		detail.UserId = user.ID
		if err = tx.Create(detail).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return 0, err
	} else {
		return user.ID, nil
	}
}

// DeleteById 重载父类方法
func (u *User) DeleteById(uid int64) error {
	err := u.DB.Transaction(func(tx *gorm.DB) (err error) {
		var user User
		if err = tx.First(&user, uid).Error; err != nil {
			return err
		}
		if err = tx.Delete(&User{}, uid).Error; err != nil {
			return err
		}
		if err = tx.Where(&UserDetail{UserId: uid}).Delete(&UserDetail{}).Error; err != nil {
			return err
		}
		return nil
	})

	return err
}

// FindById 重载父类方法
func (u *User) FindById(uid int64) (*User, *UserDetail, error) {
	var user User
	var detail UserDetail
	var err error
	if err = u.DB.First(&user, uid).Error; err != nil {
		return nil, nil, err
	}
	if err = u.DB.First(&detail, &UserDetail{UserId: uid}).Error; err != nil {
		return nil, nil, err
	}

	return &user, &detail, err
}

// ListByGroupId 自身方法
func (u *User) ListByGroupId(groupId int64) ([]*User, error) {
	return u.List(&User{GroupId: groupId})
}

func (u *User) TableName() string {
	return "user"
}

// UserDetail 用户明细
type UserDetail struct {
	DBModel[*UserDetail]
	UserId        int64  `gorm:"column:user_id;not null;default:0;index:idx_user_detail_user_id" json:"user_id"`
	BindPhone     string `gorm:"column:bind_phone;size:20;not null;default:''" json:"bind_phone"`
	Ip            uint32 `gorm:"column:ip;not null;default:0" json:"ip"`
	Imei          string `gorm:"column:imei;size:50;not null;default:''" json:"status"`
	RegCpsId      int32  `gorm:"column:reg_cps_id;not null;default:0" json:"reg_cps_id"`
	RegMethod     int8   `gorm:"column:reg_method;not null;default:0" json:"reg_method"`
	RegVersion    int32  `gorm:"column:reg_version;not null;default:0" json:"reg_version"`
	RegPhoneModel string `gorm:"column:reg_phone_model;size:50;not null;default:''" json:"reg_phone_model"`
}

func (d *UserDetail) TableName() string {
	return "user_detail"
}
