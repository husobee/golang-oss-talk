// Package main - the main web application package namespace
package main

// import libraries needed
import (
	"fmt"
	"runtime"
	"strconv"

	"crypto/rand"
	"crypto/sha512"

	"github.com/gin-gonic/gin"
)

// main - Main entry point of the web server
func main() {
	// Create the gin
	api := gin.Default()
	// URL Routing
	// version 1 of the API, grouping
	v1 := api.Group("/v1")
	{
		// health check endpoint
		v1.GET("/hash/:number_hashes/times", HashHandler)
		v1.GET("/hash/:number_hashes/times/concurrently", HashConcurrentlyHandler)
	}

	api.Run(":8080")
}

// hashResponse - api response structure
type hashResponse struct {
	HashCount int64    `json:"count"`  // in json encoding it will be count
	Hashes    []string `json:"hashes"` // in json encoding it will be hashes
}

// randomHash - create a hash of some random
var randomHash = func() string {
	// new hash
	sha512.New()
	// instantiate data pointer
	data := make([]byte, 32)
	// read in some random data
	rand.Read(data)
	// hash sum that data and spit out hex
	hashsum := fmt.Sprintf("%x", sha512.Sum512(data))
	// return the hashsum
	return hashsum
}

// HashHandler - a gin handler for creating sequential hashes
func HashHandler(c *gin.Context) {
	// get the variable from the url
	number_hashes_requested, err := strconv.ParseInt(c.Params.ByName("number_hashes"), 10, 64)
	if err != nil {
		// send response
		c.JSON(400, &map[string]string{"status": "failed"})
	}
	// initialize response struct
	var response = new(hashResponse)
	response.HashCount = number_hashes_requested
	for i := 0; i < int(response.HashCount); i++ {
		response.Hashes = append(response.Hashes, randomHash())
	}
	// send response
	c.JSON(200, response)
}

// HashConcurrentlyHandler - a gin handler for creating concurrent hashes
func HashConcurrentlyHandler(c *gin.Context) {
	// get the variable from the url
	number_hashes_requested, err := strconv.ParseInt(c.Params.ByName("number_hashes"), 10, 64)
	if err != nil {
		// send response
		c.JSON(400, &map[string]string{"status": "failed"})
	}
	// initialize response struct
	var response = new(hashResponse)
	// create a hash_done_channel to pipe finished hashes back to handler
	var hash_done_channel = make(chan string, number_hashes_requested)
	response.HashCount = number_hashes_requested
	var processHashes = func(hash_count int) {
		for x := 0; x < hash_count; x++ {
			hash_done_channel <- randomHash()
		}
	}
	runtime.GOMAXPROCS(8)
	// left over hashes, not cleanly divisible
	left_over := int(number_hashes_requested) % 8
	go processHashes(left_over)
	number_hashes_requested -= int64(left_over)
	// spin up number of cpus goroutines
	for i := 0; i < 8; i++ {
		go processHashes(int(number_hashes_requested) / 8)
	}
	// append the finished hashes as they come in the channel
	for i := 0; i < int(response.HashCount); i++ {
		response.Hashes = append(response.Hashes, <-hash_done_channel)
	}
	// send response
	c.JSON(200, response)
}
