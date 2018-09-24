package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

type Person struct {
	Id        int    `json:"id" form:"id"`
	FirstName string `json:"first_name" form:"first_name"`
	LastName  string `json:"last_name" form:"last_name"`
}

func main() {
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/test?parseTime=true")
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalln(err)
	}

	router := gin.Default()

	router.GET("/", func(context *gin.Context) {
		context.String(http.StatusOK, "It works")
	})

	router.POST("/person", func(context *gin.Context) {
		firstName := context.Request.FormValue("first_name")
		lastName := context.Request.FormValue("last_name")

		rs, err := db.Exec("insert into person(first_name,last_name) values(?,?)", firstName, lastName)
		if err != nil {
			log.Fatalln(err)
		}
		id, err := rs.LastInsertId()
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Println("insert person Id {}", id)
		msg := fmt.Sprintf("insert successful %d", id)
		context.JSON(http.StatusOK, gin.H{
			"msg": msg,
		})
	})

	router.GET("/persons", func(context *gin.Context) {
		rows, err := db.Query("select id,first_name,last_name from person")
		defer rows.Close()

		if err != nil {
			log.Fatalln(err)
		}
		persons := make([]Person, 0)
		//var persosn []Person

		for rows.Next() {
			var person Person
			rows.Scan(&person.Id, &person.FirstName, &person.LastName)
			persons = append(persons, person)
		}
		if err = rows.Err(); err != nil {
			log.Fatalln(err)
		}

		context.JSON(http.StatusOK, gin.H{
			"persons": persons,
		})
	})

	router.GET("/person/:id", func(context *gin.Context) {
		id := context.Param("id")
		var person Person
		err := db.QueryRow("select sid,first_name,last_name from person where id=?", id).Scan(
			&person.Id, &person.FirstName, &person.LastName,
		)
		if err != nil {
			log.Fatalln(err)
			context.JSON(http.StatusOK, gin.H{
				"person": nil,
			})
			return
		}
		context.JSON(http.StatusOK, gin.H{
			"person": person,
		})
	})

	router.PUT("/person/:id", func(context *gin.Context) {
		cid := context.Param("id")
		id, err := strconv.Atoi(cid)
		if err != nil {
			log.Fatalln(err)
		}
		person := Person{Id: id}
		err = context.Bind(&person)
		if err != nil {
			log.Fatalln(err)
		}

		stmt, err := db.Prepare("update person set first_name=?,last_name=? where id=?")
		defer stmt.Close()
		if err != nil {
			log.Fatalln(err)
		}
		rs, err := stmt.Exec(person.FirstName, person.LastName, person.Id)
		if err != nil {
			log.Fatalln(err)
		}
		ra, err := rs.RowsAffected()
		if err != nil {
			log.Fatalln(err)
		}
		msg := fmt.Sprintf("Update person %d successful %d", person.Id, ra)
		context.JSON(http.StatusOK, gin.H{
			"msg": msg,
		})
	})

	router.DELETE("/person/:id", func(context *gin.Context) {
		cid := context.Param("id")
		id, err := strconv.Atoi(cid)
		if err != nil {
			log.Fatalln(err)
		}
		rs, err := db.Exec("DELETE from person where id=?", id)
		if err != nil {
			log.Fatalln(err)
		}
		ra, err := rs.RowsAffected()
		if err != nil {
			log.Fatalln(err)
		}
		msg := fmt.Sprintf("delete person %d successfule %d", id, ra)
		context.JSON(http.StatusOK, gin.H{
			"msg": msg,
		})
	})

	router.Run(":8000")
}
