package main

import (
	"context"
	"fmt"
	connection "gola1/conection"
	"log"
	"net/http"
	"strconv"
	"text/template"
	"time"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type Blog struct {
	ID          int
	Title       string
	Description string
	StartDate   string
	EndDate     string
	Author      string
	Duration    string
	Image       string
	Animal      bool
	Human       bool
	Demon       bool
	Robot       bool
}

type User struct {
	ID       int
	Name     string
	Email    string
	Password string
}

type UserData struct {
	IsLogin bool
	Name    string
	Email   string
}

type Profile struct {
	ID      int
	Phone   string
	Address string
	Hoby    string
}

var userData = UserData{}

// var dataBlog = []Blog{
// 	{
// 		Title:       "Hallo Title 1",
// 		Description: "Halo Content 1",
// 		Author:      "Alex",
// 		Image:       "franky.jpg",
// 	},
// 	{
// 		Title:       "Hallo Title 2",
// 		Description: "Halo Content 2",
// 		Author:      "Alexis",
// 		Image:       "nami.jpg",
// 	},
// }

func main() {
	connection.DatabaseConnection()

	e := echo.New()

	e.Static("/public", "public")
	e.Static("/uplouds", "uplouds")

	e.Use(session.Middleware(sessions.NewCookieStore([]byte("session"))))

	//GET
	e.GET("/home", home)
	e.GET("/contact", contact)
	e.GET("/blog", blog)
	e.GET("/blog-detail/:id", blogDetail)

	//ADD_BLOG
	e.GET("/form-blog", formAddBlog)
	e.POST("/add-blog", addBlog)

	//EDIT
	e.GET("/blog-edit/:id", formEditBlog)
	e.POST("/blog-edit/:id", editBlog)

	//REGISTER
	e.GET("/", registerForm)
	e.POST("/register", register)

	//LOGIN
	e.GET("/login-form", loginForm)
	e.POST("/login", login)

	//LOGOUT_DELETE
	e.POST("/logout", logout)
	e.POST("/blog-delete/:id", deleteBlog)

	//PROFILE
	e.GET("/profile", profile)
	e.GET("/edit-profile/:id", profileEditForm)
	e.POST("/edit-profile/:id", profileEdit)

	e.Logger.Fatal(e.Start("localhost:5000"))
}

// GET
func home(c echo.Context) error {
	data, _ := connection.Conn.Query(context.Background(), "SELECT id, title, description, image, start_date, end_date, animal, human, demon, robot, duration FROM tb_blog")

	var result []Blog
	for data.Next() {
		var each = Blog{}

		err := data.Scan(&each.ID, &each.Title, &each.Description, &each.Image, &each.StartDate, &each.EndDate, &each.Animal, &each.Human, &each.Demon, &each.Robot, &each.Duration)
		if err != nil {
			fmt.Println(err.Error())
			return c.JSON(http.StatusInternalServerError, map[string]string{"Message": err.Error()})
		}

		each.Author = "Alex"

		result = append(result, each)
	}

	sess, _ := session.Get("session", c)

	if sess.Values["isLogin"] != true {
		userData.IsLogin = false
	} else {
		userData.IsLogin = sess.Values["isLogin"].(bool)
		userData.Name = sess.Values["name"].(string)
	}

	blogs := map[string]interface{}{
		"Blogs":        result,
		"Flashstatus":  sess.Values["status"],
		"Flashmessage": sess.Values["message"],
		"DataSession":  userData,
	}

	delete(sess.Values, "message")
	delete(sess.Values, "status")
	sess.Save(c.Request(), c.Response())

	var tmpl, err = template.ParseFiles("views/index.html")

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}

	return tmpl.Execute(c.Response(), blogs)
}

func contact(c echo.Context) error {
	var tmpl, err = template.ParseFiles("views/contact.html")

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}

	sess, _ := session.Get("session", c)

	if sess.Values["isLogin"] != true {
		userData.IsLogin = false
	} else {
		userData.IsLogin = sess.Values["isLogin"].(bool)
		userData.Name = sess.Values["name"].(string)
	}

	blogs := map[string]interface{}{
		"DataSession": userData,
	}

	return tmpl.Execute(c.Response(), blogs)
}

func blog(c echo.Context) error {
	data, _ := connection.Conn.Query(context.Background(), "SELECT tb_blog.id, title, description, image, start_date, end_date, animal, human, demon, robot, duration, tb_user.name AS author FROM tb_blog JOIN tb_user ON tb_blog.author_id ORDER BY tb_blog.id DESC")

	var result []Blog
	for data.Next() {
		var each = Blog{}

		err := data.Scan(&each.ID, &each.Title, &each.Description, &each.Image, &each.StartDate, &each.EndDate, &each.Animal, &each.Human, &each.Demon, &each.Robot, &each.Duration)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"Message": err.Error()})
		}

		result = append(result, each)
	}

	sess, _ := session.Get("session", c)

	if sess.Values["isLogin"] != true {
		userData.IsLogin = false
	} else {
		userData.IsLogin = sess.Values["isLogin"].(bool)
		userData.Name = sess.Values["name"].(string)
	}

	var tmpl, errtemplate = template.ParseFiles("views/blog.html")

	if errtemplate != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": errtemplate.Error()})
	}

	blogs := map[string]interface{}{
		"Blogs":       result,
		"DataSession": userData,
	}

	return tmpl.Execute(c.Response(), blogs)
}

func blogDetail(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	var BlogDetail = Blog{}

	err := connection.Conn.QueryRow(context.Background(), "SELECT tb_blog.id, title, description, image, start_date, end_date, animal, human, demon, robot, duration, tb_user.name AS author FROM tb_blog JOIN tb_user ON tb_blog.author_id = tb_user.id WHERE tb_blog.id=$1", id).Scan(
		&BlogDetail.ID, &BlogDetail.Title, &BlogDetail.Description, &BlogDetail.Image, &BlogDetail.StartDate, &BlogDetail.EndDate, &BlogDetail.Animal, &BlogDetail.Human, &BlogDetail.Demon, &BlogDetail.Robot, &BlogDetail.Duration, &BlogDetail.Author)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}

	sess, _ := session.Get("session", c)

	if sess.Values["isLogin"] != true {
		userData.IsLogin = false
	} else {
		userData.IsLogin = sess.Values["isLogin"].(bool)
		userData.Name = sess.Values["isLogin"].(string)
	}

	data := map[string]interface{}{
		"Blog":        BlogDetail,
		"DataSession": userData,
	}

	var tmpl, errtemplate = template.ParseFiles("views/blog-detail.html")

	if errtemplate != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}

	return tmpl.Execute(c.Response(), data)
}

// ADD_BLOG
func formAddBlog(c echo.Context) error {
	var tmpl, err = template.ParseFiles("views/form-blog.html")

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}

	sess, _ := session.Get("session", c)

	if sess.Values["isLogin"] != true {
		userData.IsLogin = true
	} else {
		userData.IsLogin = sess.Values["isLogin"].(bool)
		userData.Name = sess.Values["name"].(string)
	}

	blogs := map[string]interface{}{
		"DataSession": userData,
	}

	return tmpl.Execute(c.Response(), blogs)
}

func addBlog(c echo.Context) error {
	title := c.FormValue("input-tittle")
	description := c.FormValue("input-description")
	image := c.Get("input-image").(string)
	startdate := c.FormValue("input-start-date")
	enddate := c.FormValue("input-end-date")
	duration := countDuration(startdate, enddate)

	var animal bool
	if c.FormValue("check-animal") == "yes" {
		animal = true
	}

	var human bool
	if c.FormValue("check-human") == "yes" {
		human = true
	}

	var demon bool
	if c.FormValue("check-demon") == "yes" {
		demon = true
	}

	var robot bool
	if c.FormValue("check-robot") == "yes" {
		robot = true
	}

	sess, _ := session.Get("session", c)
	author := sess.Values["id"].(int)

	_, err := connection.Conn.Exec(context.Background(), "INSERT INTO tb_blog (title, description, start_date, end_date, image, animal, human, demon, robot, duration, author_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)", title, description, startdate, enddate, image, animal, human, demon, robot, duration, author)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}

	return c.Redirect(http.StatusMovedPermanently, "/blog")
}

// EDIT
func formEditBlog(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	var EditBlog = Blog{}

	err := connection.Conn.QueryRow(context.Background(), "SELECT id, title, description, image, start_date, end_date, animal, human, demon, robot, duration FROM tb_blog WHERE id=$1", id).Scan(
		&EditBlog.ID, &EditBlog.Title, &EditBlog.Description, &EditBlog.Image, &EditBlog.StartDate, &EditBlog.EndDate, &EditBlog.Animal, &EditBlog.Human, &EditBlog.Demon, &EditBlog.Robot, &EditBlog.Duration)

	Blogs := map[string]interface{}{
		"Blogs": EditBlog,
	}

	var tmpl, errtemplate = template.ParseFiles("views/form-edit-blog.html")

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": errtemplate.Error()})
	}

	return tmpl.Execute(c.Response(), Blogs)
}

func editBlog(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	title := c.FormValue("input-tittle")
	description := c.FormValue("input-description")
	image := c.FormValue("input-image")
	startdate := c.FormValue("input-start-date")
	enddate := c.FormValue("input-end-date")
	duration := countDuration(startdate, enddate)

	var animal bool
	if c.FormValue("check-animal") == "yes" {
		animal = true
	}

	var human bool
	if c.FormValue("check-human") == "yes" {
		human = true
	}

	var demon bool
	if c.FormValue("check-demon") == "yes" {
		demon = true
	}

	var robot bool
	if c.FormValue("check-robot") == "yes" {
		robot = true
	}

	_, err := connection.Conn.Exec(context.Background(), "UPDATE tb_blog SET title=$1, description=$2, start_date=$3, end_date=$4, image=$5, animal=$6, human=$7, demon=$8, robot=$9, duration=$10 WHERE id=$11", title, description, startdate, enddate, image, animal, human, demon, robot, duration, id)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}

	return c.Redirect(http.StatusMovedPermanently, "/blog")
}

// REGISTER
func registerForm(c echo.Context) error {
	var tmpl, err = template.ParseFiles("views/register-form.html")

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}

	return tmpl.Execute(c.Response(), nil)
}

func register(c echo.Context) error {
	err := c.Request().ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	name := c.FormValue("input-name")
	email := c.FormValue("input-email")
	password := c.FormValue("input-pw")

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)

	_, err = connection.Conn.Exec(context.Background(), "INSERT INTO tb_user(name, email, password) VALUES ($1, $2, $3)", name, email, passwordHash)

	if err != nil {
		redirectWithMessage(c, "Register Failed, Please Try Again", false, "/register-form")
	}

	return redirectWithMessage(c, "Register Success !", true, "/login-form")
}

// LOGIN
func loginForm(c echo.Context) error {
	sess, _ := session.Get("session", c)

	flash := map[string]interface{}{
		"FlashStatus":  sess.Values["status"],
		"FlashMessage": sess.Values["message"],
	}

	delete(sess.Values, "status")
	delete(sess.Values, "message")
	sess.Save(c.Request(), c.Response())

	var tmpl, err = template.ParseFiles("views/login-form.html")

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}

	return tmpl.Execute(c.Response(), flash)
}

func login(c echo.Context) error {
	err := c.Request().ParseForm()

	if err != nil {
		log.Fatal(err)
	}

	email := c.FormValue("input-email")
	password := c.FormValue("input-pw")

	user := User{}

	err = connection.Conn.QueryRow(context.Background(), "SELECT * FROM tb_user WHERE email=$1", email).Scan(
		&user.ID, &user.Name, &user.Email, &user.Password)
	if err != nil {
		return redirectWithMessage(c, "Email Incorrect", false, "/login-form")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return redirectWithMessage(c, "Password Incorrect", false, "/login-form")
	}

	sess, _ := session.Get("session", c)
	sess.Options.MaxAge = 108000
	sess.Values["message"] = "Login is Success !"
	sess.Values["status"] = true
	sess.Values["id"] = user.ID
	sess.Values["name"] = user.Name
	sess.Values["email"] = user.Email
	sess.Values["isLogin"] = true
	sess.Save(c.Request(), c.Response())

	return c.Redirect(http.StatusMovedPermanently, "/home")
}

// LOGOUT_DELETE
func deleteBlog(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	fmt.Println("ID : ", id)

	_, err := connection.Conn.Exec(context.Background(), "DELETE FROM tb_blog WHERE id=$1", id)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}

	return c.Redirect(http.StatusMovedPermanently, "/blog")
}

func logout(c echo.Context) error {
	sess, _ := session.Get("session", c)
	sess.Options.MaxAge = -1
	sess.Save(c.Request(), c.Response())

	return c.Redirect(http.StatusMovedPermanently, "/")
}

// PROFILE
func profile(c echo.Context) error {
	data, _ := connection.Conn.Query(context.Background(), "SELECT id, phone, address, hoby FROM tb_profile")

	var EditProfile []Profile

	for data.Next() {
		var each = Profile{}

		err := data.Scan(&each.ID, &each.Phone, &each.Address, &each.Hoby)

		if err != nil {
			fmt.Println(err.Error())
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
		}

		EditProfile = append(EditProfile, each)
	}

	sess, _ := session.Get("session", c)

	if sess.Values["isLogin"] != true {
		userData.IsLogin = false
	} else {
		userData.IsLogin = sess.Values["isLogin"].(bool)
		userData.Name = sess.Values["name"].(string)
		userData.Email = sess.Values["email"].(string)
	}

	EditProfiles := map[string]interface{}{
		"Blogs":       EditProfile,
		"EditProfile": EditProfile,
		"DataSession": userData,
	}

	var tmpl, errtemplate = template.ParseFiles("views/profile.html")

	if errtemplate != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": errtemplate.Error()})
	}

	return tmpl.Execute(c.Response(), EditProfiles)
}

func profileEditForm(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	var EditProfile = Profile{}

	err := connection.Conn.QueryRow(context.Background(), "SELECT id, phone, address, hoby FROM tb_profile WHERE id=$1", id).Scan(
		&EditProfile.ID, &EditProfile.Phone, &EditProfile.Address, &EditProfile.Hoby)

	Blogs := map[string]interface{}{
		"Blogs": EditProfile,
	}

	var tmpl, errtemplate = template.ParseFiles("views/edit-profile.html")

	if errtemplate != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}

	sess, _ := session.Get("session", c)

	if sess.Values["isLogin"] != true {
		userData.IsLogin = false
	} else {
		userData.IsLogin = sess.Values["isLogin"].(bool)
		userData.Name = sess.Values["name"].(string)
	}

	return tmpl.Execute(c.Response(), Blogs)
}

func profileEdit(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))

	phone := c.FormValue("input-number-profile")
	address := c.FormValue("input-address-profile")
	hoby := c.FormValue("input-hoby-profile")

	_, err := connection.Conn.Exec(context.Background(), "UPDATE tb_profile SET phone=$1, address=$2, hoby=$3 WHERE id=$4", phone, address, hoby, id)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": err.Error()})
	}

	return c.Redirect(http.StatusMovedPermanently, "/profile")
}

// FUNCTION
func countDuration(inputStartDate string, inputEndDate string) string {
	startDate, _ := time.Parse("2006-01-02", inputStartDate)
	endDate, _ := time.Parse("2006-01-02", inputEndDate)

	durationDate := int(endDate.Sub(startDate).Hours())
	durationDays := durationDate / 24
	durationMonths := durationDays / 30
	durationYears := durationMonths / 12

	var duration string

	if durationYears > 1 {
		duration = strconv.Itoa(durationYears) + "Years"
	} else if durationYears > 0 {
		duration = strconv.Itoa(durationYears) + "year"
	} else {
		if durationMonths > 1 {
			duration = strconv.Itoa(durationMonths) + "Months"
		} else if durationMonths > 0 {
			duration = strconv.Itoa(durationMonths) + "Month"
		} else {
			if durationDays > 1 {
				duration = strconv.Itoa(durationDays) + "Days"
			} else if durationDays > 0 {
				duration = strconv.Itoa(durationDays) + "Day"
			}
		}
	}

	return duration
}

func redirectWithMessage(c echo.Context, message string, status bool, path string) error {
	sess, _ := session.Get("session", c)
	sess.Values["message"] = message
	sess.Values["status"] = status
	sess.Save(c.Request(), c.Response())

	return c.Redirect(http.StatusMovedPermanently, path)
}
