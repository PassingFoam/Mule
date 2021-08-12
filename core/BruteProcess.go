package core

import (
	"Mule/utils"
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/panjf2000/ants"
	"github.com/projectdiscovery/cdncheck"
	"go.uber.org/zap"
	"net"
	"regexp"
	"strings"
	"sync"
	"time"
)

var PathLength int
var Countchan = make(chan struct{}, 10000)

var CurCancel context.CancelFunc
var CurContext context.Context
var CheckChan = make(chan int, 10000)
var Block int

type PoolPara struct {
	ctx      context.Context
	wordchan chan utils.PathDict
	custom   *CustomClient
	target   string
	wgs      *sync.WaitGroup
	wdmap    map[string]*WildCard
}

func ScanPrepare(ctx context.Context, client *CustomClient, target string) (*ReqRes, error) {

	_, err := client.RunRequest(ctx, target)

	if err != nil {
		return nil, fmt.Errorf("cann't connect to %s", target)
	}

	RandomPath = utils.RandStringBytesMaskImprSrcUnsafe(12)

	// TODO 暂时是不以/结尾所以在随机资源这里加了一个斜杠
	wildcard, err := client.RunRequest(ctx, target+"/"+RandomPath)

	if err != nil {
		return nil, err
	}

	return wildcard, nil

}

func ScanPrepare2(ctx context.Context, client *CustomClient, target string, root string) (map[string]*WildCard, error) {

	var WdMap map[string]*WildCard

	_, err := client.RunRequest(ctx, target)

	if err != nil {
		return nil, fmt.Errorf("cann't connect to %s", target)
	}

	RandomPath = utils.RandStringBytesMaskImprSrcUnsafe(12)

	//wildcard, err := client.RunRequest(ctx, target+"/"+RandomPath)

	WdMap, err = GenWildCardMap(ctx, client, RandomPath, target, root)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return WdMap, nil

}

func ScanTask(ctx context.Context, Opts Options, client *CustomClient) error {

	taskroot, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, curtarget := range Opts.Target {
		CheckFlag = 0

		//t1 := time.Now()
		CurContext, CurCancel = context.WithCancel(taskroot)

		// 做访问前准备，判断是否可以连通，以及不存在路径的返回情况

		wildcardmap, err := ScanPrepare2(ctx, client, curtarget, Opts.DirRoot)

		if err != nil {
			fmt.Println(err)
			continue
		}

		// 加入cdn检测
		r, _ := regexp.Compile("((2(5[0-5]|[0-4]\\d))|[0-1]?\\d{1,2})(\\.((2(5[0-5]|[0-4]\\d))|[0-1]?\\d{1,2})){3}")
		if len(r.Find([]byte(curtarget))) != 0 {
			ipv4 := string(r.Find([]byte(curtarget)))
			client, err := cdncheck.NewWithCache()
			if err == nil {
				if found, err := client.Check(net.ParseIP("173.245.48.12")); found && err == nil {
					fmt.Printf("%v is a part of cdn, so pass", ipv4)
					continue
				}
			}

		}

		//// 完成对不存在页面的处理
		//
		//wd, err := HandleWildCard(wildcard)

		//读取字典返回管道
		WordChan := MakeWordChan(Opts.Dictionary, Opts.DirRoot)

		go TimingCheck(CurContext, client, curtarget, wildcardmap["default"], CheckChan, CurCancel)

		go BruteProcessBar(CurContext, PathLength, curtarget, Countchan)

		//  开启线程池
		ScanPool, _ := ants.NewPoolWithFunc(Opts.Thread, func(Para interface{}) {
			CuPara := Para.(PoolPara)
			AccessWork(&CuPara)
		}, ants.WithExpiryDuration(2*time.Second))

		var wgs sync.WaitGroup

		PrePara := PoolPara{
			ctx:      CurContext,
			wordchan: WordChan,
			custom:   client,
			target:   curtarget,
			wgs:      &wgs,
			wdmap:    wildcardmap,
		}

		go ResHandle(ResChan)

		for i := 0; i < Opts.Thread; i++ {
			wgs.Add(1)
			_ = ScanPool.Invoke(PrePara)
		}

		wgs.Wait()

		//elapsed := time.Since(t1)
		//fmt.Println("App elapsed: ", elapsed)

		// TODO 根据hits更新json
		//for _, i := range ResSlice{
		//	fmt.Println(i.Path)
		//}
		time.Sleep(500 * time.Millisecond)

		UpdateDict(Opts.Dictionary, Opts.DirRoot)
		CurCancel()
	}

	return nil
}

//func MakeWordChan(DicPath string) chan string {
//	file, err := os.Open(DicPath)
//
//	WordChan := make(chan string)
//
//	if err != nil {
//		panic("please check your dictionary")
//	}
//
//	buf := bufio.NewReader(file)
//
//	go func() {
//		for {
//			line, _, err := buf.ReadLine()
//			if err == io.EOF {
//				break
//			}
//			WordChan <- string(line)
//
//		}
//		file.Close()
//		close(WordChan)
//	}()
//
//	return WordChan
//}

func MakeWordChan(DicSlice []string, DirRoot string) chan utils.PathDict {
	WordChan := make(chan utils.PathDict)

	alljson := utils.ReadDict(DicSlice, DirRoot)

	if len(alljson) == 0 {
		panic("please check your dict")
	}

	PathLength = len(alljson)

	go func() {
		for _, info := range alljson {
			WordChan <- info
		}

		close(WordChan)
	}()

	return WordChan
}

func f(x string) string {
	return x
}

func AccessWork(WorkPara *PoolPara) {
	defer WorkPara.wgs.Done()
	//result,err := custom.RunRequest(ctx, Url)

	for {
		select {
		case <-WorkPara.ctx.Done():
			return

		case word, ok := <-WorkPara.wordchan:
			if !ok {
				return
			}

			Countchan <- struct{}{}
			CheckChan <- 1

			path := word.Path

			PreHandleWord := strings.TrimSpace(path)
			if strings.HasPrefix(PreHandleWord, "#") || len(PreHandleWord) == 0 {
				break
			}

			if !strings.HasPrefix(PreHandleWord, "/") {
				PreHandleWord = "/" + PreHandleWord
			}

			result, err := WorkPara.custom.RunRequest(WorkPara.ctx, WorkPara.target+PreHandleWord)

			if err != nil {
				// TODO 错误处理
				continue
			}

			// 和资源不存在页面进行比较
			comres, err := CustomCompare(WorkPara.wdmap, PreHandleWord, result)
			//comres, err := CompareWildCard(WorkPara.wdmap["default"], result)

			if comres {
				ProBar.Clear()
				blue := color.New(color.FgBlue).SprintFunc()
				cy := color.New(color.FgCyan).SprintFunc()
				red := color.New(color.FgHiMagenta).SprintFunc()
				fmt.Printf("Path: %s \t Code:%s \t Length:%s\n", blue(WorkPara.target+PreHandleWord), cy(result.StatusCode), red(result.Length))
				word.Hits += 1
				Logger.Info("Success",
					zap.String("Path", WorkPara.target+PreHandleWord),
					zap.Int("Code", result.StatusCode),
					zap.Int64("Length", result.Length))
				ResChan <- word

			}

		}
	}

}
