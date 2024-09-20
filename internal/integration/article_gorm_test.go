package integration

import (
	"bytes"
	"encoding/json"
	"geektime/webook/internal/domain"
	"geektime/webook/internal/integration/startup"
	dao "geektime/webook/internal/repository/dao/article"
	jwt2 "geektime/webook/internal/web/jwt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
)

// 测试套件
type ArticleTestSuite struct {
	suite.Suite
	server *gin.Engine
	db     *gorm.DB
}

// 在所有测试执行之前，初始化一下内容
func (s *ArticleTestSuite) SetupSuite() {
	s.server = gin.Default()
	//设置用户id
	s.server.Use(func(ctx *gin.Context) {
		ctx.Set("claims", &jwt2.UserClaims{
			Uid: 123,
		})
	})
	s.db = startup.InitDB()
	articleHandler := startup.InitArticleHandler(dao.NewGROMArticleDAO(s.db))
	articleHandler.RegisterRoutes(s.server)
}

// 每个测试用例执行完都会执行这个方法
func (s *ArticleTestSuite) TearDownTest() {
	//清空所有数据，并将自增主键恢复为1
	s.db.Exec("truncate table articles")
	s.db.Exec("truncate table publish_articles")
}

func (s *ArticleTestSuite) TestEdit() {
	testCases := []struct {
		name string
		//集成测试准备数据
		before func(t *testing.T)
		//集成测试验证并数据
		after func(t *testing.T)
		//输入
		art      Article
		wantCode int
		//希望http响应带上帖子的id
		wantRes Result[int64]
	}{
		{
			name: "新建帖子-保存成功",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {
				//验证数据库
				var art dao.Article
				err := s.db.Where("id=?", 1).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				art.Ctime = 0
				art.Utime = 0
				assert.Equal(t, dao.Article{
					Id:       1,
					Title:    "我的标题",
					Content:  "我的内容",
					Status:   domain.ArticleStatusUnpublished,
					AuthorId: 123,
				}, art)
			},
			art: Article{
				Title:   "我的标题",
				Content: "我的内容",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Data: 1,
				Msg:  "Ok",
			},
		},
		{
			name: "修改已有的帖子，并保存",
			before: func(t *testing.T) {
				s.db.Create(dao.Article{
					Id:       2,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 123,
					Status:   domain.ArticleStatusUnpublished,
					Ctime:    123,
					Utime:    123,
				})
			},
			after: func(t *testing.T) {
				//验证数据库
				var art dao.Article
				err := s.db.Where("id=?", 2).First(&art).Error
				assert.NoError(t, err)
				//确保我更新了Utime
				assert.True(t, art.Utime > 0)
				art.Utime = 0
				assert.Equal(t, dao.Article{
					Id:       2,
					Title:    "我的标题1233",
					Content:  "我的内容1234",
					Status:   domain.ArticleStatusUnpublished,
					Ctime:    123,
					AuthorId: 123,
				}, art)
			},
			art: Article{
				Id:      2,
				Title:   "我的标题1233",
				Content: "我的内容1234",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Data: 2,
				Msg:  "Ok",
			},
		},
		{
			name: "修改别人的帖子",
			before: func(t *testing.T) {
				//测试用户是123，意味着修改别人的数据
				err := s.db.Create(dao.Article{
					Id:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 789,
					Status:   domain.ArticleStatusUnpublished,
					Ctime:    123,
					Utime:    123,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				//验证数据库
				var art dao.Article
				err := s.db.Where("id=?", 3).First(&art).Error
				assert.NoError(t, err)
				//确保我更新了Utime
				assert.True(t, art.Utime > 0)
				art.Utime = 0
				assert.Equal(t, dao.Article{
					Id:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorId: 789,
					Status:   domain.ArticleStatusUnpublished,
					Ctime:    123,
					Utime:    0,
				}, art)
			},
			art: Article{
				Id:      3,
				Title:   "我的标题1233",
				Content: "我的内容1234",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Code: 5,
				Msg:  "系统错误",
			},
		},
	}

	t := s.T()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)

			// 准备Req请求
			reqBody, err := json.Marshal(tc.art)
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost,
				"/articles/edit", bytes.NewReader([]byte(reqBody)))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// 执行请求并获取响应
			recorder := httptest.NewRecorder()
			s.server.ServeHTTP(recorder, req)
			// 断言结果
			assert.Equal(t, tc.wantCode, recorder.Code)
			if tc.wantCode != http.StatusOK {
				return
			}
			var res Result[int64]
			err = json.NewDecoder(recorder.Body).Decode(&res)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantRes, res)
		})
	}
}

func (s *ArticleTestSuite) TestPublish() {
	t := s.T()

	testCases := []struct {
		name string
		// 要提前准备数据
		before func(t *testing.T)
		// 验证并且删除数据
		after func(t *testing.T)
		req   Article

		// 预期响应
		wantCode   int
		wantResult Result[int64]
	}{
		{
			name: "新建帖子并发表",
			before: func(t *testing.T) {
				// 什么也不需要做
			},
			after: func(t *testing.T) {
				// 验证一下数据
				var art dao.Article
				err := s.db.Where("author_id = ?", 123).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.Id > 0)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				art.Utime = 0
				art.Ctime = 0
				art.Id = 0
				assert.Equal(t, dao.Article{
					Title:    "hello,你好",
					Content:  "随便试试",
					Status:   domain.ArticleStatusPublished,
					AuthorId: 123,
				}, art)
				var publishedArt dao.PublishArticle
				err = s.db.Where("author_id = ?", 123).First(&publishedArt).Error
				assert.NoError(t, err)
				assert.True(t, publishedArt.Id > 0)
				assert.True(t, publishedArt.Ctime > 0)
				assert.True(t, publishedArt.Utime > 0)
				publishedArt.Id = 0
				publishedArt.Ctime = 0
				publishedArt.Utime = 0
				assert.Equal(t, dao.PublishArticle(
					dao.Article{
						Title:    "hello,你好",
						Content:  "随便试试",
						Status:   domain.ArticleStatusPublished,
						AuthorId: 123,
					},
				), publishedArt)
			},
			req: Article{
				Title:   "hello,你好",
				Content: "随便试试",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Msg:  "Ok",
				Data: 1,
			},
		},
		{
			// 制作库有，但是线上库没有
			name: "更新帖子并发表",
			before: func(t *testing.T) {
				// 模拟已经存在的帖子
				s.db.Create(&dao.Article{
					Id:       2,
					Title:    "我的标题",
					Content:  "我的内容",
					Ctime:    456,
					Status:   domain.ArticleStatusPublished,
					Utime:    234,
					AuthorId: 123,
				})
			},
			after: func(t *testing.T) {
				// 验证一下数据
				var art dao.Article
				err := s.db.Where("id = ?", 2).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.Id > 0)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				art.Utime = 0
				art.Ctime = 0
				art.Id = 0
				assert.Equal(t, dao.Article{
					Title:    "新的标题",
					Content:  "新的内容",
					Status:   domain.ArticleStatusPublished,
					AuthorId: 123,
				}, art)
				var publishedArt dao.PublishArticle
				err = s.db.Where("id = ?", 2).First(&publishedArt).Error
				assert.NoError(t, err)
				assert.True(t, publishedArt.Id > 0)
				assert.True(t, publishedArt.Ctime > 0)
				assert.True(t, publishedArt.Utime > 0)
				publishedArt.Id = 0
				publishedArt.Ctime = 0
				publishedArt.Utime = 0
				assert.Equal(t, dao.PublishArticle(
					dao.Article{
						Title:    "新的标题",
						Content:  "新的内容",
						Status:   domain.ArticleStatusPublished,
						AuthorId: 123,
					}), publishedArt)
			},
			req: Article{
				Id:      2,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Msg: "Ok",
			},
		},
		{
			name: "更新帖子，并且重新发表",
			before: func(t *testing.T) {
				art := dao.Article{
					Id:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					Ctime:    456,
					Status:   1,
					Utime:    234,
					AuthorId: 123,
				}
				err := s.db.Create(&art).Error
				assert.NoError(t, err)
				part := dao.PublishArticle(art)
				err = s.db.Create(&part).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// 验证一下数据
				var art dao.Article
				err := s.db.Where("id = ?", 3).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.Id > 0)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				art.Utime = 0
				art.Ctime = 0
				art.Id = 0
				assert.Equal(t, dao.Article{
					Title:    "新的标题",
					Content:  "新的内容",
					Status:   domain.ArticleStatusPublished,
					AuthorId: 123,
				}, art)
				var publishedArt dao.PublishArticle
				err = s.db.Where("id = ?", 3).First(&publishedArt).Error
				assert.NoError(t, err)
				assert.True(t, publishedArt.Id > 0)
				assert.True(t, publishedArt.Ctime > 0)
				assert.True(t, publishedArt.Utime > 0)
				publishedArt.Id = 0
				publishedArt.Ctime = 0
				publishedArt.Utime = 0
				assert.Equal(t, dao.PublishArticle(
					dao.Article{
						Title:    "新的标题",
						Content:  "新的内容",
						Status:   domain.ArticleStatusPublished,
						AuthorId: 123,
					}), publishedArt)
			},
			req: Article{
				Id:      3,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Msg: "Ok",
			},
		},
		{
			name: "更新别人的帖子，并且发表失败",
			before: func(t *testing.T) {
				art := dao.Article{
					Id:      4,
					Title:   "我的标题",
					Content: "我的内容",
					Ctime:   456,
					Utime:   234,
					Status:  1,
					// 注意。这个 AuthorID 我们设置为另外一个人的ID
					AuthorId: 789,
				}
				s.db.Create(&art)
				part := dao.PublishArticle(dao.Article{
					Id:       4,
					Title:    "我的标题",
					Content:  "我的内容",
					Ctime:    456,
					Status:   2,
					Utime:    234,
					AuthorId: 789,
				})
				s.db.Create(&part)
			},
			after: func(t *testing.T) {
				// 验证一下数据
				var art dao.Article
				err := s.db.Where("id = ?", 4).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.Id > 0)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				art.Utime = 0
				art.Ctime = 0
				art.Id = 0
				assert.Equal(t, dao.Article{
					Title:    "我的标题",
					Content:  "我的内容",
					Status:   1,
					AuthorId: 789,
				}, art)
				var publishedArt dao.PublishArticle
				err = s.db.Where("id = ?", 4).First(&publishedArt).Error
				assert.NoError(t, err)
				assert.True(t, publishedArt.Id > 0)
				assert.True(t, publishedArt.Ctime > 0)
				assert.True(t, publishedArt.Utime > 0)
				publishedArt.Id = 0
				publishedArt.Ctime = 0
				publishedArt.Utime = 0
				assert.Equal(t, dao.PublishArticle(
					dao.Article{
						Title:    "我的标题",
						Content:  "我的内容",
						Status:   2,
						AuthorId: 789,
					},
				), publishedArt)
			},
			req: Article{
				Id:      4,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Code: 5,
				Msg:  "系统错误",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			data, err := json.Marshal(tc.req)
			// 不能有 error
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost,
				"/articles/publish", bytes.NewReader(data))
			assert.NoError(t, err)
			req.Header.Set("Content-Type",
				"application/json")
			recorder := httptest.NewRecorder()

			s.server.ServeHTTP(recorder, req)
			code := recorder.Code
			assert.Equal(t, tc.wantCode, code)
			if code != http.StatusOK {
				return
			}
			// 反序列化为结果
			// 利用泛型来限定结果必须是 int64
			var result Result[int64]
			err = json.Unmarshal(recorder.Body.Bytes(), &result)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantResult, result)
			tc.after(t)
		})
	}
}

func (s *ArticleTestSuite) TestABC() {
	s.T().Log("hello, 这是测试套件")
}

func TestArticle(t *testing.T) {
	suite.Run(t, &ArticleTestSuite{})
}

type Article struct {
	Id      int64  `json:"id"`
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
}

type Result[T any] struct {
	Code int    `json:"code,omitempty"`
	Msg  string `json:"msg,omitempty"`
	Data T      `json:"data,omitempty"`
}
