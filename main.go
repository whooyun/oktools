package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

//go:embed templates
var tmpl embed.FS

//go:embed static
var assets embed.FS

func main() {
	gin.SetMode(config.App.Mode)

	if gin.Mode() == gin.ReleaseMode {
		gin.DisableConsoleColor()

		logfile := config.App.LogFile
		if logfile == "" {
			log.Fatalln("Please set the log file path!")
		}

		file, err := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
		if err != nil {
			file, err = os.Create(logfile)
			if file == nil {
				log.Fatalln(err)
			}
		}

		defer func() {
			err := file.Close()
			if err != nil {
				log.Fatalln(err)
			}
		}()

		gin.DefaultWriter = file
		log.SetOutput(file)
	}

	r := initRouter()

	var err error

	if gin.Mode() == gin.ReleaseMode {
		runNoTLS()

		err = r.RunTLS(":"+config.Http.Port, config.Http.SSL.Crt, config.Http.SSL.Key)
	} else {
		err = r.Run(":" + config.Http.Port)
	}

	if err != nil {
		log.Fatalln("Something terrible happened:", err)
	}
}

func runNoTLS() {
	go func() {
		e := gin.Default()
		e.SetHTMLTemplate(template.Must(template.New("").ParseFS(tmpl, "templates/*.html")))

		e.GET("/*path", func(c *gin.Context) {
			uri := c.Request.RequestURI
			if "/websocket" == uri {
				WebSocket(c)
			} else {
				c.Redirect(http.StatusMovedPermanently, "https://t.fuzhicode.com"+uri)
			}
		})

		err := e.Run(":80")
		if err != nil {
			log.Fatalln("Something terrible happened:", err)
		}
	}()
}

func initRouter() *gin.Engine {
	r := gin.Default()

	r.Use(gzip.Gzip(gzip.DefaultCompression))

	r.SetHTMLTemplate(template.Must(template.New("").ParseFS(tmpl, "templates/*.html")))
	r.Any("/static/*filepath", func(c *gin.Context) {
		c.Header("Cache-Control", "max-age=3153600")
		staticServer := http.FileServer(http.FS(assets))
		staticServer.ServeHTTP(c.Writer, c.Request)
	})

	r.GET("/favicon.ico", func(c *gin.Context) {
		c.Header("Cache-Control", "max-age=3153600")
		c.File("./static/favicon.ico")
	})

	// cache 30 minutes
	cacheEtag := Cache(30 * time.Minute)

	g := r.Group("/", cacheEtag)

	g.GET("/", Index)
	g.GET("/base64", Base64)
	g.GET("/image2base64", Image2Base64)
	g.GET("/tinyimg", TinyImage)
	g.GET("/hash", Hash)
	g.GET("/file-hash", FileHash)
	g.GET("/json", JSONView)
	g.GET("/number", Number)
	g.GET("/qrcode", QRCode)
	g.GET("/regex", Regex)
	g.GET("/timestamp", Timestamp)
	g.GET("/color", Color)
	g.GET("/aes", AES)
	g.GET("/des", DES)
	g.GET("/rsa", RSA)
	g.GET("/morse", Morse)
	g.GET("/url", URL)
	g.GET("/unicode", Unicode)
	g.GET("/json2go", JSON2GO)
	g.GET("/json2xml", JSON2XML)
	g.GET("/json2yaml", JSON2YAML)
	g.GET("/pdf2img", PDF2IMG)
	g.GET("/websocket", WebSocket)
	return r
}

func Cache(duration time.Duration) func(*gin.Context) {
	lastModify := time.Now()
	lastModifyFormat := lastModify.Format(http.TimeFormat)
	return func(c *gin.Context) {
		c.Header("Cache-Control", fmt.Sprintf("max-age=%d", int64(duration.Seconds())))
		c.Header("Last-Modified", lastModifyFormat)

		if c.Request.Method == "GET" && c.GetHeader("If-Modified-Since") == lastModifyFormat {
			c.AbortWithStatus(http.StatusNotModified)
			return
		}
		c.Next()
	}
}
