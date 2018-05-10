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
	"google.golang.org/api/iterator"
)

type Waktu struct {
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
		port = "8080"
	}
	if port == "" {
		log.Fatal("$PORT must be set")
	}

	//log.Println(port)

	router := gin.New()
	router.Use(cors.Default())
	router.Use(gin.Logger())
	router.LoadHTMLGlob("templates/pages/*")
	router.Static("/vendor", "./templates/vendor")
	router.Static("/dist", "./templates/dist")
	router.Static("/data", "./templates/data")
	router.Static("/js", "./templates/js")
	router.Static("/less", "./templates/less")
	//Routing
	router.GET("/", index)

	router.GET("/kondisi", func(c *gin.Context) {
		c.HTML(200, "kondisigudang.html", "")
	})

	router.GET("/input", func(c *gin.Context) {
		suhu := c.Query("suhu")
		klb := c.Query("klb")
		suhu2, err := strconv.Atoi(suhu)
		klb2, err := strconv.Atoi(klb)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
		}
		err = sendData(suhu2, klb2, ctx, client)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
		}

		statustemp(suhu2, klb2, ctx, client, c)

	})

	router.GET("/last", func(c *gin.Context) {
		data, err := lastData(ctx, client)

		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
		}

		c.JSON(http.StatusOK, data)
	})

	router.GET("/status", func(c *gin.Context) {
		data := getStatus(ctx, client)

		c.JSON(http.StatusOK, data)
	})

	router.GET("/masuk", func(c *gin.Context) {
		id := c.Query("id")
		ref, err := client.Collection("Obat").Doc(id).Get(ctx)
		if err != nil {
			log.Printf("Failed: %v", err)
		}
		if ref.Exists() {
			nama, err := ref.DataAt("nama")
			if err != nil {
				log.Printf("Failed: %v", err)
			}
			c.HTML(http.StatusOK, "inout.html", gin.H{
				"id":   id,
				"nama": nama,
			})
		} else {
			c.Redirect(301, "/baru?id="+id)
		}
	})

	router.GET("/baru", func(c *gin.Context) {
		id := c.Query("id")
		c.HTML(http.StatusOK, "baru.html", gin.H{
			"id": id,
		})
	})

	router.POST("/masuk", func(c *gin.Context) {
		id := c.PostForm("id")
		jml, err := strconv.Atoi(c.PostForm("jml"))
		last := getJml(id, ctx, client)
		_, err = client.Collection("Obat").Doc(id).Update(ctx, []firestore.Update{
			{Path: "jumlah", Value: last + jml},
		})

		if err != nil {
			log.Printf("Failed: %v", err)
		}
		c.String(200, "OK")
	})

	router.POST("/baru", func(c *gin.Context) {
		id := c.PostForm("id")
		nama := c.PostForm("nama")
		jenis := c.PostForm("jenis")
		pd := c.PostForm("pd")
		ket := c.PostForm("ket")
		jml, err := strconv.Atoi(c.PostForm("jml"))
		kdl := c.PostForm("kdl")

		data := map[string]interface{}{
			"nama":       nama,
			"jenis":      jenis,
			"produsen":   pd,
			"keterangan": ket,
			"jumlah":     jml,
			"kadaluarsa": kdl,
		}
		_, err = client.Collection("Obat").Doc(id).Set(ctx, data)
		if err != nil {
			log.Printf("Failed: %v", err)
		}
		c.String(200, "OK")
	})

	router.GET("/keluar", func(c *gin.Context) {
		id := c.Query("id")
		ref, err := client.Collection("Obat").Doc(id).Get(ctx)
		if err != nil {
			log.Printf("Failed: %v", err)
		}
		if ref.Exists() {
			nama, err := ref.DataAt("nama")
			if err != nil {
				log.Printf("Failed: %v", err)
			}
			c.HTML(http.StatusOK, "inout.html", gin.H{
				"id":   id,
				"nama": nama,
			})
		} else {
			c.String(http.StatusNotFound, "Error")
		}
	})

	router.POST("/keluar", func(c *gin.Context) {
		id := c.PostForm("id")
		jml, err := strconv.Atoi(c.PostForm("jml"))
		last := getJml(id, ctx, client)
		_, err = client.Collection("Obat").Doc(id).Update(ctx, []firestore.Update{
			{Path: "jumlah", Value: last - jml},
		})

		if err != nil {
			log.Printf("Failed: %v", err)
		}
		c.String(200, "OK")
	})

	router.GET("/index", func(c *gin.Context) {
		data := getAll(ctx, client)
		c.HTML(http.StatusOK, "index.html", gin.H{
			"values": data,
		})
	})

	router.GET("/delete", func(c *gin.Context) {
		id := c.Query("id")
		hapus(id, ctx, client)
		c.Redirect(301, "/stok")
	})

	router.GET("/setmaxmin", func(c *gin.Context) {
		s1, _ := strconv.Atoi(c.Query("s1"))
		s2, _ := strconv.Atoi(c.Query("s2"))
		k1, _ := strconv.Atoi(c.Query("k1"))
		k2, _ := strconv.Atoi(c.Query("k2"))

		setmaxmin(s1, s2, k1, k2, ctx, client)
	})

	router.GET("/getmax", func(c *gin.Context) {
		max,_:=checkData(ctx,client)
		c.JSON(200,max)
	})
	router.GET("/getmin", func(c *gin.Context) {
		_,min:=checkData(ctx,client)
		c.JSON(200,min)
	})

	//Run
	router.Run(":" + port)
}

func index(c *gin.Context) {
	c.String(http.StatusOK, "HELLO")
}

func hapus(id string, ctx context.Context, client *firestore.Client) {
	_, err := client.Collection("Obat").Doc(id).Delete(ctx)
	if err != nil {
		log.Printf("Failed: %v", err)
	}
}

func getJml(id string, ctx context.Context, client *firestore.Client) int {
	ref, err := client.Collection("Obat").Doc(id).Get(ctx)
	if err != nil {
		log.Printf("Failed: %v", err)
	}

	last, err := ref.DataAt("jumlah")
	if err != nil {
		log.Printf("Failed: %v", err)
	}

	return int(last.(int64))
}

func checkData(ctx context.Context, client *firestore.Client) (map[string]interface{}, map[string]interface{}) {
	max, err := client.Collection("data").Doc("max").Get(ctx)
	min, err := client.Collection("data").Doc("min").Get(ctx)
	if err != nil {
		log.Printf("Failed: %v", err)
	}

	datamax := max.Data()
	datamin := min.Data()

	return datamax, datamin
}

func sendData(s int, k int, ctx context.Context, client *firestore.Client) error {
	now := getTgl()
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		panic(err)
	}
	ref, _, err := client.Collection("data").Doc(now.thn).Collection(now.bln).Doc(now.tgl).Collection(now.sub).Add(ctx, map[string]interface{}{
		"suhu":       s,
		"kelembapan": k,
		"waktu":      time.Now().In(loc),
	})

	if err != nil {
		log.Printf("Failed: %v", err)
	}

	_, err = client.Collection("last").Doc("id").Set(ctx, map[string]interface{}{
		"last": ref,
	})

	if err != nil {
		log.Printf("Failed: %v", err)
	}

	return err
}

func getTgl() Waktu {
	now := time.Now()
	date := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	var wkt Waktu

	wkt.tgl = strconv.Itoa(now.Day())
	wkt.bln = now.Month().String()
	wkt.thn = strconv.Itoa(now.Year())
	wkt.sub = strconv.FormatFloat(now.Sub(date).Truncate(time.Second).Seconds(), 'f', -1, 64)

	return wkt
}

func getStatus(ctx context.Context, client *firestore.Client) map[string]interface{} {
	ref, err := client.Collection("data").Doc("kondisi").Get(ctx)
	if err != nil {
		log.Printf("Failed: %v", err)
	}

	data := ref.Data()
	if err != nil {
		log.Printf("Failed: %v", err)
	}

	return data
}

func lastData(ctx context.Context, client *firestore.Client) (map[string]interface{}, error) {
	ref, err := client.Collection("last").Doc("id").Get(ctx)
	if err != nil {
		log.Printf("Failed: %v", err)
	}

	last, err := ref.DataAt("last")
	if err != nil {
		log.Printf("Failed: %v", err)
	}

	lastref := last.(*firestore.DocumentRef)

	dataref, err := lastref.Get(ctx)
	if err != nil {
		log.Printf("Failed: %v", err)
	}

	data := dataref.Data()

	return data, err
}

func getAll(ctx context.Context, client *firestore.Client) map[string]interface{} {
	iter := client.Collection("Obat").Documents(ctx)
	j := make(map[string]interface{})
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("Failed to iterate: %v", err)
		}
		j[doc.Ref.ID] = doc.Data()
	}

	return j
}

func statustemp(suhu2 int, klb2 int, ctx context.Context, client *firestore.Client, c *gin.Context) {
	max, min := checkData(ctx, client)
	smax, kmax := max["suhu"].(int64), max["kelembapan"].(int64)
	smin, kmin := min["suhu"].(int64), min["kelembapan"].(int64)
	smax2, kmax2 := int(smax), int(kmax)
	smin2, kmin2 := int(smin), int(kmin)
	if suhu2 > smax2 {
		if klb2 > kmax2 {
			_, _ = client.Collection("data").Doc("kondisi").Update(ctx, []firestore.Update{
				{Path: "suhu", Value: "Lebih Tinggi"},
				{Path: "kelembapan", Value: "Lebih Tinggi"},
			})
			c.String(http.StatusOK, "NO1")
		} else if klb2 < kmin2 {
			_, _ = client.Collection("data").Doc("kondisi").Update(ctx, []firestore.Update{
				{Path: "suhu", Value: "Lebih Tinggi"},
				{Path: "kelembapan", Value: "Lebih Rendah"},
			})
			c.String(200, "NO2")
		} else {
			_, _ = client.Collection("data").Doc("kondisi").Update(ctx, []firestore.Update{
				{Path: "suhu", Value: "Lebih Tinggi"},
				{Path: "kelembapan", Value: "Aman"},
			})
			c.String(200, "NO3")
		}
	} else if suhu2 < smin2 {
		if klb2 > kmax2 {
			_, _ = client.Collection("data").Doc("kondisi").Update(ctx, []firestore.Update{
				{Path: "suhu", Value: "Lebih Rendah"},
				{Path: "kelembapan", Value: "Lebih Tinggi"},
			})
			c.String(http.StatusOK, "NO4")
		} else if klb2 < kmin2 {
			_, _ = client.Collection("data").Doc("kondisi").Update(ctx, []firestore.Update{
				{Path: "suhu", Value: "Lebih Rendah"},
				{Path: "kelembapan", Value: "Lebih Rendah"},
			})
			c.String(200, "NO5")
		} else {
			_, _ = client.Collection("data").Doc("kondisi").Update(ctx, []firestore.Update{
				{Path: "suhu", Value: "Lebih Rendah"},
				{Path: "kelembapan", Value: "Aman"},
			})
			c.String(200, "NO6")
		}
	} else {
		if klb2 > kmax2 {
			_, _ = client.Collection("data").Doc("kondisi").Update(ctx, []firestore.Update{
				{Path: "suhu", Value: "Aman"},
				{Path: "kelembapan", Value: "Lebih Tinggi"},
			})
			c.String(http.StatusOK, "NO7")
		} else if klb2 < kmin2 {
			_, _ = client.Collection("data").Doc("kondisi").Update(ctx, []firestore.Update{
				{Path: "suhu", Value: "Aman"},
				{Path: "kelembapan", Value: "Lebih Rendah"},
			})
			c.String(200, "NO8")
		} else {
			_, _ = client.Collection("data").Doc("kondisi").Update(ctx, []firestore.Update{
				{Path: "suhu", Value: "Aman"},
				{Path: "kelembapan", Value: "Aman"},
			})
			c.String(200, "OKK")
		}
	}
}

func setmaxmin(s1, s2, k1, k2 int, ctx context.Context, client *firestore.Client) {
	_, _ = client.Collection("data").Doc("max").Update(ctx, []firestore.Update{
		{Path: "suhu", Value: s2},
		{Path: "kelembapan", Value: k2},
	})
	_, _ = client.Collection("data").Doc("min").Update(ctx, []firestore.Update{
		{Path: "suhu", Value: s1},
		{Path: "kelembapan", Value: k1},
	})
}
