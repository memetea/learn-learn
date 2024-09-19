//go:build !swagger

package main

import (
	"log"

	"github.com/gorilla/mux"
)

// enableSwagger 禁用 Swagger 文档
func enableSwagger(_ *mux.Router, _ string) {
	log.Println("Swagger documentation is disabled")
}
