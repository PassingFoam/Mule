package core

import (
	"Mule/utils"
	"context"
	"fmt"
	"github.com/antlabs/strsim"
	"path/filepath"
	"strings"
)

var RandomPath string
var BlackList []int

func Compare30x(WdLoc string, Res string) (bool, error) {
	// 改为url对比 会出现由于参数问题导致的location对比不合理
	ratio := 0.98

	HandleWd, err := utils.HandleLocation(WdLoc)

	if err != nil {
		HandleWd = WdLoc
	}

	HandleRes, err := utils.HandleLocation(Res)

	if err != nil {
		HandleRes = Res
	}

	ComRatio := strsim.Compare(HandleWd, HandleRes)

	if ratio > ComRatio {
		return true, nil
	}

	return false, nil

}

func Compare200(WdBody *[]byte, ResBody *[]byte) (bool, error) {
	ratio := 0.98

	ComRatio := strsim.Compare(string(*WdBody), string(*ResBody))

	if ratio > ComRatio {
		return true, nil
	}

	return false, nil
}

func CompareWildCard(wd *WildCard, result *ReqRes) (bool, error) {
	switch wd.Type {

	// 类型1即资源不存在页面状态码为200
	case 1:
		if result.StatusCode == 200 {
			comres, err := Compare200(&wd.Body, &result.Body)
			return comres, err
		} else if (result.StatusCode > 300 && result.StatusCode < 404) || result.StatusCode == 503 && !IntInSlice(result.StatusCode, BlackList) {
			return true, nil
		}
	// 类型2即资源不存在页面状态码为30x
	case 2:
		if result.StatusCode == 200 {
			return true, nil
		} else if result.StatusCode == wd.StatusCode {
			comres, err := Compare30x(wd.Location, result.Header.Get("Location"))
			return comres, err
		} else if (result.StatusCode > 300 && result.StatusCode < 404) || result.StatusCode == 503 && !IntInSlice(result.StatusCode, BlackList) {
			return true, nil
		}
		// 类型3 即资源不存在页面状态码404或者奇奇怪怪
	case 3:
		// TODO 存在nginx的类似301跳转后的404页面
		if wd.StatusCode != result.StatusCode {
			if result.StatusCode == 200 || (result.StatusCode > 300 && result.StatusCode < 404) && result.StatusCode == 503 || !IntInSlice(result.StatusCode, BlackList) {
				return true, nil
			}
		}

	}

	return false, nil

}

func CustomCompare(wdmap map[string]*WildCard, path string, result *ReqRes) (bool, error) {
	for key := range wdmap {
		if key == "default" {
			continue
		} else {
			if strings.Contains(path, key) {
				res, err := CompareWildCard(wdmap[key], result)
				return res, err
			}
		}
	}

	key := "default"
	res, err := CompareWildCard(wdmap[key], result)
	return res, err

}

func HandleWildCard(wildcard *ReqRes) (*WildCard, error) {

	if wildcard.StatusCode == 200 {
		wd := WildCard{
			StatusCode: wildcard.StatusCode,
			Body:       wildcard.Body,
			Length:     int64(len(wildcard.Body)),
			Type:       1,
		}
		return &wd, nil

	} else if wildcard.StatusCode > 300 && wildcard.StatusCode < 404 {
		wd := WildCard{
			StatusCode: wildcard.StatusCode,
			Location:   wildcard.Header.Get("Location"),
			Type:       2,
		}
		return &wd, nil
	} else {
		wd := WildCard{
			StatusCode: wildcard.StatusCode,
			Type:       3,
		}
		return &wd, nil
	}

}

func GenWildCardMap(ctx context.Context, client *CustomClient, random string, target string, proroot string) (map[string]*WildCard, error) {
	var Testpath string
	var err error
	var wd *WildCard
	resmap := make(map[string]*WildCard)

	wdlist, err := GetExPathList(proroot)

	wdlist = append(wdlist, "")

	if err != nil {
		return nil, err
	}

	for _, ex := range wdlist {
		if ex == "" {
			Testpath = "/" + RandomPath
			wd, err = GenWd(ctx, client, target, Testpath)
			if err != nil {
				return nil, err
			}
			resmap["default"] = wd

		} else if !strings.Contains(ex, "$$") {
			fmt.Printf("%s don't have symbol $$, please check\n", ex)
			continue
		} else {
			Testpath = strings.Replace(ex, "$$", random, 1)
			wd, err = GenWd(ctx, client, target, Testpath)
			if err != nil {
				return nil, fmt.Errorf("When you test %s, there is something error\n", ex)
			}
			in := strings.Index(ex, "$$")
			key := ex[in+2:]
			resmap[key] = wd
		}
	}

	return resmap, nil

}

func GetExPathList(root string) ([]string, error) {
	// TODO 将加入参数dirroot,这里为了测试方便使用了绝对路径
	expath := filepath.Join(root, "Data", "SpecialList", "exwildcard.txt")

	ex, err := utils.ReadLines(expath)
	if err != nil {
		println(err.Error())
		return nil, err
	}

	return ex, nil
}

func GenWd(ctx context.Context, client *CustomClient, target string, Tpath string) (*WildCard, error) {
	wildcard, err := client.RunRequest(ctx, target+Tpath)

	if err != nil {
		return nil, err
	}

	wd, err := HandleWildCard(wildcard)

	return wd, nil
}
