package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type post struct { //구조체 선언
	ID       string `json:"Id"`
	Contents string `json:"contents"`
	Pwd      string `json:"pwd"`
}
type updatepost struct {
	Contents string `json:"contents"`
	Pwd      string `json:"pwd"`
}

func main() {
	r := gin.Default()
	r.GET("", ReadPosts)
	r.POST("", CreatePosts)
	r.PUT("/post/:id", UpdatePosts)
	r.DELETE("/post/:id", DeletePosts)
	r.Run(":8080")
}

func Getconnect() (client *mongo.Client, ctx context.Context, cancel context.CancelFunc) { //MongoDB를 연결함
	clientoptions := options.Client().ApplyURI("mongodb://localhost:27017").SetAuth(options.Credential{ //setAuth로 권한 설정
		Username: "root",
		Password: "root",
	})
	ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second) //timeout시간이 지나면 context종료
	client, err := mongo.Connect(ctx, clientoptions)
	if err != nil {
		log.Fatal(err)
	}
	return client, ctx, cancel
}

var data []post

func ReadPosts(c *gin.Context) {
	client, ctx, cancel := Getconnect()
	defer cancel() // 함수 종료 뒤 DB연결을 끊어지도록 설정, 가장 마지막에 실행한다.
	defer client.Disconnect(ctx)
	collection := client.Database("test").Collection("inventory") //리턴받은 Client를 통해 MongoDB에서 사용할 Database와 Collections을 지정한다.

	cursor, err := collection.Find(ctx, bson.D{{}}) //mongodb의 collection의 내용을 모두 가져옴 find({}) -> 모두가져온다는 의미
	if err != nil {
		log.Fatal(err)
	}
	data = []post{}
	for cursor.Next(context.TODO()) { //가져온 몽고디비의 내용을 하나씩 읽음
		elem := post{}
		err := cursor.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}
		data = append(data, elem)
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"data":   data,
	})
}
func CreatePosts(c *gin.Context) {
	client, ctx, cancel := Getconnect()
	defer cancel() // 함수 종료 뒤 DB연결을 끊어지도록 설정
	defer client.Disconnect(ctx)
	collection := client.Database("test").Collection("inventory") //DB랑 연결하고, collection 가져옴

	p := post{}
	if err := c.ShouldBindJSON(&p); err != nil {
		log.Fatal(err)
		c.JSON(http.StatusBadRequest, gin.H{"msg": err})
		return
	}
	fmt.Println(p)
	collection.InsertOne(ctx, p)

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"data":    p,
		"message": "Insert Success",
	})
}

// update a document
func UpdatePosts(c *gin.Context) {
	client, ctx, cancel := Getconnect()
	defer cancel() // 함수 종료 뒤 DB연결을 끊어지도록 설정
	defer client.Disconnect(ctx)
	collection := client.Database("test").Collection("inventory") //DB랑 연결하고, collection 가져옴

	updatedata := updatepost{}
	if err := c.ShouldBind(&updatedata); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": err.Error(),
		})
		return
	}
	id := c.Param("id")
	collection.FindOneAndUpdate(ctx, bson.M{"id": id}, bson.M{"$set": updatedata})
	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"data":    updatedata,
		"message": "Update Success",
	})
}

//delete a document
func DeletePosts(c *gin.Context) {
	client, ctx, cancel := Getconnect()
	defer cancel() // 함수 종료 뒤 DB연결을 끊어지도록 설정
	defer client.Disconnect(ctx)
	collection := client.Database("test").Collection("inventory") //DB랑 연결하고, collection 가져옴

	id := c.Param("id")
	collection.FindOneAndDelete(ctx, bson.M{"id": id})

	c.JSON(http.StatusOK, gin.H{
		"Contents": id,
		"status":   http.StatusOK,
		"message":  "Delete Success.",
	})
}
