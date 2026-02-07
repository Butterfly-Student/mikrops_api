package model

import (
	"time"

	"mikrops/utils"
)

const (
	UpsertClientMessage      = "client.upsert"
	UpsertClientWorkflowName = "UpsertClientWorkflow"
)

type Client struct {
	ID int `json:"id" gorm:"primaryKey;autoIncrement"`
	ClientInput
}

func (Client) TableName() string {
	return "clients"
}

type ClientInput struct {
	Name      string    `json:"name" gorm:"column:name;type:varchar(100)"`
	BearerKey string    `json:"bearer_key,omitempty" gorm:"column:bearer_key;type:varchar(255);uniqueIndex"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

type ClientFilter struct {
	IDs        []int    `json:"ids"`
	Names      []string `json:"names"`
	BearerKeys []string `json:"bearer_keys"`
}

func ClientPrepare(v *ClientInput) {
	v.CreatedAt = time.Now()
	v.UpdatedAt = time.Now()
	if v.BearerKey == "" {
		v.BearerKey = utils.GenerateSecureToken(25)
	}
}

func (c ClientFilter) IsEmpty() bool {
	return len(c.IDs) == 0 && len(c.Names) == 0 && len(c.BearerKeys) == 0
}
