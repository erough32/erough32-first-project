package main

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ctx context.Context
var mongoCollectionBook,
	mongoCollectionUser,
	mongoCollectionBigTag,
	mongoCollectionTags,
	mongoCollectionMsg,
	mongoCollectionProportion,
	mongoCollectionDownloads *mongo.Collection

func init() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://127.0.0.1:27017"))
	if err != nil {
		fmt.Println(err)
	}

	mongoDatabase := mongoClient.Database("book")

	mongoCollectionBook = mongoDatabase.Collection("book_similar")
	mongoCollectionUser = mongoDatabase.Collection("user_similar")
	mongoCollectionBigTag = mongoDatabase.Collection("bigtag")
	mongoCollectionTags = mongoDatabase.Collection("tags")
	mongoCollectionMsg = mongoDatabase.Collection("msg")
	mongoCollectionProportion = mongoDatabase.Collection("proportion")
	mongoCollectionDownloads = mongoDatabase.Collection("downloads")
}
