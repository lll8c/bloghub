package service

import (
	"context"
	"errors"
	"geektime/webook/internal/domain"
	"geektime/webook/internal/repository"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserDuplicate         = repository.ErrUserDuplicate
	ErrInvalidUserOrPassword = errors.New("用户不存在或密码不对")
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (svc *UserService) SignUp(ctx context.Context, u domain.User) error {
	//加密
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return svc.repo.Create(ctx, u)
}

func (svc *UserService) Login(ctx context.Context, email string, password string) (domain.User, error) {
	//根据邮箱查找用户
	u, err := svc.repo.FindByEmail(ctx, email)
	//用户不存在
	if err != nil {
		if err == repository.ErrUserNotFound {
			return domain.User{}, ErrInvalidUserOrPassword
		}
		return u, err
	}
	//比较密码
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	//密码不对
	if err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return u, err
}

func (svc *UserService) Profile(ctx context.Context, id int64) (domain.User, error) {
	return svc.repo.FindById(ctx, id)
}

func (svc *UserService) FindOrCreate(ctx *gin.Context, phone string) (domain.User, error) {
	//这个叫快路径，数据库中找到了就不用去创建了
	//快路径可以不执行，直接执行慢路径，直接去创建，冲突则创建失败
	u, err := svc.repo.FindByPhone(ctx, phone)
	//找到了 或者 报错
	if err != repository.ErrUserNotFound {
		return u, err
	}
	/*	//在系统资源不足，触发降级之后，不执行慢路径了
		if ctx.Value("降级") == "true" {
			return domain.User{}, errors.New("系统降级了")
		}*/
	//这个叫慢路径
	//没有找到，就用手机号码注册一个新用户
	err = svc.repo.Create(ctx, domain.User{
		Phone: phone,
	})
	//手机号冲突就说明用户存在
	if err != nil && err != ErrUserDuplicate {
		return domain.User{}, err
	}
	//查找新创建的或已存在的用户，获取id
	return svc.repo.FindByPhone(ctx, phone)
}
