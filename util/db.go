package util

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"

)
var DB *gorm.DB
var err error
func init() {
  DB,err = gorm.Open("mysql","root:@(127.0.0.1:3306)/yellow?charset=utf8")
  if err != nil {
    return
  }
}