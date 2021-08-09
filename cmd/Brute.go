/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"Mule/core"
	"Mule/utils"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

// BruteCmd represents the Brute command
var BruteCmd = &cobra.Command{
	Use:   "Brute",
	Short: "a weak Directory Blasting",
	Long:  `I'm too lazy to write more introduction`,
	RunE:  StartBrute,
}

func init() {
	rootCmd.AddCommand(BruteCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// BruteCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// BruteCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	BruteCmd.Flags().StringP("url", "u", "", "brute target(currently only single url)")
	BruteCmd.Flags().StringP("urls", "U", "", "targets from file")
	BruteCmd.Flags().StringP("dic", "d", "", "dictionary to brute")
	BruteCmd.Flags().StringP("flag", "f", "", "use default dictionary in /Data")
	BruteCmd.Flags().StringP("output", "o", "./res.log", "output res default in ./res.log")
	BruteCmd.Flags().StringArrayP("Headers", "H", []string{}, "Request's Headers")
	BruteCmd.Flags().StringP("Cookie", "C", "", "Request's Cookie")
	BruteCmd.Flags().IntP("timeout", "", 5, "request's timeout")
	BruteCmd.Flags().IntP("Thread", "t", 30, "the size of thread pool")
	BruteCmd.Flags().IntP("block", "b", 4, "the number of auto stop brute")
	BruteCmd.Flags().IntSlice("blacklist", []int{}, "the black list of statuscode")

}

func StartBrute(cmd *cobra.Command, args []string) error {

	//start := time.Now() // 获取当前时间

	fmt.Println(core.Mule)

	fmt.Println(core.Version)

	opts, err := ParseInput(cmd)

	if err != nil {
		panic(err)
		return nil
	}

	CustomClient := &core.CustomClient{}

	CustomClient, err = CustomClient.NewHttpClient(opts)

	CustomClient.Headers = opts.Headers

	err = core.ScanTask(maincontext, *opts, CustomClient)

	if err != nil {
		return err
	}

	//elapsed := time.Since(start)
	//fmt.Println("该函数执行完成耗时：", elapsed)

	return nil
}

func ParseInput(cmd *cobra.Command) (*core.Options, error) {
	opts := core.Options{}

	var err error
	var FTargets []string

	DefaultDic, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	opts.DirRoot = DefaultDic

	// 预处理url
	STarget, err := cmd.Flags().GetString("url")

	if err != nil {
		return nil, fmt.Errorf("invalid value for url: %w", err)
	}

	FTarget, err := cmd.Flags().GetString("urls")

	if err != nil {
		return nil, fmt.Errorf("invalid value for urls: %w", err)
	}

	if FTarget == "" && STarget == "" {
		return nil, fmt.Errorf("Please input the target")
	} else if FTarget != "" && STarget == "" {
		FTargets, err = utils.ReadTarget(FTarget)

		if err != nil {
			return nil, fmt.Errorf("please check target file")
		}
	} else if FTarget == "" && STarget != "" {
		FTargets = append(FTargets, STarget)
	} else {
		return nil, fmt.Errorf("only input u or U,cannot use in the same time")
	}

	for _, t := range FTargets {
		temp, err := utils.HandleTarget(t)
		if err != nil {
			return nil, err
		}

		opts.Target = append(opts.Target, temp)
	}

	// 字典存活验证(现在放到后面读取处进行验证)
	dicstring, err := cmd.Flags().GetString("dic")
	if err != nil {
		return nil, fmt.Errorf("invalid value for dictionary: %w", err)
	}

	if dicstring != "" {
		opts.Dictionary = append(opts.Dictionary, dicstring)
	}

	defaultstring, err := cmd.Flags().GetString("flag")
	if err != nil {
		return nil, fmt.Errorf("invalid value for dictionary: %w", err)
	}

	defslice := utils.GetDefaultList(defaultstring)

	opts.Dictionary = append(opts.Dictionary, defslice...)

	//alive, err := core.PathExists(opts.Dictionary)
	//
	//if !alive{
	//	panic("please check your dict")
	//}

	// 处理block数量
	core.Block, err = cmd.Flags().GetInt("block")

	if err != nil {
		return nil, fmt.Errorf("invalid value for url: %w", err)
	}

	// 处理输入的header
	headers, err := cmd.Flags().GetStringArray("Headers")
	if err != nil {
		return nil, fmt.Errorf("invalid value for headers: %w", err)
	}

	for _, h := range headers {
		keyAndValue := strings.SplitN(h, ":", 2)
		if len(keyAndValue) != 2 {
			return &opts, fmt.Errorf("invalid header format for header %q", h)
		}
		key := strings.TrimSpace(keyAndValue[0])
		value := strings.TrimSpace(keyAndValue[1])
		if len(key) == 0 {
			return &opts, fmt.Errorf("invalid header format for header %q - name is empty", h)
		}
		header := core.HTTPHeader{Name: key, Value: value}
		opts.Headers = append(opts.Headers, header)
	}

	// 处理blasklist

	core.BlackList, err = cmd.Flags().GetIntSlice("blacklist")

	if err != nil {
		return nil, fmt.Errorf("invalid value for blacklist: %w", err)
	}

	opts.Cookie, err = cmd.Flags().GetString("Cookie")

	if err != nil {
		return nil, fmt.Errorf("invalid value for cookie: %w", err)
	}

	opts.Thread, err = cmd.Flags().GetInt("Thread")

	if err != nil {
		return nil, fmt.Errorf("invalid value for Thread: %w", err)
	}

	opts.Timeout, err = cmd.Flags().GetInt("timeout")

	if err != nil {
		return nil, fmt.Errorf("invalid value for timeout: %w", err)
	}

	LogFile, err := cmd.Flags().GetString("output")

	if err != nil {
		return nil, fmt.Errorf("invalid value for LogFile: %w", err)
	}

	core.InitLogger(LogFile)

	return &opts, nil

}
