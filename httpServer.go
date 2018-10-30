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
	"flag"
	"net"
	"os/signal"
	"time"
	"syscall"
)


var currentPid int

func getMoviePaths(paths []string) []string {
	var files []string
	count := 0
	for _, pth := range paths {
		// check whether all path are empty
		if strings.TrimSpace(pth) == "" {
			count++
			//fmt.Printf("count: %d, len: %d\n", count, len(paths))
			if count == len(paths) {
				fmt.Println("ERROR: Each path item are empty in config file.")
				log.Fatal("Each path item are empty in config file.")
			}
			continue
		}

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
					strings.HasSuffix(f.Name(), ".mp4") ||
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

// check if the port is used
func PortUsed() bool {
	var isUsed bool = false

	ln, err := net.Listen("tcp", Port)
	if err != nil {
		//fmt.Println(err.Error())
		if strings.Contains(err.Error(), "Only one usage of each socket address") || strings.Contains(err.Error(), "address already in use") 	{
			isUsed = true
		} else {
			log.Fatal(err.Error())
			//os.Exit(-1)
		}
	} else {
		ln.Close()
	}

	return isUsed
}

// handle customized signal
//func signalHandler(sig string, channel chan string) {
//	signal.Notify()
//}

func CreatePidFile()  {
	currentPid = os.Getpid()
	_, err := os.Create(strconv.Itoa(currentPid) + ".pid.vdst")
	if err != nil {
		log.Println(err.Error())
	}
}


func RemovePidFile(filename string) {
	f, err := os.Open(filename)
	if err != nil {
		log.Println(err.Error())
	} else {
		f.Close()
		err := os.Remove(filename)
		if err != nil {
			log.Println(err.Error())
			fmt.Println("ERROR: Remove .pid file failed! Please remove it manually!")
		}
	}
}

// get existed instance pid
func GetExistedPid() int {
	dir, err := ioutil.ReadDir(".")
	if err != nil {
		log.Println(err.Error())
	}

	found := false
	var pid int
	for _, f := range dir {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".pid.vdst") {
			found = true

			strs := strings.Split(f.Name(), ".")
			pid, err = strconv.Atoi(strs[0])
			if err != nil {
				log.Println(err.Error())
			}
			break
		}
	}
	// if pid file is not found, find it manually.
	if PortUsed() {
		if !found {
			log.Println("Not found .pid.vdst file. Please find the pid and create a new one manually!")
		}
	}

	return pid
}

func SendSpecificSignal(sig os.Signal) {
	pr, err := os.FindProcess(GetExistedPid())
	if err != nil {
		log.Println(err.Error())
	}
	err = pr.Signal(sig)
	if err != nil {
		log.Println(err.Error())
	}
}

func main() {
	//sch := make(chan string, 1)
	var sOpt string
	flag.StringVar(&sOpt, "s", "reload", "send signal to the master process: stop, quit, exit, reload")
	flag.Parse()

	sOpt = strings.ToLower(sOpt)
	if sOpt == "exit" || sOpt == "quit" || sOpt == "stop" {
		sOpt = "exit"
	}

	sch := make(chan os.Signal, 1)
	errch := make(chan error, 1)


	signal.Notify(sch)   // redirect all signals to sch channel

	go func() {
		for {
			time.Sleep(10 * time.Second)
			sch <- syscall.SIGALRM
		}
	}()

	// if port is not used, do the initialization work,
	// else, do the stop, quit, reload work
	switch PortUsed() {
	case false :
		fmt.Printf("pid: '%d'\n", os.Getpid())
		log.Printf("pid: '%d'\n", os.Getpid())

		LogModuleInit()
		con := ConfigInit()
		info := generateEntryPage(con)

		// remove pid file created on last time
		RemovePidFile( strconv.Itoa(GetExistedPid()) + ".pid.vdst" )

		serverIPs := GetLocalIpAddrs()
		var ipAddv4 string    // local ip v4 address

		if serverIPs["wireless_ipv4"] != "" {
			ipAddv4 = serverIPs["wireless_ipv4"]
		} else {
			ipAddv4 = serverIPs["eth_ipv4"]
		}

		if ipAddv4 == "" {
			log.Fatal("Cannot find local ip address!")
		}

		fmt.Printf("server ip: %s\n", ipAddv4)

		// create a .pid file to save pid, which will be used by signal handler.
		CreatePidFile()

		var hypers []string
		for k, v := range info {
			log.Printf("%d -> %s\n", k, v)
			str := "<a href=\"http://" + ipAddv4 + Port + "/" + strconv.Itoa(k) + "\">" + path.Base(v) + "</a>"
			hypers = append(hypers, str)
			http.HandleFunc("/" + strconv.Itoa(k), setLinkSimple(k, v))
		}

		http.HandleFunc("/", setEntryPage(hypers))

		fmt.Printf("Start to Listen on '%s%s'\n", ipAddv4, Port)

		go func() {
			errch <- http.ListenAndServe(Port, nil)
		}()

	case true :
		fmt.Println("port is used!")
		switch sOpt {
		case "exit" :
			RemovePidFile(strconv.Itoa(currentPid) + "pid.vdst")
			pid2 := GetExistedPid()
			log.Printf("stop option received! Exist process %d!", pid2)
			SendSpecificSignal(syscall.SIGQUIT)
			goto end

		case "reload" :
			log.Println("Get reload signal.")
			//pid1 := GetExistedPid()
			log.Printf("reload config file '%s'\n", configFileName)
			SendSpecificSignal(syscall.SIGHUP)
			goto end
		default:
			log.Printf("Unkown option: '%s'!", sOpt)
		}
	}

	for {
		select {
		case err := <-errch:
			log.Fatal(err)

		case sig := <-sch:
			if sig == syscall.SIGHUP {
				fmt.Printf("INFO: signal '%v' from channel.\n", sig)

			} else if sig == syscall.SIGTERM || sig == syscall.SIGKILL || sig == syscall.SIGQUIT || sig == syscall.SIGINT {
				fmt.Printf("INFO: signal '%v' got. Exit the process!\n", sig)
				RemovePidFile(strconv.Itoa(currentPid) + "pid.vdst")
				goto end
			} else {
				fmt.Printf("WARN: signal '%v' received!\n", sig)
			}
		}
	}

	end:
	fmt.Println("INFO: Bye bye! :)")
	log.Println("INFO: Bye bye! :)")

}
