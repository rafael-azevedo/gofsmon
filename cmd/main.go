package main

import (
	"log"
	"os"

	"github.com/rafael-azevedo/gofsmon"
)

var config string

func init() {
	config = os.Getenv("GOFSMONCONF")
}

func main() {

	conf, err := gofsmon.ReadYamal(config)
	if err != nil {
		log.Fatalf("%s %s\n", "Could Not Read Config ::", err)
	}

	mc := gofsmon.MCleanService{}

	err = mc.NewTFS(conf)
	if err != nil {
		log.Fatal(err)
	}

	err = mc.CleanDir()
	if err != nil {
		log.Fatal(err)
	}

}
