package main

import (
	"fmt"
	"log"
	"simulation/bean"
	"simulation/database"
	util "simulation/utils"
	"time"

	"github.com/bwmarrin/snowflake"
	"gorm.io/gorm"
)

func main() {

	gormDB, err := database.InitializeDB()
	if err != nil {
		log.Fatal("failed to connect to the database", err)
		return
	}

	actionArray := []string{"C", "X", "P", "I"}

	