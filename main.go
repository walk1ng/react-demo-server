package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"database/sql"

	"github.com/gin-gonic/gin"

	_ "github.com/go-sql-driver/mysql"

	"github.com/dgrijalva/jwt-go"
	"github.com/walk1ng/react-demo-server/k8s"
)

type Student struct {
	Id     int    `json:"id,omitempty"`
	Name   string `json:"name,omitempty"`
	Age    int    `json:"age,omitempty"`
	Gender string `json:"gender,omitempty"`
}

type Resp struct {
	Code    int         `json:"code,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

type PageResp struct {
	Total int       `json:"total,omitempty"`
	List  []Student `json:"list,omitempty"`
}

type RouteAndMenu struct {
	MenuTree  []MenuWithChildrend `json:"menuTree,omitempty"`
	RouteList []Route             `json:"routeList,omitempty"`
}

type Menu struct {
	ID           int    `json:"id,omitempty"`
	Title        string `json:"title,omitempty"`
	PID          int    `json:"pid,omitempty"`
	Icon         string `json:"icon,omitempty"`
	RoutePath    string `json:"routePath,omitempty"`
	RouteElement string `json:"routeElement,omitempty"`
}

type MenuWithChildrend struct {
	Menu
	Children []Menu `json:"children,omitempty"`
}

type Route struct {
	Path    string `json:"path,omitempty"`
	Element string `json:"element,omitempty"`
}

type LoginReq struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type LoginResp struct {
	Token string `json:"token,omitempty"`
}

var db *sql.DB
var err error
var jwtSecret = []byte("your_jwt_secret_key")
var jwtExpirationTime = time.Hour * 24 // 设置为一天，可以根据需求进行调整

func init() {

	// db
	// 连接数据库

	// get mysql user and password from env
	mysqlUser := os.Getenv("MYSQL_USER")
	mysqlPassword := os.Getenv("MYSQL_PASSWORD")
	mysqlDB := os.Getenv("MYSQL_DB")
	mysqlSvc := os.Getenv("MYSQL_SVC")

	dbConnStr := fmt.Sprintf("%s:%s@tcp(%s)/%s", mysqlUser, mysqlPassword, mysqlSvc, mysqlDB)
	fmt.Println(dbConnStr)

	// db, err = sql.Open("mysql", "root:your-root-password@tcp(192.168.0.107:3306)/itcast")
	db, err = sql.Open("mysql", dbConnStr)
	if err != nil {
		panic(err.Error())
	}
	// 尝试连接数据库
	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}

	fmt.Println("成功连接到MySQL数据库！")

}

func main() {
	r := gin.Default()
	r.Use(corsMiddleware())

	r.GET("/api/students/:id", GetStudentByID)
	r.GET("/api/students", GetStudents)
	r.GET("/api/students/q", GetStudentsPagination)
	r.POST("/api/students", InsertStudent)
	r.POST("/api/studentsx", InsertStudentNew)
	r.DELETE("/api/student/:id", DeleteStudent)
	r.DELETE("/api/students", DeleteStudents)
	r.PUT("/api/student/:id", UpdateStudent)
	r.GET("/api/menu/:username", GetMenuByUser)
	r.POST("/api/login", loginHandler)
	r.GET("/api/pods", PodHandler)
	r.GET("/api/namespaces", nsHandler)

	// r.OPTIONS("/api/students", func(ctx *gin.Context) {
	// 	return
	// })

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Authorization, Content-Type")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}

		c.Next()
	}
}
func GetStudentByID(c *gin.Context) {

	id, _ := strconv.Atoi(c.Param("id"))
	var name, gender string
	var age int

	rows, err := db.Query(fmt.Sprintf("SELECT name, gender, age FROM students where id = %d", id))
	if err != nil {
		panic(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&name, &gender, &age)
		if err != nil {
			panic(err.Error())
		}
		fmt.Println(id, name, gender, age)
	}

	if err = rows.Err(); err != nil {
		panic(err.Error())
	}

	stu := Student{
		Id:     id,
		Age:    age,
		Name:   name,
		Gender: gender,
	}

	resp := Resp{
		Code: 200,
		Data: stu,
	}

	c.Header("Access-Control-Allow-Origin", "*")
	time.Sleep(time.Second * 2)
	c.JSON(http.StatusOK, resp)
}

func GetStudentsPagination(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")

	var students []Student
	var size, page, offset int

	name := c.Query("name")
	gender := c.Query("gender")
	age := c.Query("age")
	size = tryParseAtoI(c.Query("size"))
	page = tryParseAtoI(c.Query("page"))
	offset = (page - 1) * size

	// gender
	var genderQ string
	if len(gender) != 0 {
		genderQ = "AND gender = '" + gender + "'"
	}
	var ageQ string
	if len(age) != 0 {
		scope := strings.Split(age, ",")
		min := scope[0]
		max := scope[1]
		ageQ = fmt.Sprintf("AND age between %s and %s", min, max)
	}

	// basic
	query := fmt.Sprintf("SELECT id, name, gender, age FROM students WHERE name LIKE ? %s %s LIMIT %d OFFSET %d", genderQ, ageQ, size, offset)
	fmt.Println(query)

	rows, err := db.Query(query, "%"+name+"%")
	if err != nil {
		panic(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var student Student
		err = rows.Scan(&student.Id, &student.Name, &student.Gender, &student.Age)
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("%+v\n", student)
		students = append(students, student)
	}

	if err = rows.Err(); err != nil {
		panic(err.Error())
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM students").Scan(&count); err != nil {
		panic(err.Error())
	}

	resp := Resp{
		Code: 200,
		Data: &PageResp{
			List:  students,
			Total: count,
		},
	}

	// time.Sleep(time.Second * 2)
	c.JSON(http.StatusOK, resp)
}

func GetStudents(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")

	var students []Student
	var name, gender string
	var id, age int

	rows, err := db.Query("SELECT id, name, gender, age FROM students")
	if err != nil {
		panic(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&id, &name, &gender, &age)
		if err != nil {
			panic(err.Error())
		}
		fmt.Println(id, name, gender, age)

		stu := Student{
			Id:     id,
			Age:    age,
			Name:   name,
			Gender: gender,
		}
		students = append(students, stu)
	}

	if err = rows.Err(); err != nil {
		panic(err.Error())
	}

	resp := Resp{
		Code: 200,
		Data: students,
	}

	time.Sleep(time.Second * 2)
	c.JSON(http.StatusOK, resp)
}

func InsertStudent(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")

	name := c.PostForm("name")
	age := tryParseAtoI(c.PostForm("age"))
	gender := c.PostForm("gender")

	// 执行插入语句
	insertStmt, err := db.Prepare("INSERT INTO students (name, age, gender) VALUES (?, ?, ?)")
	if err != nil {
		panic(err.Error())
	}
	defer insertStmt.Close()

	result, err := insertStmt.Exec(name, age, gender)
	if err != nil {
		panic(err.Error())
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("成功插入 %d 行数据\n", affectedRows)

	stu := Student{
		Name:   name,
		Age:    age,
		Gender: gender,
	}

	fmt.Println(stu)
	c.JSON(http.StatusOK, Resp{
		Code:    200,
		Data:    stu,
		Message: "新增成功",
	})
}

func InsertStudentNew(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")

	var student Student
	if err := c.ShouldBindJSON(&student); err != nil {
		panic(err.Error())
	}

	// 执行插入语句
	insertStmt, err := db.Prepare("INSERT INTO students (name, age, gender) VALUES (?, ?, ?)")
	if err != nil {
		panic(err.Error())
	}
	defer insertStmt.Close()

	result, err := insertStmt.Exec(student.Name, student.Age, student.Gender)
	if err != nil {
		panic(err.Error())
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("成功插入 %d 行数据\n", affectedRows)

	c.JSON(http.StatusOK, Resp{
		Code:    200,
		Data:    student,
		Message: "新增成功",
	})
}
func DeleteStudent(c *gin.Context) {
	// c.Header("Access-Control-Allow-Origin", "*")
	// c.Header("Access-Control-Allow-Methods", "DELETE")

	id := tryParseAtoI(c.Param("id"))

	query := "DELETE FROM students where id = ?"
	result, err := db.Exec(query, id)
	if err != nil {
		panic(err.Error())
	}

	affecedRows, err := result.RowsAffected()
	if err != nil {
		panic(err.Error())
	}

	fmt.Println("受影响的行数:", affecedRows)

	fmt.Println(id)
	c.JSON(http.StatusOK, Resp{
		Code: 200,
		Data: "删除成功",
	})
}

func DeleteStudents(c *gin.Context) {
	// c.Header("Access-Control-Allow-Origin", "*")
	// c.Header("Access-Control-Allow-Methods", "DELETE")

	var param struct {
		IDs []int `json:"ids,omitempty"`
	}

	if err := c.ShouldBindJSON(&param); err != nil {
		panic(err.Error())
	}

	query := "DELETE FROM students where id in (?)"

	var placeholder []string = make([]string, len(param.IDs))
	for i := range param.IDs {
		placeholder[i] = "?"
	}
	inQuery := strings.Join(placeholder, ", ")
	query = strings.Replace(query, "?", inQuery, 1)

	// result, err := db.Exec(query, 19, 20, 21)
	var interfaceIDs []interface{}
	for _, id := range param.IDs {
		interfaceIDs = append(interfaceIDs, id)

	}
	result, err := db.Exec(query, interfaceIDs...)
	if err != nil {
		panic(err.Error())
	}

	affecedRows, err := result.RowsAffected()
	if err != nil {
		panic(err.Error())
	}

	fmt.Println("受影响的行数:", affecedRows)

	fmt.Println(param.IDs)
	c.JSON(http.StatusOK, Resp{
		Code: 200,
		Data: "删除成功",
	})
}

func UpdateStudent(c *gin.Context) {
	// c.Header("Access-Control-Allow-Origin", "*")
	// c.Header("Access-Control-Allow-Methods", "DELETE")

	id := tryParseAtoI(c.Param("id"))

	var student Student
	if err := c.ShouldBindJSON(&student); err != nil {
		panic(err.Error())
	}

	query := "UPDATE students set name = ?, age = ?, gender = ? where id = ?"
	result, err := db.Exec(query, student.Name, student.Age, student.Gender, id)
	if err != nil {
		panic(err.Error())
	}

	affecedRows, err := result.RowsAffected()
	if err != nil {
		panic(err.Error())
	}

	fmt.Println("受影响的行数:", affecedRows)

	fmt.Println(id)
	c.JSON(http.StatusOK, Resp{
		Code: 200,
		Data: "修改成功",
	})
}

func GetMenuByUser(c *gin.Context) {
	username := c.Param("username")

	// routesQuery := fmt.Sprintf("select id, title, pid, icon, route_path, route_element from menu where id in (select menu_id from role_menu where role_id = (select role_id from user_role where user_id = (select id from user where username = '%s'))) and route_path is not null", username)
	routesQuery := fmt.Sprintf("select route_path, route_element from menu where id in (select menu_id from role_menu where role_id = (select role_id from user_role where user_id = (select id from user where username = '%s'))) and route_path is not null", username)
	rows, err := db.Query(routesQuery)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var routeList []Route
	for rows.Next() {
		var route Route
		err = rows.Scan(&route.Path, &route.Element)
		if err != nil {
			panic(err.Error())
		}
		routeList = append(routeList, route)
	}
	fmt.Println(routeList)

	if err = rows.Err(); err != nil {
		panic(err.Error())
	}

	// query parent menu
	parentMenuQuery := fmt.Sprintf("select id, title, pid, icon, route_path, route_element from menu where id in (select menu_id from role_menu where role_id = (select role_id from user_role where user_id = (select id from user where username = '%s'))) and pid = 0", username)
	rows, err = db.Query(parentMenuQuery)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var parentMenus []MenuWithChildrend
	for rows.Next() {
		var menu Menu
		var routePath, routeElement sql.NullString
		err = rows.Scan(&menu.ID, &menu.Title, &menu.PID, &menu.Icon, &routePath, &routeElement)
		if err != nil {
			panic(err.Error())
		}

		if routePath.Valid {
			menu.RoutePath = routePath.String
		}

		if routeElement.Valid {
			menu.RouteElement = routeElement.String
		}

		parentMenus = append(parentMenus, MenuWithChildrend{Menu: menu})
	}
	fmt.Println(parentMenus)
	if err = rows.Err(); err != nil {
		panic(err.Error())
	}

	// query children menu

	for i, pm := range parentMenus {
		childrenMenuQuery := fmt.Sprintf("select id, title, pid, icon, route_path, route_element from menu where id in (select menu_id from role_menu where role_id = (select role_id from user_role where user_id = (select id from user where username = '%s'))) and pid = ?", username)
		rows, err = db.Query(childrenMenuQuery, pm.ID)
		if err != nil {
			panic(err)
		}
		defer rows.Close()

		var childMenus []Menu
		for rows.Next() {
			var menu Menu
			var routePath, routeElement sql.NullString
			err = rows.Scan(&menu.ID, &menu.Title, &menu.PID, &menu.Icon, &routePath, &routeElement)
			if err != nil {
				panic(err.Error())
			}

			if routePath.Valid {
				menu.RoutePath = routePath.String
			}

			if routeElement.Valid {
				menu.RouteElement = routeElement.String
			}
			childMenus = append(childMenus, menu)
		}

		if err = rows.Err(); err != nil {
			panic(err.Error())
		}

		parentMenus[i].Children = append(parentMenus[i].Children, childMenus...)
		fmt.Println(parentMenus)
	}

	routeAndMenu := &RouteAndMenu{
		MenuTree:  parentMenus,
		RouteList: routeList,
	}

	c.JSON(http.StatusOK, &Resp{
		Code: 200,
		Data: routeAndMenu,
	})

}

func Login(c *gin.Context) {
	// mock
	availableUsers := []string{"admin", "li", "zhang"}
	password := "123"

	var loginReq LoginReq
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		panic(err.Error())
	}

	if InSlice(loginReq.Username, availableUsers) && loginReq.Password == password {
		c.JSON(http.StatusOK, &Resp{
			Code: 999,
			Data: &LoginResp{
				Token: "success string",
			},
			Message: "登录成功",
		})
	} else {
		c.JSON(http.StatusOK, &Resp{
			Code: http.StatusUnauthorized,
			Data: &LoginResp{
				Token: "",
			},
			Message: "登录失败",
		})
	}
}

func loginHandler(c *gin.Context) {
	var loginReq LoginReq
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		panic(err.Error())
	}
	// 在这里进行身份验证，检查用户名和密码是否正确
	query := "select id from user where username = ? and password = ?"
	rows, err := db.Query(query, loginReq.Username, loginReq.Password)
	if err != nil {
		panic(err.Error())
	}

	defer rows.Close()

	var userid int
	for rows.Next() {
		err := rows.Scan(&userid)
		if err != nil {
			panic(err.Error())
		}
		fmt.Println("userid:", userid)
	}

	if err = rows.Err(); err != nil {
		panic(err.Error())
	}

	if userid == 0 {
		c.JSON(http.StatusOK, &Resp{
			Code: http.StatusUnauthorized,
			Data: &LoginResp{
				Token: "",
			},
			Message: "登录失败",
		})
		return
	}

	// generate token by username

	// 假设验证成功并获得用户ID
	// userId := "12345"

	// 生成 JWT Token
	token, err := generateToken(loginReq.Username, userid)

	if err != nil {
		// 处理生成 Token 错误
		c.JSON(http.StatusInternalServerError,
			&Resp{
				Code: http.StatusInternalServerError,
				Data: &LoginResp{
					Token: "",
				},
				Message: "Failed to generate token",
			})
		return
	}

	// 返回 Token 给客户端
	c.JSON(http.StatusOK, &Resp{
		Code: 999,
		Data: &LoginResp{
			Token: token,
		},
		Message: "登录成功",
	})
}

func PodHandler(c *gin.Context) {
	namespace := c.Query("namespace")
	if len(namespace) == 0 {
		namespace = "default"
	}

	podInfo, err := k8s.GetPodsByNS(namespace)
	if err != nil {
		panic(err.Error())
	}

	time.Sleep(time.Second * 2)

	c.JSON(http.StatusOK, &Resp{
		Code: 999,
		Data: podInfo,
	})
}

func nsHandler(c *gin.Context) {
	nsInfo, err := k8s.GetNs()
	if err != nil {
		panic(err.Error())
	}

	c.JSON(http.StatusOK, &Resp{
		Code: 999,
		Data: nsInfo,
	})
}

func tryParseAtoI(a string) int {
	i, _ := strconv.Atoi(a)
	return i
}

func InSlice(a string, s []string) bool {
	for _, v := range s {
		if a == v {
			return true
		}
	}
	return false
}

func generateToken(username string, userId int) (string, error) {
	tokenClaims := jwt.MapClaims{
		"username": username,
		"userId":   userId,
		"exp":      time.Now().Add(jwtExpirationTime).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)
	tokenString, err := token.SignedString(jwtSecret)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}
