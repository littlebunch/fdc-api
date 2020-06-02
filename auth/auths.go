package auth

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/littlebunch/fdc-api/ds"
	fdc "github.com/littlebunch/fdc-api/model"
	"golang.org/x/crypto/bcrypt"
)

//User is basic authentication information
type User struct {
	ID       string `json:"_id"`
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email,omitempty"`
	Role     string `json:"role"`
	Type     string `json:"type"`
}
type login struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

// RoleType defines role for a User
type RoleType int

const (
	ADMIN RoleType = iota
	USER
)

// ToRole converts string to RoleType
func (r *RoleType) ToRole(t string) RoleType {
	switch t {
	case "ADMIN":
		return ADMIN
	case "USER":
		return USER
	default:
		return 0
	}
}

//ToString converts RoleType to string
func (r *RoleType) ToString(rt RoleType) string {
	switch rt {
	case ADMIN:
		return "ADMIN"
	case USER:
		return "USER"
	default:
		return ""
	}
}

var identityKey = "role"

// AuthMiddleware initializes our jwt components
func (u *User) AuthMiddleware(bucket string, d ds.DataSource) *jwt.GinJWTMiddleware {

	a, _ := jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "bfpd zone",
		Key:         []byte("secret key"),
		Timeout:     time.Hour,
		MaxRefresh:  time.Hour,
		IdentityKey: identityKey,
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var (
				u  User
				rc bool
				l  login
			)
			if err := c.BindJSON(&l); err != nil {
				return "", jwt.ErrMissingLoginValues
			}

			if u, rc = findUser(l.Username, d); rc == false {
				return "", jwt.ErrFailedAuthentication
			}

			if rc = CheckPasswordHash(l.Password, u.Password); rc == false {
				return "", jwt.ErrFailedAuthentication
			}
			return u, nil
		},
		// load user's roles into a []string and put into the claim
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(User); ok {
				//u, _ = findUser(v.ID, d)
				return jwt.MapClaims{
					identityKey: v.Role,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			return &User{
				Role: claims[identityKey].(string),
			}
		},
		// must have ADMIN role from the roles assigned to the current user
		Authorizator: func(data interface{}, c *gin.Context) bool {
			var rt RoleType
			if v, ok := data.(*User); ok && v.Role == rt.ToString(ADMIN) {
				return true
			}
			return false

		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		},
		// TokenLookup is a string in the form of "<source>:<name>" that is used
		// to extract token from the request.
		// Optional. Default value "header:Authorization".
		// Possible values:
		// - "header:<name>"
		// - "query:<name>"
		// - "cookie:<name>"
		TokenLookup: "header:Authorization",
		// TokenLookup: "query:token",
		// TokenLookup: "cookie:token",

		// TokenHeadName is a string in the header. Default value is "Bearer"
		TokenHeadName: "Bearer",

		// TimeFunc provides the current time. You can override it to use another time value. This is useful for testing or if your server uses a different time zone than your tokens.
		TimeFunc: time.Now,
	})
	return a
}

// add a user named 'bfpdadmin' with ADMIN  role
func (u *User) BootstrapUsers(defaultuser *string, d ds.DataSource) error {
	var (
		user User
		rt   RoleType
		dt   fdc.DocType
		err  error
	)
	userinfo := strings.Split(*defaultuser, ":")
	if userinfo[0] == "" {
		return errors.New("user is required")
	}
	if len(userinfo) == 1 || userinfo[1] == "" {
		return errors.New("password is required")
	}
	user.Name = userinfo[0]
	user.Password, err = HashPassword(userinfo[1])
	if err != nil {
		log.Println(err)
	} else {
		user.Role = rt.ToString(ADMIN)
		user.ID = fmt.Sprintf("%s:%s", dt.ToString(fdc.USER), user.Name)
		user.Type = dt.ToString(fdc.USER)
		err = d.Update(user.ID, user)
		if err != nil {
			log.Println(err)
		}
	}
	return err
}

// generates an encrypted password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// compares a plain text password with a hash and returns true for matches
// otherwise false
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
func findUser(name string, dc ds.DataSource) (User, bool) {
	var (
		u  User
		dt fdc.DocType
	)
	rc := true
	id := fmt.Sprintf("%s:%s", dt.ToString(fdc.USER), name)
	if err := dc.Get(id, &u); err != nil {
		log.Println(err)
		rc = false
	}
	return u, rc
}
