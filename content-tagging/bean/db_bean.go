package bean

import (
	"time"
)

type FakeDataRequestMaster struct {
	RowId   int    `gorm:"column:ROW_ID"`
	Content string `gorm:"column:CONTENT"`
}

func (FakeDataRequestMaster) TableName() string {
	return "FAKE_DATA_REQUEST_MASTER"
}

type FakeDomainNameMaster struct {
	DomainName string `gorm:"column:DOMAIN_NAME"`
}

func (FakeDomainNameMaster) TableName() string {
	return "FAKE_DOMAIN_NAME_MASTER"
}

type FakeUsernameMaster struct {
	Username string `gorm:"column:USERNAME"`
}

func (FakeUsernameMaster) TableName() string {
	return "FAKE_USERNAME_MASTER"
}

type AuditRequestMaster struct {
	RowId        string    `gorm:"column:ARM_ROW_ID"`
	UserId       string    `gorm:"column:USER_ID"`
	UserIpAddr   string    `gorm:"column:USER_IP_ADDR"`
	UserAction   string    `gorm:"column:USER_ACTION"`
	ContentShort string    `gorm:"column:CONTENT_SHORT"`
	UrlDomain    string    `gorm:"column:URL_DOMAIN"`
	UrlPath      string    `gorm:"column:URL_PATH"`
	CreatedDt    time.Time `gorm:"column:CREATED_DATETIME"`
	Timezone     string    `gorm:"column:TIME_ZONE"`
}

func (AuditRequestMaster) TableName() string {
	return "AUDIT_REQUEST_MASTER"
}

type AuditRequestContent struct {
	RowId       string `gorm:"column:ARM_ROW_ID"`
	ContentFull string `gorm:"column:CONTENT_FULL"`
}

func (AuditRequestContent) TableName() string {
	return "AUDIT_REQUEST_CONTENT"
}

type AppConfig struct {
	ConfigName  string `gorm:"column:CONFIG_NAME"`
	ConfigValue string `gorm:"column:CONFIG_VALUE"`
}

func (AppConfig) TableName() string {
	return "APP_CONFIG"
}

type SummarySearchByUser struct {
	UserID  string `gorm:"column:USER_ID"`
	Summary string `gorm:"column:SUMMARY_CONTENT"`
	RowIds  string `gorm:"column:ROW_IDS"`
}

type AuditUserSearchMaster struct {
	AusmID    string    `gorm:"column:AUSM_ID"`
	UserID    string    `gorm:"column:USER_ID"`
	CreatedDt time.Time `gorm:"column:CREATED_DT"`
}

func (AuditUserSearchMaster) TableName() string {
	return "AUDIT_USER_SEARCH_MASTER"
}

type AuditUserSearchTags struct {
	RowID     string    `gorm:"column:ROW_ID"`
	AusmID    string    `gorm:"column:AUSM_ID"`
	TagName   string    `gorm:"column:TAG_NAME"`
	CreatedDt time.Time `gorm:"column:CREATED_DT"`
}

func (AuditUserSearchTags) TableName() string {
	return "AUDIT_USER_SEARCH_TAGS"
}

type AuditUserSearchContentMapping struct {
	AusmID    string    `gorm:"column:AUSM_ID"`
	ArmRowID  string    `gorm:"column:ARM_ROW_ID"`
	CreatedDt time.Time `gorm:"column:CREATED_DT"`
}

func (AuditUserSearchContentMapping) TableName() string {
	return "AUDIT_USER_SEARCH_CONTENT_MAPPING"
}
