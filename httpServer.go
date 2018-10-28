package main

import (
	"net/http"
	"log"
	"os"
	"bufio"
	"io"
	"path"
	"fmt"
	"io/ioutil"
	"strings"
	"strconv"
	"time"
)

const Port  = ":3000"

//func handler4Root(res http.ResponseWriter, req *http.Request)  {
//	//fmt.Fprint(res, "Hello world!")
//
//	if err != nil {
//		log.Println(err.Error())
//	}
//
//	reader := bufio.NewReader(file)
//	io.Copy(res, reader)
//
//	//res.Write()
//}

func getMoviePaths(paths []string) []string {
	var files []string
	for _, pth := range paths {
		fds, err := ioutil.ReadDir(pth)
		if err != nil {
			log.Println(err.Error())
		}

		for _, f := range fds {
			if f.IsDir() { // folder
				for _, pt := range getMoviePaths( []string{ path.Join(pth, f.Name()) } ) {
					files = append(files, pt)
				}
			} else {
				if strings.HasSuffix(f.Name(), ".rmvb") ||
					strings.HasSuffix(f.Name(), ".mkv") ||
					strings.HasSuffix(f.Name(), ".avi") {
					files = append(files, path.Join(pth, f.Name()))
				}
			}
		}
	}

	return files
}

func generateEntryPage(paths []string) map[int]string {
	info := make(map[int]string)
	for index, pth := range getMoviePaths(paths) {
		info[index] = pth
	}

	return info
}

func setLinkSimple(num int, vlink string) func( http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		log.Printf("--> '%s' is selected!\n", path.Base(vlink))
		file, err := os.Open(vlink)
		if err != nil {
			log.Println(err.Error())
		}
		reader := bufio.NewReader(file)
		io.Copy(res, reader)
	}
}

func setLink(k int, v string) func( http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		file, err := os.Open(v)
		if err != nil {
			log.Println(err.Error())
			http.NotFound(res, req)
			return
		}

		var start, end int64
		size := req.Header.Get("content-length")
		fmt.Println("content-length: " + size)
		fmt.Sscanf(req.Header.Get("Range"), "bytes=%d-%d", &start, &end)
		info, err := file.Stat()
		if err != nil {
			log.Println(err.Error())
			http.NotFound(res, req)
			return
		}

		fmt.Printf("file size: %d", info.Size())

		if start < 0 ||start >= info.Size() ||end < 0 || end >= info.Size(){
			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte(fmt.Sprintf("out of index, length:%d",info.Size())))
			return
		}

		if end == 0 {
			end = info.Size() - 1
		}

		res.Header().Add("Accept-ranges", "bytes")
		res.Header().Add("Content-Length", strconv.FormatInt(end-start+1, 10))
		res.Header().Add("Content-Range", "bytes "+strconv.FormatInt(start, 10)+"-"+strconv.FormatInt(end, 10)+"/"+strconv.FormatInt(info.Size()-start, 10))
		//res.Header().Add("Content-Disposition", "attachment; filename="+info.Name())
		_, err = file.Seek(start, 0)
		if err != nil {
			log.Println(err.Error())
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, err = io.CopyN(res, file, end-start+1)
		if err != nil {
			log.Println(err.Error())
			return
		}

	}
}

func setEntryPage(hypers []string) func(http.ResponseWriter, *http.Request) {
	html := ""
	for _, hyper := range hypers {
		html += "<p>" + hyper + "<button>拷贝链接</button></p>"
	}
	return func(res http.ResponseWriter, req *http.Request) {
		fmt.Fprint(res, html)
	}
}

func main() {
	// 检查端口是否被占用，如果占用，说明已经有instance了，那就只能reload config操作，
	// 如果，没有被占用，那就执行全部初始化操作，然后启动整个instance

	year, month, day := time.Now().Date()
	logFileName := fmt.Sprintf("Server-%d-%d-%d.log", year, month, day)
	file, err := os.Create(logFileName)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.SetOutput(file)
	contents, err := ioutil.ReadFile("dir_list.cf")
	if err != nil {
		log.Fatal(err.Error())
	}

	dirs := strings.Split(string(contents), "\n")

	con := dirs
	info := generateEntryPage(con)
	// 需要通过代码获取ip地址 ？？？
	serverIP := "192.168.1.99" + Port

	var hypers []string
	for k, v := range info {
		log.Printf("%d -> %s\n", k, v)
		str := "<a href=\"http://" + serverIP + "/" + strconv.Itoa(k) + "\">" + path.Base(v) + "</a>"
		hypers = append(hypers, str)
		http.HandleFunc("/" + strconv.Itoa(k), setLinkSimple(k, v))
	}

	http.HandleFunc("/", setEntryPage(hypers))

	fmt.Println("Start to Listen on " + Port)
	log.Fatal(http.ListenAndServe(Port, nil))
}
