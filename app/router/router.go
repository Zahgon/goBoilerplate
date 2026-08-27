package router

import (
	"goBoilterplate/app/controllers"
	"goBoilterplate/app/helpers"
	"goBoilterplate/app/middlewares"
	_ "goBoilterplate/docs" // For Swagger

	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const apiPrefix = "/api"

// route records a registered pattern and the methods it answers. A route with
// anyMethod set answers every method, which is how the API prefix behaves: its
// group middleware has to run even for requests that match no handler inside
// it, so unmatched paths and methods below the prefix are still routed rather
// than falling through to the router defaults.
type route struct {
	pattern   string
	methods   map[string]bool
	anyMethod bool
}

// allowMethodOrder is the order methods are listed in an Allow header.
var allowMethodOrder = []string{
	http.MethodConnect, http.MethodDelete, http.MethodGet, http.MethodHead,
	http.MethodOptions, http.MethodPatch, http.MethodPost, "PROPFIND",
	http.MethodPut, http.MethodTrace, "REPORT",
}

func segments(p string) []string {
	return strings.Split(strings.TrimPrefix(p, "/"), "/")
}

// matches reports whether a request path is answered by a route pattern.
func (r route) matches(path string) bool {
	pattern := segments(r.pattern)
	request := segments(path)

	for i, seg := range pattern {
		if strings.HasPrefix(seg, "*") {
			return i < len(request)
		}
		if i >= len(request) {
			return false
		}
		if strings.HasPrefix(seg, ":") {
			if request[i] == "" {
				return false
			}
			continue
		}
		if seg != request[i] {
			return false
		}
	}

	return len(pattern) == len(request)
}

func buildAllow(methods map[string]bool) string {
	allow := http.MethodOptions
	for _, method := range allowMethodOrder {
		if methods[method] {
			allow += ", " + method
		}
	}
	return allow
}

// allowHeader advertises the methods a known path answers whenever the request
// method is not one of them. It runs ahead of every other middleware because
// the header has to be on the response even when a later middleware short
// circuits the request, as the CORS preflight reply does. The route table is
// resolved on first use because middleware has to be registered before the
// routes it is meant to wrap.
func allowHeader(app *gin.Engine) gin.HandlerFunc {
	var once sync.Once
	var routes []route

	return func(c *gin.Context) {
		once.Do(func() { routes = routeTable(app) })

		path := c.Request.URL.Path

		found := false
		methods := map[string]bool{}
		for _, r := range routes {
			if !r.matches(path) {
				continue
			}
			if r.anyMethod {
				c.Writer.Header().Del("Allow")
				return
			}
			found = true
			for method := range r.methods {
				methods[method] = true
			}
		}

		if !found || methods[c.Request.Method] {
			return
		}

		c.Writer.Header().Set("Allow", buildAllow(methods))
	}
}

func underAPIPrefix(path string) bool {
	return path == apiPrefix || strings.HasPrefix(path, apiPrefix+"/")
}

// apiFallback answers requests below the API prefix that matched no handler.
// The prefix middleware runs first, so an unauthenticated caller is rejected
// before the request is reported as unknown.
func apiFallback(jwt gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		jwt(c)
		if c.IsAborted() {
			return
		}
		helpers.ErrorJSON(c, http.StatusNotFound)
	}
}

func fallback(jwt gin.HandlerFunc, code int) gin.HandlerFunc {
	api := apiFallback(jwt)
	return func(c *gin.Context) {
		if underAPIPrefix(c.Request.URL.Path) {
			api(c)
			return
		}
		helpers.ErrorJSON(c, code)
	}
}

// Init Router
func Init(app *gin.Engine) {
	app.RedirectTrailingSlash = false
	app.RedirectFixedPath = false
	app.HandleMethodNotAllowed = true

	jwt := middlewares.Jwt()

	app.Use(allowHeader(app))
	app.Use(middlewares.Cors())
	app.Use(middlewares.Gzip())
	app.Use(middlewares.Logger())
	app.Use(middlewares.Secure())
	app.Use(middlewares.Recover())

	app.GET("/", controllers.Index)
	app.GET("/test", controllers.Test)
	app.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := app.Group(apiPrefix, jwt)
	{
		api.POST("/login", controllers.Login)
		api.GET("/logout", controllers.Logout)

		users := api.Group("/users")
		{
			users.GET("", controllers.UserList)
			users.POST("", controllers.UserStore)
			users.GET("/:id", controllers.UserShow)
			users.PUT("/:id", controllers.UserUpdate)
			users.DELETE("/:id", controllers.UserDelete)
		}
	}

	app.NoRoute(fallback(jwt, http.StatusNotFound))
	app.NoMethod(fallback(jwt, http.StatusMethodNotAllowed))

	log.Printf("Server started...")
}

func routeTable(app *gin.Engine) []route {
	table := []route{
		{pattern: apiPrefix, anyMethod: true},
		{pattern: apiPrefix + "/*any", anyMethod: true},
	}

	return append(table, engineRoutes(app)...)
}

func engineRoutes(app *gin.Engine) []route {
	byPattern := map[string]map[string]bool{}
	for _, info := range app.Routes() {
		if byPattern[info.Path] == nil {
			byPattern[info.Path] = map[string]bool{}
		}
		byPattern[info.Path][info.Method] = true
	}

	table := make([]route, 0, len(byPattern))
	for pattern, methods := range byPattern {
		table = append(table, route{pattern: pattern, methods: methods})
	}
	return table
}
