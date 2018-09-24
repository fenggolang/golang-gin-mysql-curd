package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type User struct {
	Username string `form:"username" json:"username" binding:"required"`
	Passwd   string `form:"passwd" json:"passwd" binding:"required"`
	Age      int    `form:"age" json:"age"`
}

func Middleware() gin.HandlerFunc {
	return func(context *gin.Context) {
		fmt.Println("before middleware")
		context.Set("request", "clinet_request")
		context.Next()
		fmt.Println("after middleware")
	}
}

func AuthMiddleware() gin.HandlerFunc {
	return func(context *gin.Context) {
		if cookie, err := context.Request.Cookie("session_id"); err != nil {
			value := cookie.Value
			fmt.Println("cookie.Value", value)
			if value == "123" {
				context.Next()
				return
			}
		}
		context.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		context.Abort()
		return
	}
}

func main() {
	router := gin.Default()

	router.GET("/user/:name", func(context *gin.Context) {
		name := context.Param("name")
		context.String(http.StatusOK, "Hello %s", name)
	})

	router.GET("/user/:name/*action", func(context *gin.Context) {
		name := context.Param("name")
		action := context.Param("action")
		message := name + " is " + action
		context.String(http.StatusOK, message)
	})

	router.POST("/form_post", func(context *gin.Context) {
		message := context.PostForm("message")
		nick := context.DefaultPostForm("nick", "anonymous")

		context.JSON(http.StatusOK, gin.H{
			"status": gin.H{
				"status_code": http.StatusOK,
				"status":      "ok",
			},
			"message": message,
			"nick":    nick,
		})
	})

	router.PUT("/post", func(context *gin.Context) {
		id := context.Query("id")
		page := context.DefaultQuery("page", "0")
		name := context.PostForm("name")
		message := context.PostForm("message")
		fmt.Printf("id: %s; page: %s; name: %s; message: %s \n", id, page, name, message)
		context.JSON(http.StatusOK, gin.H{
			"status_code": http.StatusOK,
		})
	})

	router.LoadHTMLGlob("templates/*")
	router.GET("/upload", func(context *gin.Context) {
		context.HTML(http.StatusOK, "upload.html", gin.H{
			"title": "index",
		})
	})

	// curl -XPOST http://127.0.0.1:8000/upload -F "upload=@/Users/ghost/Desttop/pic.jpg" -H "Content-Type: multipart/form-data"
	router.POST("/upload", func(context *gin.Context) {
		name := context.PostForm("name")
		fmt.Println(name)
		file, header, err := context.Request.FormFile("upload")
		if err != nil {
			context.String(http.StatusBadRequest, "Bad request")
			return
			// log.Fatal(err)
		}
		filename := header.Filename

		fmt.Println(file, err, filename)

		out, err := os.Create(filename)
		if err != nil {
			log.Fatal(err)
		}
		defer out.Close()
		_, err = io.Copy(out, file)
		if err != nil {
			log.Fatal(err)
		}
		context.String(http.StatusCreated, "upload successful")
	})

	// curl -XPOST http://127.0.0.1:8000/multi/upload -F "upload=@/Users/ghost/Desktop/pic.jpg" -F "upload=@/Users/ghost/Desktop/journey.png" -H "Content-Type: multipart/form-data"
	router.POST("/multi/upload", func(context *gin.Context) {
		err := context.Request.ParseMultipartForm(200000)
		if err != nil {
			log.Fatal(err)
		}
		formdata := context.Request.MultipartForm // ok, no problem so far, read the Form data

		// get the *fileheaders
		files := formdata.File["upload"] // grab the filenames

		for i, _ := range files { // loop through the files one by one
			file, err := files[i].Open()
			defer file.Close()
			if err != nil {
				log.Fatal(err)
			}
			out, err := os.Create(files[i].Filename)
			defer out.Close()

			if err != nil {
				log.Fatal(err)
			}

			_, err = io.Copy(out, file) // file not files[i]!
			if err != nil {
				log.Fatal(err)
			}

			context.String(http.StatusCreated, "upload successful")
		}
	})

	v1 := router.Group("/v1")
	v1.Use(Middleware())

	v1.GET("/login", func(context *gin.Context) {
		fmt.Println(context.MustGet("request").(string))
		context.String(http.StatusOK, "v1 login")
	})
	v2 := router.Group("/v2")
	v2.GET("/login", func(context *gin.Context) {
		context.String(http.StatusOK, "v2 login")
	})

	router.POST("/login", func(context *gin.Context) {
		var user User
		var err error
		contentType := context.Request.Header.Get("Content-Type")

		switch contentType {
		case "application/json":
			err = context.BindJSON(&user)
		case "application/x-www-form-urlencoded":
			err = context.BindWith(&user, binding.Form)
		}
		//err = context.Bind(&user)
		if err != nil {
			fmt.Println(err)
			log.Fatal(err)
		}
		context.JSON(http.StatusOK, gin.H{
			"username": user.Username,
			"passwd":   user.Passwd,
			"age":      user.Age,
		})
	})

	router.GET("/render", func(context *gin.Context) {
		contentType := context.DefaultQuery("content_type", "json")
		if contentType == "json" {
			context.JSON(http.StatusOK, gin.H{
				"user":   "wpc",
				"passwd": "123",
			})
		} else if contentType == "xml" {
			context.XML(http.StatusOK, gin.H{
				"user":   "wpc",
				"passwd": "123",
			})
		} else {
			context.YAML(http.StatusOK, gin.H{
				"user":   "wpc",
				"passwd": "123",
			})
		}
	})

	router.GET("/redict/google", func(context *gin.Context) {
		context.Redirect(http.StatusMovedPermanently, "https://www.google.com")
	})

	router.GET("/before", Middleware(), func(context *gin.Context) {
		request := context.MustGet("request").(string)
		fmt.Println("before handler")
		context.JSON(http.StatusOK, gin.H{
			"middle_request": request,
		})
	})

	router.Use(Middleware())
	{
		router.GET("/middleware", func(context *gin.Context) {
			request := context.MustGet("request").(string)
			req, _ := context.Get("request")
			fmt.Println(req)
			context.JSON(http.StatusOK, gin.H{
				"middle_request": request,
				"request":        req,
			})
		})
	}

	router.GET("/after", func(context *gin.Context) {
		request := context.MustGet("request").(string)
		context.JSON(http.StatusOK, gin.H{
			"middle_request": request,
		})
	})

	router.GET("/auth/signin", func(context *gin.Context) {
		cookie := &http.Cookie{
			Name:     "session_id",
			Value:    "123",
			Path:     "/",
			HttpOnly: true,
		}
		http.SetCookie(context.Writer, cookie)
		context.String(http.StatusOK, "Login successful")
	})

	router.GET("/home", AuthMiddleware(), func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{
			"data": "home",
		})
	})

	router.GET("/sync", func(context *gin.Context) {
		time.Sleep(5 * time.Second)
		log.Println("Done! in path" + context.Request.URL.Path)
	})

	router.GET("/async", func(context *gin.Context) {
		cCp := context.Copy()
		go func() {
			time.Sleep(5 * time.Second)
			log.Println("Done! in path" + cCp.Request.URL.Path)
		}()
	})

	router.Run(":8000")
}
