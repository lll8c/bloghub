package dao

import (
	"context"
	"errors"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"time"
)

var (
	ErrUserDuplicateEmail = errors.New("邮箱冲突")
	ErrUserNotFound       = errors.New("用户没找到")
)

type UserDAO struct {
	db *gorm.DB
}

func NewUserDao(db *gorm.DB) *UserDAO {
	return &UserDAO{
		db: db,
	}
}

// User 直接对应数据库表结构，这个User是存储时候的User
type User struct {
	Id int64 `gorm:"primaryKey, autoIncrement"`
	//用户唯一
	Email    string `gorm:"unique"`
	Password string
	//创建时间，毫秒
	Ctime int64
	//更新时间，毫秒
	Utime int64
}

func (dao *UserDAO) Insert(ctx context.Context, u User) error {
	//毫秒数
	now := time.Now().UnixMilli()
	u.Utime = now
	u.Ctime = now
	err := dao.db.WithContext(ctx).Create(&u).Error
	//获取详细的数据库错误信息，如果是MySQLError
	if mysqlError, ok := err.(*mysql.MySQLError); ok {
		//唯一索引键冲突码
		const uniqueConflictsErrNo uint16 = 1062
		if mysqlError.Number == 1062 {
			//邮箱冲突
			return ErrUserDuplicateEmail
		}
	}
	return err
}

func (dao *UserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return u, ErrUserNotFound
	}
	return u, err
}

func (dao *UserDAO) FindById(ctx context.Context, id int64) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("id=?", id).Find(&u).Error
	return u, err
}
