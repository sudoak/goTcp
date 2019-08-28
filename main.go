package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/mongodb/mongo-go-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	host     = "localhost"
	port     = "3333"
	connType = "tcp"
)

var client *mongo.Client

type Device struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	DeviceID  string             `json:"device_id,omitempty" bson:"device_id,omitempty"`
	E1        string             `json:"e1,omitempty" bson:"e1,omitempty"`
	E2        string             `json:"e2,omitempty" bson:"e2,omitempty"`
	E3        string             `json:"e3,omitempty" bson:"e3,omitempty"`
	E4        string             `json:"e4,omitempty" bson:"e4,omitempty"`
	E5        string             `json:"e5,omitempty" bson:"e5,omitempty"`
	Date      string             `json:"date" bson:"date"`
	Time      string             `json:"time" bson:"time"`
	TimeStamp string             `json:"timestamp" bson:"timestamp"`
}

func main() {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb://sudoak:sudoak1@ds237337.mlab.com:37337/askak?retryWrites=false")
	client, _ = mongo.Connect(ctx, clientOptions)

	fmt.Println("Connected to MongoDB!")

	// Listen for incoming connections.
	l, err := net.Listen(connType, host+":"+port)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()
	fmt.Println("Listening on " + host + ":" + port)
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine.
		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	message, _ := bufio.NewReader(conn).ReadString('\n')

	message = strings.Trim(string(message), "$#")

	tempContent := strings.Split(string(message), ",")

	if len(tempContent) == 8 {
		var temp Device
		temp.DeviceID = tempContent[1]

		temp.E1 = tempContent[3]
		temp.E2 = tempContent[4]
		temp.E3 = tempContent[5]
		temp.E4 = tempContent[6]
		temp.E5 = tempContent[7]

		t := time.Now()
		temp.TimeStamp = t.Format("2006-01-02 15:04:05")
		temp.Date = t.Format("2006-01-02")
		temp.Time = t.Format("15:04:05")
		fmt.Println(temp)
		collection := client.Database("askak").Collection("ceps")
		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		result, err := collection.InsertOne(ctx, temp)
		fmt.Println(result)
		if err != nil {
			log.Printf("err = %v : time = %v", err.Error(), t.Format("2006-01-02 15:04:05"))
		}
		return
	}

	// Write to client
	conn.Write([]byte("Message received."))
	// Close the connection when you're done with it.
	conn.Close()
}
