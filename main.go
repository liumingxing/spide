package main

import (
	logs "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"spide/models"
	"spide/util"
	"strconv"
	"strings"
	"time"
)

func init() {
	// 设置日志格式为json格式
	logs.SetFormatter(&logs.JSONFormatter{})

	// 设置将日志输出到标准输出（默认的输出为stderr，标准错误）
	// 日志消息输出可以是任意的io.writer类型
	//logFile, err := os.OpenFile("download.log", os.O_RDWR | os.O_CREATE, 0777)
	//if err != nil {
	//	fmt.Printf("open file error=%s\r\n", err.Error())
	//	os.Exit(-1)
	//}

	//logs.SetOutput(logFile)
	logs.SetOutput(os.Stdout)

	// 设置日志级别为warn以上
	//logs.SetLevel(logs.WarnLevel)
	logs.SetLevel(logs.InfoLevel)
}
var channels = make(chan int, 20)
var fileDir = "/Volumes/mamashaiimages/yellow/"
func Download(tsfile models.Tsfile){
	var f *os.File
	arr := strings.Split(tsfile.Name, "/")
	name := arr[len(arr)-1]
	path := fileDir + strconv.Itoa(int(tsfile.MovieID)) + "/full/" + name
	f, _ = os.Create(path) //创建文件

	client := &http.Client{
		Timeout: 5*time.Second,
	}
	resp, err := client.Get(tsfile.Name)
	if err != nil{
		logs.WithFields(logs.Fields{
			"id": tsfile.ID,
		}).Warning(string(err.Error()))
		<- channels
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		logs.WithFields(logs.Fields{
			"id": tsfile.ID,
		}).Warning(string(err.Error()))
		<- channels
		return
	}
	f.Write(body)
	f.Close()


	tsfile.Finished = true
	tsfile.Filesize = len(body)
	util.DB.Unscoped().Save(&tsfile)
	id := <- channels
	logs.WithFields(logs.Fields{
		"id":    id,
	}).Info("download ts file finished")
}

//判断一个movie是否下载完毕，同时调整m3u8文件
func SetMovieFinished(){
	for true {
		var movies []models.Movie
		offset := 0
		for true {
			result := util.DB.Where("ts_downloaded = 0").Limit(1000).Offset(offset).Find(&movies)
			if result.RowsAffected > 0 {
				for _, movie := range movies {
					//设置新的m3u8文件
					path := fileDir + strconv.Itoa(int(movie.ID)) + "/full/index.m3u8"
					newPath := fileDir + strconv.Itoa(int(movie.ID)) + "/full/movie.m3u8"
					file, _ := os.Open(path)
					newFile, _ := os.Create(newPath)
					bytes, err := ioutil.ReadAll(file)
					if err == nil {
						re := regexp.MustCompile(`(?U)https://[/|\w|\.]+(\w+.ts)`)
						newString := re.ReplaceAllString(string(bytes), "$1")

						re = regexp.MustCompile(`(?U)(https://[/|\w|\.]+key.key)`)
						newString = re.ReplaceAllString(newString, "key.key")

						newFile.Write([]byte(newString))
						newFile.Close()
					}

					//设置是否下载完成
					var count int64
					util.DB.Model(&models.Tsfile{}).Where("movie_id = ? and finished = ?", movie.ID, false).Count(&count)
					if count == 0{
						//movie.TsDownloaded = true
						//util.DB.Save(&movie)
						util.DB.Model(&movie).Update("ts_downloaded", true)
					}

					logs.WithFields(logs.Fields{
						"movie id":    movie.ID,
					}).Info("check movie finish")
				}
			} else {
				break
			}
			offset += 1000
		}
	}

	time.Sleep(time.Second*30)
}

func DownloadAllTsFile(){
	var tsfiles  []models.Tsfile
	for true{
		result := util.DB.Unscoped().Limit(5000).Where("finished = 0").Order("id desc").Find(&tsfiles)
		if result.RowsAffected > 0{
			for _, tsfile := range tsfiles {
				channels <- tsfile.ID
				go Download(tsfile)
			}
		}
	}
	m <- 1
}

var m = make(chan int, 1)
func main() {
	go SetMovieFinished()

	go DownloadAllTsFile()
	<- m
}
