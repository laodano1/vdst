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
)

const Port  = ":3000"

func handler4Root(res http.ResponseWriter, req *http.Request)  {
	//fmt.Fprint(res, "Hello world!")

	//file, err := os.Open("/Users/sophia/my_files/Media/[阳光电影www.ygdy8.com].破·局.HD.720p.国语中字.mkv")
	file, err := os.Open("/Users/sophia/my_files/Media/[阳光电影www.ygdy8.com].神偷奶爸3.HD.720p.中英双字幕.rmvb")
	if err != nil {
		log.Println(err.Error())
	}

	reader := bufio.NewReader(file)
	io.Copy(res, reader)

	//res.Write()
}

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
				        strings.HasSuffix(f.Name(), ".mp4") || 
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

		//reader := bufio.NewReader(file)
		//io.Copy(res, reader)
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

	contents, err := ioutil.ReadFile("/home/pi/Documents/goProject/dir_list.cf")
	if err != nil {
		log.Fatal(err.Error())
	}

	dirs := strings.Split(string(contents), "\n")

	//con := []string {"/Users/sophia/my_files/Media"}
	con := dirs
	info := generateEntryPage(con)
	serverIP := "192.168.1.99:3000"
	//serverIP := "192.168.1.182:3000"
	var hypers []string
	for k, v := range info {
		fmt.Printf("%d -> %s\n", k, v)
		str := "<a href=\"http://" + serverIP + "/" + strconv.Itoa(k) + "\">" + path.Base(v) + "</a>"
		hypers = append(hypers, str)
		http.HandleFunc("/" + strconv.Itoa(k), setLinkSimple(k, v))
		//http.HandleFunc("/" + strconv.Itoa(k), setLink(k, v))
	}

	//http.HandleFunc("/", handler4Root)
	http.HandleFunc("/", setEntryPage(hypers))
	log.Fatal(http.ListenAndServe(Port, nil))
}
