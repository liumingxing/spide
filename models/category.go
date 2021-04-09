package models

import (
  gorm "github.com/jinzhu/gorm"
)
type Categroy struct {
  gorm.Model
  ParentID    uint
  Name        string
}

func (Categroy) TableName() string{
  return "categories"
}