//go:build swagger

package main

import (
	"fmt"
	"learn/docs" // Swaggo 生成的文档
	"log"
	"strings"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

// enableSwagger 启用 Swagger 文档
func enableSwagger(router *mux.Router, address string) {
	// 动态设置 Swagger host
	if strings.HasPrefix(address, ":") {
		docs.SwaggerInfo.Host = fmt.Sprintf("localhost%s", address)
	} else {
		docs.SwaggerInfo.Host = address
	}
	log.Println("Swagger documentation is enabled")
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
}
