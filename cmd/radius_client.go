package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic(fmt.Sprintf("Error loading envs"))
	}

	packet := radius.New(radius.CodeAccessRequest, []byte(getEnvStr("RADIUS_SECRET", "")))
	rfc2865.UserName_SetString(packet, "danila.kazliakouski@indev.by")

	response, err := radius.Exchange(context.Background(), packet, "127.0.0.1:1812")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Ответ:", response.Code)
}

func getEnvStr(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}
