package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"log"
	"firebase.google.com/go"
	"google.golang.org/api/option"
	"golang.org/x/net/context"
	"cloud.google.com/go/firestore"
	"time"
	"strconv"
	"github.com/gin-contrib/cors"
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

	//log.Println(port)

	router := gin.New()
	router.Use(cors.Default())
	router.Use(gin.Logger())
	router.LoadHTMLGlob("templates/pages/*")
	router.Static("/vendor","./templates/vendor")
	router.Static("/dist","./templates/dist")
	router.Static("/data","./templates/data")
	router.Static("/js","./templates/js")
	router.Static("/less","./templates/less")
	//Routing
	router.GET("/", index)

	router.GET("/kondisi", func(c *gin.Context){
		c.HTML(200,"kondisigudang.html","")
	})

	router.GET("/input", func(c *gin.Context) {
		suhu:=c.Query("suhu")
		klb:=c.Query("klb")
		suhu2, err:=strconv.Atoi(suhu)
		klb2, err:=strconv.Atoi(klb)
		if err!=nil{
			c.String(http.StatusBadRequest, err.Error())
		}
		err=sendData(suhu2,klb2,ctx,client)
		if err!=nil{
			c.String(http.StatusBadRequest, err.Error())
		}
		c.String(http.StatusOK, suhu+klb)
	})

	router.GET("/last", func(c *gin.Context){
		data, err:=lastData(ctx,client)

		if err!=nil{
			c.String(http.StatusBadRequest, err.Error())
		}

		c.JSON(http.StatusOK, data)
	})

	router.GET("/masuk", func(c *gin.Context){
		id:=c.Query("id")
		ref, err := client.Collection("Obat").Doc(id).Get(ctx)
		if err != nil {
			//log.Printf("Failed: %v", err)
		}
		if ref.Exists() {
			nama, err:=ref.DataAt("nama")
			if err != nil {
				log.Printf("Failed: %v", err)
			}
			c.HTML(http.StatusOK, "masuk.html", gin.H{
				"id":id,
				"nama":nama,
			})
		} else {
			c.Redirect(301,"/baru?id="+id)
		}
	})

	router.GET("/baru", func(c *gin.Context) {
		id := c.Query("id")
		c.HTML(http.StatusOK, "baru.html", gin.H{
			"id":id,
		})
	})

	router.POST("/masuk", func(c *gin.Context){
		id := c.PostForm("id")
		jml:=c.PostForm("jml")
		_, err = client.Collection("Obat").Doc(id).Update(ctx, []firestore.Update{
			{Path: "jumlah", Value: jml},
		})

		if err != nil {
			log.Printf("Failed: %v", err)
		}
	})

	router.POST("/baru", func(c *gin.Context) {
		id := c.PostForm("id")
		nama:=c.PostForm("nama")
		jenis:=c.PostForm("jenis")
		pd:=c.PostForm("pd")
		ket:=c.PostForm("ket")
		jml:=c.PostForm("jml")
		kdl:=c.PostForm("kdl")

		data:=map[string]interface{}{
			"nama":nama,
			"jenis":jenis,
			"produsen":pd,
			"keterangan":ket,
			"jumlah":jml,
			"kadaluarsa":kdl,
		}
		_, err:=client.Collection("Obat").Doc(id).Set(ctx,data)
		if err != nil {
			log.Printf("Failed: %v", err)
		}
	})

	router.GET("/keluar", func(c *gin.Context){

	})

	router.POST("/keluar", func(c *gin.Context){

	})

	//Run
	router.Run(":" + port)
}

func index(c *gin.Context) {
	c.String(http.StatusOK, "HELLO")
}



func sendData(s int, k int, ctx context.Context, client *firestore.Client) error{
	now:=getTgl()
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		panic(err)
	}
	ref, _, err := client.Collection("data").Doc(now.thn).Collection(now.bln).Doc(now.tgl).Collection(now.sub).Add(ctx, map[string]interface{}{
		"suhu":  s,
		"kelembapan": k,
		"waktu": time.Now().In(loc),
	})

	if err != nil {
		log.Printf("Failed: %v", err)
	}

	_, err = client.Collection("last").Doc("id").Set(ctx, map[string]interface{}{
		"last":ref,
	})

	if err != nil {
		log.Printf("Failed: %v", err)
	}

	return err
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

func lastData(ctx context.Context, client *firestore.Client) (map[string]interface{},error) {
	ref, err := client.Collection("last").Doc("id").Get(ctx)
	if err != nil {
		log.Printf("Failed: %v", err)
	}

	last,err:=ref.DataAt("last")
	if err != nil {
		log.Printf("Failed: %v", err)
	}

	lastref:=last.(*firestore.DocumentRef)

	dataref, err:=lastref.Get(ctx)
	if err != nil {
		log.Printf("Failed: %v", err)
	}

	data:=dataref.Data()

	return data, err
}



