package lazada_go_sdk

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type paramsMap map[string]string
type fileParamsMap map[string][]byte

//Client 拥有的方法
type Client interface {
	GetOrders(paramsMap) (*GetOrdersResponse, error)
	GetOrderItems(paramsMap) (*GetOrderItemsResponse, error)
	GetOrder(paramsMap) (*GetOrderResponse, error)
	//todo others
}

type ClientOptions struct {
	Secret string
	AppKey string
	Region string
}
type LazadaClient struct {
	GatewayUrl string
	AppKey     string
	Secret     string
	Method     string
	SysParams  paramsMap
	APIParams  paramsMap
	FileParams fileParamsMap
	Region     string
}

func NewClient(options *ClientOptions) Client {
	return &LazadaClient{
		Secret: options.Secret,
		SysParams: map[string]string{
			"app_key":     options.AppKey,
			"sign_method": "sha256",
			"timestamp":   fmt.Sprintf("%d000", time.Now().Unix()),
			//"partner_id":  Version,
		},
		APIParams:  map[string]string{},
		FileParams: map[string][]byte{},
		Region:     options.Region,
	}
}

//生成签名
func (lc *LazadaClient) generateSign(path string, apiParams map[string]string) string {

	//将系统公共参数和请求参数合并
	var keys []string
	union := map[string]string{}
	for key, val := range lc.SysParams {
		union[key] = val
		keys = append(keys, key)
	}
	for key, val := range apiParams {
		union[key] = val
		keys = append(keys, key)
	}
	//排序
	sort.Strings(keys)
	h := hmac.New(sha256.New, []byte(lc.Secret))
	io.WriteString(h, fmt.Sprintf("%s", path))
	//拼接
	for _, key := range keys {
		io.WriteString(h, fmt.Sprintf("%s%s", key, union[key]))
	}
	return strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
}

// Response 成功返回
type Response struct {
	Code      string              `json:"code"`
	Type      string              `json:"type"`
	Message   string              `json:"message"`
	RequestID string              `json:"request_id"`
	Data      jsoniter.RawMessage `json:"data"`
}

// ResponseError 错误返回
type ResponseError struct {
	Code      string `json:"code"`
	Type      string `json:"type"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

func (e ResponseError) Error() string {
	return fmt.Sprintf("lazadaError: code:%s - type:%s - message:%s - request_id:%s", e.Code, e.Type, e.Message, e.RequestID)
}

// 组装地址 API_GATEWAY_URL+path
func (lc *LazadaClient) getPathUrl(method string) string {
	switch lc.Region {
	case "SG":
		lc.GatewayUrl = API_GATEWAY_URL_SG
	case "MY":
		lc.GatewayUrl = API_GATEWAY_URL_MY
	case "VN":
		lc.GatewayUrl = API_GATEWAY_URL_VN
	case "TH":
		lc.GatewayUrl = API_GATEWAY_URL_TH
	case "PH":
		lc.GatewayUrl = API_GATEWAY_URL_PH
	case "ID":
		lc.GatewayUrl = API_GATEWAY_URL_ID
	default:
		lc.GatewayUrl = API_GATEWAY_URL_TH
	}

	return fmt.Sprintf("%s%s", lc.GatewayUrl, availablePaths[method])
}

//获取path
func (lc *LazadaClient) getPath(apiName string) string {
	return fmt.Sprintf("%s", availablePaths[apiName])
}

// Execute 发送请求
func (lc *LazadaClient) Execute(apiName string, httpMethod string, apiParams paramsMap, postBody fileParamsMap) (*Response, error) {
	//获取url
	pathUrl := lc.getPathUrl(apiName)
	//签名
	sign := lc.generateSign(lc.getPath(apiName), apiParams)
	//组装请求url
	values := url.Values{}
	//公共参数
	for key, value := range lc.SysParams {
		values.Add(key, value)
	}
	//get请求
	if httpMethod == http.MethodGet {
		//请求参数
		for key, value := range apiParams {
			values.Add(key, value)
		}
	}
	//初始化个bytes.Buffer
	bodyBuffer := new(bytes.Buffer)
	var contentType string
	//post请求
	if httpMethod == http.MethodPost {
		//以form-data格式上传
		//NewWriter returns a new multipart Writer with a random boundary, writing to w.
		bodyWriter := multipart.NewWriter(bodyBuffer)
		contentType = bodyWriter.FormDataContentType()
		//请求参数
		if len(postBody) > 0 {
			for name, content := range postBody {
				file, err := bodyWriter.CreateFormFile("files", name)
				if err != nil {
					return nil, err
				}
				_, err = file.Write(content)
				if err != nil {
					return nil, err
				}
			}
		}
		for key, val := range apiParams {
			_ = bodyWriter.WriteField(key, val)
		}

		//关闭writer
		err := bodyWriter.Close()
		if err != nil {
			return nil, err
		}
	}

	//加签名字符串
	values.Add("sign", sign)
	//urlencode
	reqUrl := fmt.Sprintf("%s?%s", pathUrl, values.Encode())
	fmt.Println(reqUrl)
	//新建request
	req, err := http.NewRequest(httpMethod, reqUrl, bodyBuffer)
	if err != nil {
		return nil, err
	}
	//设置header
	if contentType != "" {
		req.Header.Add("Content-Type", contentType)
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	respBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var errResp ResponseError
	_ = json.Unmarshal(respBody, &errResp)
	if errResp.Code != "0" {
		return nil, errResp
	}
	resp := &Response{}
	err = json.Unmarshal(respBody, &resp)
	return resp, nil
}

func (lc *LazadaClient) GetOrder(apiParams paramsMap) (resp *GetOrderResponse, err error) {
	b, err := lc.Execute("GetOrder", "GET", apiParams, nil)
	if err != nil {
		return
	}
	fmt.Println(string(b.Data))
	err = json.Unmarshal(b.Data, &resp)
	if err != nil {
		return
	}
	return
}

func (lc *LazadaClient) GetOrderItems(apiParams paramsMap) (resp *GetOrderItemsResponse, err error) {
	b, err := lc.Execute("GetOrderItems", "GET", apiParams, nil)
	if err != nil {
		return
	}
	fmt.Println(string(b.Data))
	err = json.Unmarshal(b.Data, &resp)
	if err != nil {
		return
	}
	return
}

func (lc *LazadaClient) GetOrders(apiParams paramsMap) (resp *GetOrdersResponse, err error) {
	b, err := lc.Execute("GetOrders", "GET", apiParams, nil)
	if err != nil {
		return
	}
	fmt.Println(string(b.Data))
	err = json.Unmarshal(b.Data, &resp)
	if err != nil {
		return
	}
	return
}
