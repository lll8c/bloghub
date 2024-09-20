package dao

import (
	"context"
	dao "geektime/webook/internal/repository/dao/article"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"gorm.io/gorm"
	"time"
)

func InitTables(db *gorm.DB) error {
	return db.AutoMigrate(&User{}, &dao.Article{}, &dao.PublishArticle{})
}

func InitCollection(mdb *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	col := mdb.Collection("articles")
	//定义articles集合索引
	_, err := col.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{bson.E{"id", 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{bson.E{"author_id", 1}},
		},
	})
	if err != nil {
		return err
	}
	//定义published_articles索引
	liveCol := mdb.Collection("published_articles")
	_, err = liveCol.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{bson.E{"id", 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{bson.E{"author_id", 1}},
		},
	})
	return err
}
