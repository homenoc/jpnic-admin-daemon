package api

import (
	"github.com/gin-gonic/gin"
	"github.com/homenoc/jpnic-admin-daemon/pkg/api/core/tool/config"
	"log"
	"net/http"
	"strconv"
)

func RestAPI() {
	router := gin.Default()
	router.Use(cors)

	api := router.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			//
			// Assignment
			//
			v1.POST("/assignment", notice.AddByAdmin)
			v1.DELETE("/assignment/:id", notice.DeleteByAdmin)
			v1.GET("/assignment", notice.GetAllByAdmin)
			v1.GET("/assignment/:id", notice.GetByAdmin)
		}
	}

	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(config.Conf.Port), router))
}

func cors(c *gin.Context) {

	//c.Header("Access-Control-Allow-Headers", "Accept, Content-ID, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Access-Control-Request-Headers, Access-Control-Request-Method, Connection, Host, Origin, User-Agent, Referer, Cache-Control, X-header")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "*")
	c.Header("Access-Control-Allow-Headers", "*")
	c.Header("Content-ID", "application/json")
	c.Header("Access-Control-Allow-Credentials", "true")
	//c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

	if c.Request.Method != "OPTIONS" {
		c.Next()
	} else {
		c.AbortWithStatus(http.StatusOK)
	}
}
