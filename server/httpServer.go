package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-ini/ini"
	_ "github.com/go-sql-driver/mysql"
	"lms/middleware"
	. "lms/services"
	"lms/util"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"net/smtp"
)

var agent DBAgent

func loginHandler(context *gin.Context) {
	username := context.PostForm("username")
	password := context.PostForm("password")
	loginResult, userID := agent.AuthenticateUser(username, password)
	if loginResult.Status == UserLoginOK {
		token := util.GenToken(userID, util.UserKey)
		context.JSON(http.StatusOK, gin.H{"status": loginResult.Status, "msg": loginResult.Msg, "token": token})
	} else {
		context.JSON(http.StatusOK, gin.H{"status": loginResult.Status, "msg": loginResult.Msg})
	}
}

func adminLoginHandler(context *gin.Context) {
	username := context.PostForm("username")
	password := context.PostForm("password")
	loginResult, userID := agent.AuthenticateAdmin(username, password)
	if loginResult.Status == AdminLoginOK {
		token := util.GenToken(userID, util.AdminKey)
		context.JSON(http.StatusOK, gin.H{"status": loginResult.Status, "msg": loginResult.Msg, "token": token})
	} else {
		context.JSON(http.StatusOK, gin.H{"status": loginResult.Status, "msg": loginResult.Msg, "token": ""})
	}
}

func registerHandler(context *gin.Context) {
	username := context.PostForm("username")
	password := context.PostForm("password")
	registerResult := agent.RegisterUser(username, password)
	context.JSON(http.StatusOK, gin.H{"status": registerResult.Status, "msg": registerResult.Msg})
}

func getCountHandler(context *gin.Context) {
	bookCount := agent.GetBookNum()
	context.JSON(http.StatusOK, gin.H{"count": bookCount})
}

func getBooksHandler(context *gin.Context) {
	pageString := context.PostForm("page")
	page, _ := strconv.Atoi(pageString)
	books := agent.GetBooksByPage(page)

	bf := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(bf)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(books)

	_, _ = context.Writer.Write(bf.Bytes())
}

func getBorrowTimeHandler(context *gin.Context) {
	bookIDString := context.PostForm("bookID")
	bookID, _ := strconv.Atoi(bookIDString)
	UserIDString := context.PostForm("userID")
	userID, _ := strconv.Atoi(UserIDString)
	subTime := agent.GetBorrowTime(bookID, userID)

	bf := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(bf)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(subTime)

	_, _ = context.Writer.Write(bf.Bytes())
}

func getUserBooksHandler(context *gin.Context) {
	iUserID, _ := context.Get("userID")
	userID := iUserID.(int)
	pageString := context.PostForm("page")
	page, _ := strconv.Atoi(pageString)
	books := agent.GetUserBooksByPage(userID, page)

	bf := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(bf)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(books)

	_, _ = context.Writer.Write(bf.Bytes())
}

func getReserveBooksHandler(context *gin.Context) { //数据类型待定（不知道是int还是string）,则前面的可能要改
	iUserID, _ := context.Get("userID")
	userID := iUserID.(int)
	bookIDString := context.PostForm("bookID")
	bookID, _ := strconv.Atoi(bookIDString)
	result := agent.ReserveBook(userID, bookID)
	context.JSON(http.StatusOK, gin.H{"status": result.Status, "msg": result.Msg})
}

func getCancelReserveBooksHandler(context *gin.Context) {
	iUserID, _ := context.Get("userID")
	userID := iUserID.(int)
	bookIDString := context.PostForm("bookID")
	bookID, _ := strconv.Atoi(bookIDString)
	result := agent.CancelReserveBook(userID, bookID)
	context.JSON(http.StatusOK, gin.H{"status": result.Status, "msg": result.Msg})
}

func borrowBookHandler(context *gin.Context) {
	iUserID, _ := context.Get("userID")
	userID := iUserID.(int)
	bookIDString := context.PostForm("bookID")
	bookID, _ := strconv.Atoi(bookIDString)
	result := agent.BorrowBook(userID, bookID)
	context.JSON(http.StatusOK, gin.H{"status": result.Status, "msg": result.Msg})
}

func returnBookHandler(context *gin.Context) {
	iUserID, _ := context.Get("userID")
	userID := iUserID.(int)
	bookIDString := context.PostForm("bookID")
	bookID, _ := strconv.Atoi(bookIDString)
	result := agent.ReturnBook(userID, bookID)
	context.JSON(http.StatusOK, gin.H{"status": result.Status, "msg": result.Msg})
}

func updateBookStatusHandler(context *gin.Context) {
	bookStatusString := context.PostForm("bookStatus")
	bookStatusMap := make(map[string]string)
	err := json.Unmarshal([]byte(bookStatusString), &bookStatusMap)
	if err != nil {
		log.Println(err.Error())
	}
	book := new(Book)
	book.Id, _ = strconv.Atoi(bookStatusMap["id"])
	book.Name = bookStatusMap["name"]
	book.Author = bookStatusMap["author"]
	book.Isbn = bookStatusMap["isbn"]
	book.Address = bookStatusMap["address"]
	book.Language = bookStatusMap["language"]
	book.Count, _ = strconv.Atoi(bookStatusMap["count"])
	result := agent.UpdateBookStatus(book)
	context.JSON(http.StatusOK, gin.H{"status": result.Status, "msg": result.Msg})
}

// /addbook?isbn=&count=&location=
func addBookHandler(context *gin.Context) {
	var err error
	bookStatusString := context.PostForm("bookStatus")
	bookStatusMap := make(map[string]string)
	err = json.Unmarshal([]byte(bookStatusString), &bookStatusMap)
	if err != nil {
		log.Println(err.Error())
	}
	isbn := bookStatusMap["isbn"]
	count := bookStatusMap["count"]
	location := bookStatusMap["location"]
	var book Book

	book, err = GetMetaDataByISBN(isbn)
	if err != nil {
		log.Println("metadata retriever failure: " + err.Error())
		book.Name = "Unknown"
		book.Author = "Unknown"
		book.Language = "Unknown"
		book.Isbn = isbn
	}
	book.Count, _ = strconv.Atoi(count)
	book.Location = location
	result := agent.AddBook(&book)
	if result.Status == UpdateOK {
		log.Printf("Add Book %v (ISBN:%v) Successfully \n", book.Name, book.Isbn)
	} else {
		log.Printf("FAIL TO Add Book %v (ISBN:%v)  \n", book.Name, book.Isbn)
	}
	context.JSON(http.StatusOK, gin.H{"status": result.Status, "msg": result.Msg})
}

func deleteBookHandler(context *gin.Context) {
	bookID, err := strconv.Atoi(context.PostForm("bookID"))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	result := agent.DeleteBook(bookID)
	context.JSON(http.StatusOK, gin.H{"status": result.Status, "msg": result.Msg})
}

// ------Book BarCode Handle Section Start-------

// GetBookBarimg Handler:
// Method:GET param={id,isbn}
func getBookBarcodeImageHandler(context *gin.Context) {
	idString := context.Query("id")
	id, _ := strconv.Atoi(idString)
	isbn := context.Query("isbn")
	result, path := agent.GetBookBarcodePath(id, isbn)
	if result.Status == BookBarcodeFailed {
		log.Println(result.Msg)
		context.Data(http.StatusInternalServerError, "image/png", nil)
		return
	} else {
		data, err := os.ReadFile(path)
		if err != nil {
			log.Println(err.Error())
			context.Data(http.StatusInternalServerError, "image/png", nil)
		}
		context.Data(http.StatusOK, "image/png", data)
	}
}

// send email

func AnnouncementHandler(context *gin.Context) {
     UserID := context.PostForm("UseID")
                                // 需要用户id来查询邮箱 status 三种状态 1 有罚款  2 有预定书籍 3 二者都有
     contentStatus := context.PostForm("Status")

     contentStatusNum, _ := strconv.Atoi(contentStatus)

     UserIDN, _ := strconv.Atoi(UserID)

     result := agent.MailQuery(UserIDN)

     email_address := result.Msg

     if result.Status == NOEmailAddress{

          result.Msg = "No emailAddress"

		  context.JSON(http.StatusOK, gin.H{"status": result.Status, "msg": result.Msg})
          return
	 }

     stmpAddr := "stmp.qq.com" // 邮件服务器地址 若qq邮箱 则为 stmp.qq.com

     port := "25" // qq邮箱端口

     secret := "utpqwjsghhfbhdii"   //邮箱密钥

     fromAddr := "1483681501@qq.com" // 用来发送邮件的邮箱地址

	 contentType := "Content-Type: text/html; charset=UTF-8"

	 sendAddr := stmpAddr+":"+port

	 content := ""

	 switch contentStatusNum{
	    case 1: content = "You have fine, need to be paid, or you can't borrow or reserve books"

	    case 2: content = "You have reserved books, please take them in 4h, or they will be cancelled."

	    case 3: content = "You have both find and reserved books, please login the website for more details."
	 
	 }

	 auth := smtp.PlainAuth("", fromAddr, secret, stmpAddr)

     msg := []byte("To: " + email_address + "\r\n" +
		 "From: " + fromAddr + "\r\n" +
		 "Subject: " + "Announcement from Library" + "\r\n" +
		 contentType + "\r\n\r\n" +
		 "<html><h1>" + content + "</h1></html>")

     err := smtp.SendMail(sendAddr, auth, fromAddr, []string{email_address}, msg)

     if err != nil{

     	result.Status = SendEmailFail
     	result.Msg = "SendEmailFail"
		 context.JSON(http.StatusOK, gin.H{"status": result.Status, "msg": result.Msg})
     	return
	 }

	   result.Status = SendEmailOK
	   result.Msg = "SendEmailOK"
	   context.JSON(http.StatusOK, gin.H{"status": result.Status, "msg": result.Msg})

}


// ------Book BarCode Handle Section End-------
func loadConfig(configPath string) {
	Cfg, err := ini.Load(configPath)
	if err != nil {
		log.Fatal("Fail to Load config: ", err)
	}

	server, err := Cfg.GetSection("server")
	if err != nil {
		log.Fatal("Fail to load section 'server': ", err)
	}
	httpPort := server.Key("port").MustInt(80)
	path := server.Key("path").MustString("")
	staticPath := server.Key("staticPath").MustString("")
	Jikeapikey = server.Key("JiKeAPIKey").MustString("")

	mysql, err := Cfg.GetSection("mysql")
	if err != nil {
		log.Fatal("Fail to load section 'mysql': ", err)
	}
	username := mysql.Key("username").MustString("")
	password := mysql.Key("password").MustString("")
	address := mysql.Key("address").MustString("")
	tableName := mysql.Key("table").MustString("")

	db, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v)/%v?parseTime=true", username, password, address, tableName))
	if err != nil {
		panic("connect to DB failed: " + err.Error())
	}
	agent.DB = db

	MediaPath = filepath.Join(path, "media")

	err = os.MkdirAll(MediaPath, os.ModePerm)
	if err != nil {
		log.Fatal("file system failed to create path: " + err.Error())
	}
	startService(httpPort, path, staticPath)

}

func startService(port int, path string, staticPath string) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	//router.LoadHTMLFiles(fmt.Sprintf("%v/index.html", path))
	//router.Use(static.Serve("/static", static.LocalFile(staticPath, true)))

	router.GET("/", func(context *gin.Context) {
		context.HTML(http.StatusOK, "index.html", nil)
	})
	//router.GET("/test", func(context *gin.Context) {
	//	context.String(http.StatusOK, "test")
	//})

	g1 := router.Group("/")
	g1.Use(middleware.UserAuth())
	{
		g1.POST("/getUserBooks", getUserBooksHandler)
		g1.POST("/getBorrowTime", getBorrowTimeHandler)
		g1.POST("/reserveBooks", getReserveBooksHandler)
		g1.POST("/cancelReserveBooks", getCancelReserveBooksHandler)
		g1.POST("/borrowBook", borrowBookHandler)
		g1.POST("/returnBook", returnBookHandler)
		g1.POST("/getFine", AnnouncementHandler)
	}

	g2 := router.Group("/")
	g2.Use(middleware.AdminAuth())
	{
		g2.POST("/updateBookStatus", updateBookStatusHandler)
		g2.POST("/deleteBook", deleteBookHandler)
		g2.POST("/addBook", addBookHandler)
	}

	router.POST("/login", loginHandler)
	router.POST("/admin", adminLoginHandler)
	router.POST("/register", registerHandler)
	router.GET("/getCount", getCountHandler)
	router.GET("/getBooks", getBooksHandler)
	router.POST("/getBooks", getBooksHandler)
	router.GET("/getBookBarcode", getBookBarcodeImageHandler)

	g3 := router.Group("/pay")
	{
		g3.GET("/mobile", AliPayHandlerMobile)
		g3.GET("/pc", AliPayHandlerPC)
		g3.GET("/signcheck", SignCheck)
	}
	//router.StaticFile("/favicon.ico", fmt.Sprintf("%v/favicon.ico", staticPath))

	err := router.Run(":" + strconv.Itoa(port))
	if err != nil {
		fmt.Println(err)
		return
	} else {
		log.Println("running")
		return
	}
}

func main() {
	var configPath = flag.String("config", "./app.ini", "配置文件路径")
	flag.Parse()
	loadConfig(*configPath)
}
