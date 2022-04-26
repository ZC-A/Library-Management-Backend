package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"io"
	"lms/middleware"
	"lms/services"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testingConnecttoDB() {
	username := "root"
	password := "123456"
	address := "192.168.1.105:3306"
	tableName := "library_sys"

	services.Jikeapikey = "12444.6076a457ef7282751a39cc00e90ab6ab.fc8577f5db168733c16e3fd2641ab4ce"

	db, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v)/%v", username, password, address, tableName))
	if err != nil {
		panic("connect to DB failed: " + err.Error())
	}
	agent.DB = db
}

func dummyRouter() *gin.Engine {

	testingConnecttoDB()

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	//router.LoadHTMLFiles(fmt.Sprintf("%v/index.html", path))
	//router.Use(static.Serve("/static", static.LocalFile(staticPath, true)))
	log.Println("running dummyRouter for testing purpose")

	g1 := router.Group("/")
	g1.Use(middleware.UserAuth())
	{
		g1.POST("/getUserBooks", getUserBooksHandler)
		g1.POST("/borrowBook", borrowBookHandler)
		g1.POST("/returnBook", returnBookHandler)
	}

	g2 := router.Group("/")

	g2.Use(middleware.AdminAuth())
	{
		g2.POST("/updateBookStatus", updateBookStatusHandler)
		g2.POST("/deleteBook", deleteBookHandler)

	}

	router.POST("/addBook", addBookHandler)
	router.POST("/login", loginHandler)
	router.POST("/admin", adminLoginHandler)
	router.POST("/register", registerHandler)
	router.GET("/getCount", getCountHandler)
	router.GET("/getBooks", getBooksHandler)
	router.POST("/getBooks", getBooksHandler)
	router.POST("/getFine", getFineHandler)
	router.Static("/media", "F:\\Project\\Library-Management\\media")
	//router.StaticFile("/favicon.ico", fmt.Sprintf("%v/favicon.ico", staticPath))

	return router
}

//
//func stringMapToBytesReader(arguments map[string]interface{}) (*bytes.Reader, error) {
//	args, jmerr := json.Marshal(arguments)
//	return bytes.NewReader(args), jmerr
//}

func readFromJsonBody(body io.ReadCloser, target *map[string]interface{}) error {
	return json.NewDecoder(body).Decode(target)
}

func Test_addBookHandler(t *testing.T) {
	//type response struct {
	//	status string
	//	msg    string
	//}

	router := dummyRouter()
	w := httptest.NewRecorder()

	argsMap := map[string]interface{}{
		"isbn":     "9787115539168",
		"count":    "3",
		"location": "2A",
	}
	args, jmerr := json.Marshal(argsMap)
	if jmerr != nil {
		t.Fatalf("json marshal error: " + jmerr.Error())
	}
	req, err := http.NewRequest("POST", "/addBook", bytes.NewReader(args))
	if err != nil {
		t.Errorf("post error : " + err.Error())
	}
	req.ParseForm()
	req.PostForm.Add("isbn", "9787302423287")
	req.PostForm.Add("count", "3")
	req.PostForm.Add("location", "2A")
	router.ServeHTTP(w, req)
	result := make(map[string]interface{})
	err = readFromJsonBody(w.Result().Body, &result)
	if err != nil {
		t.Fatalf("json error " + err.Error())
	}

	log.Println(result["msg"])

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, float64(services.UpdateOK), result["status"].(float64))
}
