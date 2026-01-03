package main
import (
	"database/sql"
	"fmt"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)
func main() {
	InitDB()
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.GET("/", func(c *gin.Context) { c.Redirect(http.StatusFound, "/login") })
	r.GET("/login", func(c *gin.Context) { c.HTML(http.StatusOK, "login.html", nil) })
	r.POST("/login", loginHandler)
	r.GET("/logout", logoutHandler)
	r.GET("/dashboard", authMiddleware(), dashboardHandler)
	r.GET("/admin/users", authMiddleware(), adminHandler)
	fmt.Println("Server starting on :8080...")
	r.Run(":8080")
}
func logoutHandler(c *gin.Context) {
	c.SetCookie("session_id", "", -1, "/", "", false, true)
	c.SetCookie("is_admin", "", -1, "/", "", false, false)
	c.Redirect(http.StatusFound, "/login")
}
func loginHandler(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	var id int
	var role string
	err := DB.QueryRow("SELECT id, role FROM users WHERE username=? AND password=?", username, password).Scan(&id, &role)
	if err == sql.ErrNoRows {
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{"Error": "Invalid credentials"})
		return
	}
	sessionToken := uuid.New().String()
	DB.Exec("INSERT INTO sessions (session_token, user_id) VALUES (?, ?)", sessionToken, id)
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("session_id", sessionToken, 3600, "/", "", false, true)
	
	// VULNERABLE LOGIC
	isAdmin := "false"
	if role == "admin" { isAdmin = "true" }
	c.SetCookie("is_admin", isAdmin, 3600, "/", "", false, false)
	c.Redirect(http.StatusFound, "/dashboard")
}
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie("session_id")
		if err != nil { c.Redirect(http.StatusFound, "/login"); c.Abort(); return }
		var userID int
		err = DB.QueryRow("SELECT user_id FROM sessions WHERE session_token=?", cookie).Scan(&userID)
		if err != nil { c.Redirect(http.StatusFound, "/login"); c.Abort(); return }
		c.Set("user_id", userID)
		c.Next()
	}
}
func dashboardHandler(c *gin.Context) {
	roleCookie, _ := c.Cookie("is_admin")
	isAdmin := roleCookie == "true"
	c.HTML(http.StatusOK, "dashboard.html", gin.H{"IsAdmin": isAdmin, "User": "Alice"})
}
func adminHandler(c *gin.Context) {
	roleCookie, _ := c.Cookie("is_admin")
	if roleCookie != "true" { c.JSON(http.StatusForbidden, gin.H{"error": "Access Denied: Admin Privileges Required"}); return }
	
	// JUICY DATA LEAK
	type SensitiveProfile struct {
		ID int `json:"user_id"`
		Username string `json:"username"`
		FullName string `json:"full_name"`
		Email string `json:"email"`
		Phone string `json:"phone_number"`
		Address string `json:"home_address"`
		SSN string `json:"ssn"`
		CreditScore int `json:"credit_score"`
		MothersMaiden string `json:"security_question_mom"`
		Balance string `json:"account_balance"`
		LastLogin string `json:"last_login_ip"`
		AccountType string `json:"account_tier"`
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"query_time": "0.024s",
		"total_results": 2,
		"data": []SensitiveProfile{
			{101, "alice", "Alice P. Hacker", "alice@gopher.com", "+1-555-0199", "123 Cyber Lane, Silicon Valley, CA", "992-00-1122", 720, "Smith", "$1,050.00", "192.168.1.10", "Standard"},
			{102, "admin", "System Administrator", "root@gopher.internal", "+1-555-0000", "Server Room B, Data Center 4", "000-00-0000", 850, "R00T", "$9,999,999.00", "10.0.0.5", "Platinum_Admin"},
		},
	})
}