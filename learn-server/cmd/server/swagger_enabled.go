//go:build swagger

package main

import (
	"fmt"
	"learn/docs" // Swaggo 生成的文档
	"log"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

// enableSwagger 启用 Swagger 文档
func enableSwagger(router *mux.Router, port string) {
	// 动态设置 Swagger host
	docs.SwaggerInfo.Host = fmt.Sprintf("localhost:%s", port)
	log.Println("Swagger documentation is enabled")
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
}
