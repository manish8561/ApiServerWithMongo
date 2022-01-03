package main

import (
	"context"
	// "encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	// "go.mongodb.org/mongo-driver/mongo/readpref"
)

type Person struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Firstname string             `json:"firstname,omitempty", bson:"firstname,omitempty"`
	Lastname  string             `json:"lastname,omitempty", bson:"lastname,omitempty"`
	Role      string             `json:"role,omitempty", bson:"role,omitempty"`
}

var client *mongo.Client

func main() {
	fmt.Println("Starting the application...........")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, _ = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))

	handleRequests()
}

func handleRequests() {
	// Creates a gin router with default middleware:
	// logger and recovery (crash-free) middleware
	router := gin.Default()

	router.GET("/person", GetPeopleEndpoint)
	router.GET("/person/:id", GetPersonEndpoint)
	router.POST("/person", CreatePersonEndpoint)
	// router.PUT("/somePut", putting)
	// router.DELETE("/someDelete", deleting)
	// router.PATCH("/somePatch", patching)
	// router.HEAD("/someHead", head)
	// router.OPTIONS("/someOptions", options)

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	router.Run(":8080") // listen and serve on 0.0.0.0:8080
	log.Fatal(endless.ListenAndServe(":4242", router))
}
func GetPeopleEndpoint(c *gin.Context) {
	var people []Person
	collection := client.Database("goDatabase").Collection("people")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var person Person
		cursor.Decode(&person)
		people = append(people, person)
	}
	if err := cursor.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data":    people,
	})
}

func CreatePersonEndpoint(c *gin.Context) {

	var person Person
	if c.ShouldBind(&person) == nil { // bind data in post request for json or xml
		// log.Println(person.Firstname)
		// log.Println(person.Lastname)

		collection := client.Database("goDatabase").Collection("people")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		result, _ := collection.InsertOne(ctx, person)
		c.JSON(http.StatusOK, gin.H{
			"message": "success",
			"data":    result,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "No data posted",
		})
	}
}

func GetPersonEndpoint(c *gin.Context) {
	id, _ := primitive.ObjectIDFromHex(c.Param("id"))

	var person Person

	collection := client.Database("goDatabase").Collection("people")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := collection.FindOne(ctx, bson.D{{"_id", id}}).Decode(&person)
	if err != nil {
		fmt.Println("record does not exist")
		fmt.Println(err, id)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data":    person,
	})
}
