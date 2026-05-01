package main

import "guru/backend/products/internal/container"

// @title			Products API
// @version		1.0
// @description	REST API for products management with Kafka events
// @host			localhost:8080
// @BasePath		/api/v1
func main() {
	container.Build().Run()
}
