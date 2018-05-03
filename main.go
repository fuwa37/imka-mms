package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"log"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
	"golang.org/x/net/context"
	"cloud.google.com/go/firestore"
	"time"
	"strconv"
)

type Waktu struct{
	tgl string
	bln string
	thn string
	sub string
}

func main() {
	// Use a service account
	ctx := context.Background()
	sa := option.WithCredentialsFile("kunci.json")
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Fatalln(err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	defer client.Close()

	port := os.Getenv("PORT")

	if port == "" {
		port="8080"
	}
	if port == "" {
		log.Fatal("$PORT must be set")
	}

	// log.Println(port)

	router := gin.New()
	router.Use(gin.Logger())

	router.GET("/", index)
	router.GET("/input", func(c *gin.Context) {
		suhu:=c.Query("suhu")
		klb:=c.Query("klb")
		suhu2, err:=strconv.Atoi(suhu)
		klb2, err:=strconv.Atoi(klb)
		if err==nil{

		}
		sendData(suhu2,klb2,ctx,client)
	})

	router.GET("/last", func(c *gin.Context){
		data:=lastData(ctx,client)

		c.JSON(http.StatusOK, data)
	})
	router.Run(":" + port)
}

func index(c *gin.Context) {
	c.String(http.StatusOK, "HELLO")
}

func sendData(s int, k int, ctx context.Context, client *firestore.Client) {
	now:=getTgl()
	ref, _, err := client.Collection("data").Doc(now.thn).Collection(now.bln).Doc(now.tgl).Collection(now.sub).Add(ctx, map[string]interface{}{
		"suhu":  s,
		"kelembapan": k,
		"waktu": time.Now(),
	})

	if err != nil {
		log.Fatalf("Failed: %v", err)
	}

	_, err = client.Collection("last").Doc("id").Set(ctx, map[string]interface{}{
		"last":ref,
	})

	if err != nil {
		log.Fatalf("Failed: %v", err)
	}
}

func getTgl() Waktu{
	now:=time.Now()
	date:=time.Date(now.Year(),now.Month(),now.Day(),0,0,0,0,time.Local)
	var wkt Waktu

	wkt.tgl=strconv.Itoa(now.Day())
	wkt.bln=now.Month().String()
	wkt.thn=strconv.Itoa(now.Year())
	wkt.sub=strconv.FormatFloat(now.Sub(date).Truncate(time.Second).Seconds(),'f',-1,64)

	return wkt
}

func lastData(ctx context.Context, client *firestore.Client) map[string]interface{}{
	ref, err := client.Collection("last").Doc("id").Get(ctx)
	if err != nil {
		log.Fatalf("Failed: %v", err)
	}

	last,err:=ref.DataAt("last")
	if err != nil {
		log.Fatalf("Failed: %v", err)
	}

	lastref:=last.(*firestore.DocumentRef)

	dataref, err:=lastref.Get(ctx)
	if err != nil {
		log.Fatalf("Failed: %v", err)
	}

	data:=dataref.Data()

	return data
}

