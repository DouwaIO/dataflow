package model

type Pipeline struct {
	ID string `json:"id"             gorm:"primary_key;type:varchar(50);column:pipeline_id"`
	Name string `json:"name"          gorm:"type:varchar(500);column:name"`
	Data string `json:"data"          gorm:"type:text;column:data"`
}
