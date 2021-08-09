package core

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type CustomClient struct {
	CuClient *http.Client
	Method   string
	Headers  []HTTPHeader
}

type HTTPHeader struct {
	Name  string
	Value string
}

func (custom *CustomClient) NewHttpClient(Opt *Options) (*CustomClient, error) {
	custom.CuClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			//DisableKeepAlives: true,
		},
		Timeout: time.Second * time.Duration(Opt.Timeout),
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	custom.Method = Opt.Method
	custom.Headers = Opt.Headers

	return custom, nil
}

func (custom *CustomClient) RunRequest(ctx context.Context, Url string) (*ReqRes, error) {

	response, err := custom.DoRequest(ctx, Url)

	result := ReqRes{}

	if err != nil {
		// ignore context canceled errors
		if errors.Is(ctx.Err(), context.Canceled) {
			return nil, err
		}
		// 输出错误,暂时看下来context deadline的页面都是无意义的页面,如果以后出现再解决
		//Logger.Error("RunRequestErr", zap.String("Error", err.Error()))
		return nil, err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	response.Body.Close()
	result.StatusCode = response.StatusCode
	result.Header = response.Header
	result.Length = int64(len(body))
	result.Body = body

	return &result, nil

}

func (custom *CustomClient) DoRequest(ctx context.Context, Url string) (response *http.Response, err error) {

	//request, err := http.NewRequest(custom.Method, Url, nil)

	request, err := http.NewRequest("GET", Url, nil)

	if err != nil {
		return nil, err
	}

	request = request.WithContext(ctx)

	for _, header := range custom.Headers {
		request.Header.Set(header.Name, header.Value)
	}

	response, err = custom.CuClient.Do(request)

	if err != nil {
		var ue *url.Error
		if errors.As(err, &ue) {
			if strings.HasPrefix(ue.Err.Error(), "x509") {
				return nil, fmt.Errorf("invalid certificate: %w", ue.Err)
			}
		}
		return nil, err
	}

	return response, nil

}
