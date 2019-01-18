package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

type Todo struct {
	Id         string `json:"id" binding:"required"`
	Title      string `json:"title"`
	Category   string `json:"category"`
	Progress   int    `json:"progress,string"`
	Content    string `json:"content" binding:"required"`
	UserId     string `json:"userId"`
	CreatedAt  string `json:"createdAt"`
	ModifiedAt string `json:"modifiedAt"`
}

func main() {
	router := gin.Default()
	router.Use(static.Serve("/", static.LocalFile("./views", true)))
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:3000"},
		AllowMethods: []string{"GET", "POST", "DELETE"},
		AllowHeaders: []string{"Origin"},
	}))
	api := router.Group("/api")
	{
		api.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "pong",
			})
		})
	}
	api.GET("/todo", GetTodoList)
	api.GET("/todo/:id", GetTodo)
	api.POST("/todo", UpsertTodo)
	api.DELETE("/todo/:id", DeleteTodo)
	router.Run(":8000")
}

func GetTodoListFromFile(filename string) []Todo {
	var todoList = []Todo{}
	var db, _ = ioutil.ReadFile("./todo-db.json")
	json.Unmarshal(db, &todoList)
	return todoList
}

func GetTodoList(c *gin.Context) {
	todoList := GetTodoListFromFile("./todo-db.json")
	keyword := c.DefaultQuery("keyword", "")
	category := c.DefaultQuery("category", "")
	var filterList = []Todo{}
	for _, todo := range todoList {
		if IsSearchMatch(keyword, todo) && IsCategoryMatch(category, todo) {
			filterList = append(filterList, todo)
		}
	}
	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, filterList)
}

func IsSearchMatch(keyword string, todo Todo) bool {
	if keyword == "" {
		return true
	}
	if strings.Contains(todo.Title, keyword) {
		return true
	}
	if strings.Contains(todo.Content, keyword) {
		return true
	}
	return false
}

func IsCategoryMatch(category string, todo Todo) bool {
	if category == "" {
		return true
	}
	if todo.Category == category {
		return true
	}
	return false
}

func GetTodo(c *gin.Context) {
	todoList := GetTodoListFromFile("./todo-db.json")
	c.Header("Content-Type", "application/json")
	for _, todo := range todoList {
		if todo.Id == c.Param("id") {
			c.JSON(http.StatusOK, todo)
			return
		}
	}
	c.JSON(http.StatusNotFound, "todo not found")
}

func UpsertTodo(c *gin.Context) {
	todoList := GetTodoListFromFile("./todo-db.json")
	newTodo := Todo{}
	var now = time.Now().Format(time.RFC1123Z)
	var rawData, _ = c.GetRawData()
	json.Unmarshal(rawData, &newTodo)
	fmt.Println(rawData, newTodo)
	for index, todo := range todoList {
		if todo.Id == newTodo.Id {
			todo.Title = newTodo.Title
			todo.Category = newTodo.Category
			todo.Progress = newTodo.Progress
			todo.Content = newTodo.Content
			todo.ModifiedAt = now
			todoList[index] = todo
			var data, _ = json.Marshal(todoList)
			ioutil.WriteFile("./todo-db.json", data, 0644)
			c.JSON(http.StatusOK, todoList)
			return
		}
	}
	newId, _ := uuid.NewV4()
	newTodo.Id = newId.String()
	newTodo.CreatedAt = now
	newTodo.ModifiedAt = now
	todoList = append(todoList, newTodo)
	var data, _ = json.Marshal(todoList)
	ioutil.WriteFile("./todo-db.json", data, 0644)
	c.JSON(http.StatusOK, todoList)
}

func DeleteTodo(c *gin.Context) {
	todoList := GetTodoListFromFile("./todo-db.json")
	for index, todo := range todoList {
		if todo.Id == c.Param("id") {
			todoList = append(todoList[:index], todoList[index+1:]...)
			var data, _ = json.Marshal(todoList)
			ioutil.WriteFile("./todo-db.json", data, 0644)
			c.JSON(http.StatusOK, nil)
			return
		}
	}
	c.JSON(http.StatusNotFound, "todo not found")
}
