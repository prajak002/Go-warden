package main

import (
	bean "content-tagging/bean"
	"content-tagging/database"
	"content-tagging/utils"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
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

	API_KEY := os.Getenv("CHAT_GPT_API_KEY")

	if API_KEY == "" {
		log.Fatal("Failed to get API_KEY", err)
		return
	}

	intervalStr := os.Getenv("BATCH_INTERVAL_SECONDS")
	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		log.Fatal("Invalid BATCH_INTERVAL_SECONDS value", err)
		return
	}

	// Create a new Node with a Node number of 1
	node, err := snowflake.NewNode(1)
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		go beginBatch(gormDB, API_KEY, node)
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func beginBatch(db *gorm.DB, API_KEY string, node *snowflake.Node) {

	var config = bean.AppConfig{ConfigName: "GPT_LAST_RUN"}
	db.First(&config)
	fmt.Println(config.ConfigValue)
	lastRunTime, _ := utils.StringToDate(config.ConfigValue)
	currentTime := time.Now().UTC()
	fmt.Println(currentTime)
	var summaryArray []bean.SummarySearchByUser

	// Set the group_concat_max_len session variable to 15000
	if err := db.Exec("SET SESSION group_concat_max_len=45000").Error; err != nil {
		fmt.Println("failed to set group_concat_max_len", err)
		return
	}

	result := db.Raw(`SELECT arm.USER_ID, 
	GROUP_CONCAT(arc.CONTENT_FULL  SEPARATOR ',') SUMMARY_CONTENT,
	GROUP_CONCAT(arm.ARM_ROW_ID SEPARATOR ',') ROW_IDS  
	FROM AUDIT_REQUEST_MASTER arm, AUDIT_REQUEST_CONTENT arc 
	WHERE arm.ARM_ROW_ID =arc.ARM_ROW_ID 
	AND CREATED_DATETIME >=? AND CREATED_DATETIME <?
	GROUP BY arm.USER_ID 
	`, lastRunTime, currentTime).Scan(&summaryArray)
	if result.Error != nil {
		fmt.Println("Error in contents", result.Error)
		return
	}
	config.ConfigValue = utils.DateToString(currentTime)
	db.Model(&config).Where("CONFIG_NAME=?", config.ConfigName).Update("CONFIG_VALUE", config.ConfigValue)

	if result.RowsAffected == 0 {
		fmt.Println("No Records in contents", result.Error)
		return
	}
	for _, summary := range summaryArray {
		performInsert(db, summary, API_KEY, node)
	}

}

func performInsert(db *gorm.DB, summary bean.SummarySearchByUser, API_KEY string, node *snowflake.Node) {
	fmt.Println("Start Function")

	tags, err := utils.GetTagsFromChatGPT(summary.Summary, API_KEY)
	if err != nil {
		fmt.Println("Failed to get tags")
		return
	}

	jsonTagsArray, err := utils.MergeArraysFromJSON(tags)
	if err != nil {
		codeBlockPattern := regexp.MustCompile("(?s)```([^`]+)```")
		codeBlockMatches := codeBlockPattern.FindStringSubmatch(tags)

		if len(codeBlockMatches) >= 1 {
			jsonTagsArray, err = utils.MergeArraysFromJSON(codeBlockMatches[0])
			if err != nil {
				fmt.Println("Error parsing tags  none:", err)
				return
			}
		} else {
			fmt.Println("Error parsing tags, no code blocks: ", err)
			return

		}

	}

	if len(jsonTagsArray) > 0 {
		currentTime := time.Now().UTC()
		uniqueId := node.Generate().String()
		userSearch := bean.AuditUserSearchMaster{
			AusmID:    uniqueId,
			UserID:    summary.UserID,
			CreatedDt: currentTime,
		}
		result := db.Create(&userSearch)
		if result.Error != nil {
			fmt.Println("Error inserting data user search master:", result.Error)
			return
		} else {
			fmt.Println("Data inserted successfully user search master")
		}

		stringArray := strings.Split(summary.RowIds, ",")

		for _, id := range stringArray {
			idStore := bean.AuditUserSearchContentMapping{
				AusmID:    uniqueId,
				ArmRowID:  id,
				CreatedDt: currentTime,
			}
			result := db.Create(&idStore)
			if result.Error != nil {
				fmt.Println("Error inserting data into mapping:", result.Error)
				return
			} else {
				fmt.Println("Data inserted successfully mapping")
			}
		}

		for _, tag := range jsonTagsArray {
			tagRecord := bean.AuditUserSearchTags{
				RowID:     node.Generate().String(),
				AusmID:    uniqueId,
				TagName:   strings.ToLower(tag),
				CreatedDt: currentTime,
			}
			result := db.Create(&tagRecord)
			if result.Error != nil {
				fmt.Println("Error inserting data into tags:", result.Error)
				return
			} else {
				fmt.Println("Data inserted successfully to tags")
			}

		}
	}
	fmt.Println("End Function")
}
