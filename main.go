package main

import (
	"ebupot-app/config"
	"ebupot-app/routes"
)

func main() {
	config.ConnectDatabase()

	r := routes.SetupRouter()

	r.Run(":8080")
}
