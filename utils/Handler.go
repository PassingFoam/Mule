package utils

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

var (
	src = rand.NewSource(time.Now().UnixNano())
)

const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

type PathDict struct {
	PathInfo
	Tag string
}

type PathInfo struct {
	Path string `json:"path"`
	Hits int    `json:"hits"`
}

func RandStringBytesMaskImprSrcUnsafe(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}

func HandleTarget(target string) (string, error) {
	if !strings.HasPrefix(target, "http") {
		// check to see if a port was specified
		re := regexp.MustCompile(`^[^/]+:(\d+)`)
		match := re.FindStringSubmatch(target)

		if len(match) < 2 {
			// no port, default to http on 80
			target = fmt.Sprintf("http://%s", target)
		} else {
			port, err2 := strconv.Atoi(match[1])
			if err2 != nil || (port != 80 && port != 443) {
				return target, fmt.Errorf("url scheme not specified")
			} else if port == 80 {
				target = fmt.Sprintf("http://%s", target)
			} else {
				target = fmt.Sprintf("https://%s", target)
			}
		}
	}
	//if !strings.HasSuffix(target,"/"){
	//	target = target + "/"
	//}

	if strings.HasSuffix(target, "/") {
		target = target[:len(target)-1]
	}

	return target, nil
}

func HandleLocation(location string) (string, error) {
	u, err := url.Parse(location)
	if err != nil {
		return location, err
	}

	path := u.Path

	// path存在特殊情况的处理,在某些java的场景下会出现path中带有;jsessionid的情况,会让url比较出现问题
	if strings.Contains(u.Path, ";jsessionid") {
		index := strings.Index(u.Path, ";jsessionid")
		path = u.Path[:index]
	}

	handled := u.Host + path

	return handled, nil
}

func ReadDict(info []string, root string) []PathDict {
	/*
		用来读取目录字典的数据，转换成列表的形式
	*/

	var allJson []PathDict
	var err error

	for _, dictpath := range info {
		var eachJson []PathInfo

		tagname, pathext := GetNameSuffix(dictpath)

		if pathext == "" {
			dictpath = filepath.Join(root, "Data", "DefDict", dictpath+".json")
		} else if pathext == ".txt" {
			dictpath, err = TextToJsonOfFile(dictpath, tagname, root)
			if err != nil {
				fmt.Printf("can't convert %s to json\n", dictpath)
				continue
			}

		}
		bytes, err := ioutil.ReadFile(dictpath)
		if err != nil {
			println(dictpath + " open failed")
			//panic(dictPath + " open failed")
			continue
		}

		if err := json.Unmarshal(bytes, &eachJson); err != nil {
			println(" Unmarshal failed")
			continue
		}

		for _, y := range eachJson {
			mid := PathDict{
				PathInfo: y,
				Tag:      tagname,
			}
			allJson = append(allJson, mid)
		}
		fmt.Println("use dict: " + dictpath)

	}

	// 将每个json数据按照Hits进行排序
	sort.Slice(allJson,
		func(i, j int) bool {
			return allJson[i].Hits > allJson[j].Hits
		})

	return allJson

}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func GetNameSuffix(filename string) (string, string) {
	filenameWithSuffix := path.Base(filename)

	fileSuffix := path.Ext(filenameWithSuffix)

	filenameOnly := strings.TrimSuffix(filenameWithSuffix, fileSuffix)

	return filenameOnly, fileSuffix
}

func CustomMarshal(message interface{}) (string, error) {
	/*
		自定义序列化函数，解决 "&"被转译的问题
	*/

	bf := bytes.NewBuffer([]byte{})

	jsonEncoder := json.NewEncoder(bf)
	jsonEncoder.SetEscapeHTML(false)
	jsonEncoder.SetIndent("", "    ")

	if err := jsonEncoder.Encode(message); err != nil {
		return "", err
	}

	return bf.String(), nil
}

func GetDefaultList(def string) (DefList []string) {
	if def != "" {
		if strings.Contains(def, ",") {
			userslice := strings.Split(def, ",")
			for _, user := range userslice {
				DefList = append(DefList, user)
			}
		} else {
			DefList = append(DefList, def)
		}
	}
	return DefList
}

func ReadTarget(targetfile string) (TargetList []string, err error) {
	file, err := os.Open(targetfile)
	if err != nil {
		panic("please check your file")
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		user := strings.TrimSpace(scanner.Text())
		if user != "" {
			TargetList = append(TargetList, user)
		}
	}
	return TargetList, err
}
