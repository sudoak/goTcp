package main

import (
	"bufio"
	"context"
	"fmt"
	"handlers"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/mongodb/mongo-go-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// host     = "142.93.210.144"
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

type fields struct {
	ID int `bson:"_id"`
}

func main() {
	go pollDB(time.Minute)

	// ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb://sudoak:sudoak1@ds237337.mlab.com:37337/askak?retryWrites=false")
	client, _ = mongo.Connect(context.Background(), clientOptions)

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
	fmt.Println(tempContent)
	if len(tempContent) == 8 {
		var temp Device
		temp.DeviceID = tempContent[1]

		temp.E1 = tempContent[3]
		temp.E2 = tempContent[4]
		temp.E3 = tempContent[5]
		temp.E4 = tempContent[6]
		temp.E5 = tempContent[7]

		t := time.Now()
		tt, _ := handlers.TimeIn(t, "Asia/Kolkata")
		temp.TimeStamp = tt.Format("2006-01-02 15:04:05")
		temp.Date = tt.Format("2006-01-02")
		temp.Time = tt.Format("15:04:05")
		fmt.Printf("value=> %v", temp)
		collection := client.Database("askak").Collection("ceps")
		result, err := collection.InsertOne(context.Background(), temp)
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

func pollDB(d time.Duration) {
	ticker := time.NewTicker(d)
	go func() {
		for {
			select {
			case t := <-ticker.C:
				fmt.Println("Tick at", t)
				var dev []*Device
				collection := client.Database("askak").Collection("ceps")
				// ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
				projection := fields{
					ID: 0,
				}
				// queryTime := time.Now()
				// queryTimeT, _ := handlers.TimeIn(queryTime, "Asia/Kolkata")
				// dateToQuery := queryTimeT.Format("2006-01-02")

				result, err := collection.Find(context.TODO(), bson.D{{"device_id", "IOC1"}}, options.Find().SetProjection(projection))
				if err != nil {
					log.Fatal(err)
				}
				// fmt.Println(result)
				for result.Next(context.TODO()) {
					var d Device
					err = result.Decode(&d)
					if err != nil {
						log.Fatal("Error on Decoding the document", err)
					}
					dev = append(dev, &d)
				}
				if err := result.Err(); err != nil {
					log.Fatal(err)
				}

				// Close the cursor once finished
				result.Close(context.TODO())
				go makeExcelSheet(dev)

			}
		}
	}()
}

func makeExcelSheet(data []*Device) {
	f := excelize.NewFile()
	f.SetCellValue("Sheet1", "A1", "Devide ID")
	f.SetCellValue("Sheet1", "B1", "TImeStamp")
	f.SetCellValue("Sheet1", "C1", "Value(v) E1")
	f.SetCellValue("Sheet1", "D1", "Value(v) E2")
	f.SetCellValue("Sheet1", "E1", "Value(v) E3")
	f.SetCellValue("Sheet1", "F1", "Value(v) E4")
	f.SetCellValue("Sheet1", "G1", "Value(v) E5")
	for i, v := range data {
		f.SetCellValue("Sheet1", fmt.Sprintf("A%v", i+2), v.DeviceID)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%v", i+2), v.TimeStamp)
		f.SetCellValue("Sheet1", fmt.Sprintf("C%v", i+2), v.E1)
		f.SetCellValue("Sheet1", fmt.Sprintf("D%v", i+2), v.E2)
		f.SetCellValue("Sheet1", fmt.Sprintf("E%v", i+2), v.E3)
		f.SetCellValue("Sheet1", fmt.Sprintf("F%v", i+2), v.E4)
		f.SetCellValue("Sheet1", fmt.Sprintf("G%v", i+2), v.E5)
	}
	errRemoveContents := removeContents("./files/")
	if errRemoveContents != nil {
		fmt.Println(errRemoveContents)
	}
	err := f.SaveAs("./files/ayasta.xlsx")
	if err != nil {
		fmt.Println(err)
	}
}

func removeContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}
