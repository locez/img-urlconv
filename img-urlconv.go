package main

import (
	"errors"
	"flag"
	"fmt"
	"gopkg.in/russross/blackfriday.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var (
	failog    *log.Logger
	BEGIN     int
	END       int
	PATH      string
	SINGAL    chan string
	waitGroup sync.WaitGroup
)

func init() {
	flag.IntVar(&BEGIN, "b", -1, "the begbin ID of article")
	flag.IntVar(&END, "e", -1, "the end ID of article")
	flag.StringVar(&PATH, "f", "", "the published dir")
}

func main() {
	flag.Parse()
	SINGAL = make(chan string, 10)
	logFile, err := os.OpenFile("failurls.log", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()
	failog = log.New(logFile, "", log.Ldate|log.Ltime)
	waitGroup = sync.WaitGroup{}
	start()
}

func getImgUrl(html []byte) []string {
	regstr := `<img .*?src="(\w+://.+?)"`
	re, _ := regexp.Compile(regstr)
	results := re.FindAllStringSubmatch(string(html), -1)
	urls := make([]string, 0)
	for i := 0; i < len(results); i++ {
		urls = append(urls, results[i][1])
	}
	return urls
}

func getArticle(url string) ([]byte, []byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, nil, err
	}
	html, _ := ioutil.ReadAll(resp.Body)
	headreg, _ := regexp.Compile(`(.|\n)+?class="clear" />(.|\n)+?<div class="d">`)
	tailreg, _ := regexp.Compile(`<script type="text/javascript">((.|\n)+)+?</script>`)
	viareg, _ := regexp.Compile(`>(英文原文|via|原文)(:){0,1}.*?<a .*?href="(\w+://.+?)".+?>(http\w+://.+)?</`)
	html = headreg.ReplaceAll(html, []byte(""))
	html = tailreg.ReplaceAll(html, []byte(""))
	if viareg.Match(html) {
		via := viareg.FindSubmatch(html)[3]
		return html, via, nil
	}
	return nil, nil, errors.New("No via found")
}

func findFile(via []byte) (string, error) {
	cmd := exec.Command("/bin/sh", "-c", `ack -l "(via\s{0,1}:\s{0,1}`+
		string(via)+`|^原文.{0,2} `+string(via)+`)" `+PATH)
	path, err := cmd.Output()

	//fmt.Printf("%v \n", string(path))
	if err != nil {
		return "", errors.New("No file found")
	}
	file := strings.TrimSuffix(string(path), "\n")
	if file == "" {
		return "", errors.New("No file found")
	}
	//	fmt.Printf("%v\n", file)
	return file, nil
}
func insertTitleImg(article string, titleImg string) string {
	lines := strings.Split(article, "\n")
	for i := 0; i < 10; i++ {
		if strings.Contains(lines[i], "#") || strings.Contains(lines[i], "===") {
			lines[i] = lines[i] + "\n\n![](" + titleImg + ")"
			break
		}
	}
	return strings.Join(lines, "\n")
}

func deal(url string) {
	defer func() {
		<-SINGAL
		waitGroup.Done()
	}()
	lcHtml, via, err := getArticle(url)
	if err != nil {
		log.Printf("%v : %v", err, url)
		return
	}
	lcImgs := getImgUrl(lcHtml)

	file, err := findFile(via)
	if err != nil {
		log.Printf("\n%v \n    url:%v \n    via:%v", err, url, string(via))
		return
	}
	mdData, err := ioutil.ReadFile(file)
	if err != nil {
		failog.Printf("%v \n\n    url: %v\n    md: %v\n", err, url, file)
		return
	}
	mdstr := string(mdData)

	mdHtml := blackfriday.Run(mdData)
	mdImgs := getImgUrl(mdHtml)

	mdNum := len(mdImgs)
	lcNum := len(lcImgs)

	if mdNum == lcNum {
		for i := mdNum - 1; i >= 0; i-- {
			mdstr = strings.Replace(mdstr, mdImgs[i], lcImgs[lcNum-1], -1)
			lcNum--
		}
	} else if mdNum > 0 || lcNum > 0 {
		// 是否只含有题图
		if mdNum == 0 && lcNum == 1 {
			mdstr = insertTitleImg(mdstr, lcImgs[0])
		} else {

			bak := strings.TrimSuffix(file, ".md") + ".backup.md"
			err := ioutil.WriteFile(bak, mdData, 0644)

			if err != nil {
				log.Printf("error: backup fail : %v", err)
				failog.Printf("\n\n    url: %v\n    md: %v\n", url, file)
				return
			}
			//	fmt.Printf("---\nurl:%v\nnum:%v\nmd:%v\nnum:%v\n---", url,
			//		lcNum, file, mdNum)
			if mdNum > 0 && lcNum > 0 && mdNum < lcNum {
				for i := mdNum - 1; i >= 0; i-- {
					mdstr = strings.Replace(mdstr, mdImgs[i], lcImgs[lcNum-1], -1)
					lcNum--
				}
				mdstr = insertTitleImg(mdstr, lcImgs[0])
			} else {
				os.Remove(bak)
				failog.Printf("\n\n    url: %v\n    md: %v\n", url, file)
				return
			}
		}
	} else {
		failog.Printf("\n\n    url: %v\n    md: %v\n", url, file)
		return
	}
	err = ioutil.WriteFile(file, []byte(mdstr), 0644)
	if err != nil {
		log.Printf("error: write file fail : %v", err)
		failog.Printf("url: %v    md: %v", url, file)
		return
	}
}

func start() {
	if BEGIN < 0 || END < 0 || END < BEGIN {
		fmt.Printf("Illegal Article ID range\n")
		return
	}
	if PATH == "" {
		fmt.Printf("please provide a dir via -f\n")
		return
	}
	PATH = strings.Replace(PATH, "~", os.Getenv("HOME"), 1)
	waitGroup.Add(END - BEGIN + 1)
	for i := BEGIN; i <= END; i++ {
		current := strconv.FormatInt(int64(i), 10)
		go deal("https://linux.cn/article-" + current + "-1.html?pr")
		SINGAL <- "+1"
	}
	waitGroup.Wait()
}
