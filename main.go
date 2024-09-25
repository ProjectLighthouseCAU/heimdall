package main

import (
	"log"

	"github.com/ProjectLighthouseCAU/heimdall/setup"
)

// @Title		Heimdall Lighthouse API
// @Version		0.1
// @Description	This is the REST API of Project Lighthouse that manages users, roles, registration keys, API tokens and everything about authentication and authorization.
// @Description NOTE: This API is an early alpha version that still needs a lot of testing (unit tests, end-to-end tests and security tests)
// @Host		https://lighthouse.uni-kiel.de
// @BasePath	/api
func main() {
	app := setup.Setup()
	log.Println("Setup done. Listening until Ragnarök...")
	log.Fatal(app.Listen(":8080"))
}
