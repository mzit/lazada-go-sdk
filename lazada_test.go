package lazada_go_sdk

import (
	"fmt"
	"testing"
)

const (
	AppKey      = ""
	Secret      = ""
	accessToken = ""
	Region      = "TH"
)

func TestGetOrders(t *testing.T) {
	var clientOptions = ClientOptions{
		AppKey: AppKey,
		Secret: Secret,
		Region: Region,
	}
	lc := NewClient(&clientOptions)
	p := paramsMap{
		"access_token":  accessToken,
		"update_after":  "2021-06-25T15:59:32+07:00",
		"update_before": "2021-06-25T16:12:31+07:00",
		"offset":        "0",
		"limit":         "10",
	}
	res, err := lc.GetOrders(p)
	if err != nil {
		fmt.Printf("%s", err)
	}
	//marshal, err := json.MarshalIndent(res, "", " ")
	//fmt.Println(string(marshal))
	fmt.Println(res)

}

func TestGetOrder(t *testing.T) {
	var clientOptions = ClientOptions{
		AppKey: AppKey,
		Secret: Secret,
		Region: Region,
	}
	lc := NewClient(&clientOptions)
	p := paramsMap{
		"access_token": accessToken,
		"order_id":     "406948191627528",
	}
	res, err := lc.GetOrder(p)
	if err != nil {
		fmt.Printf("%s", err)
	}
	//marshal, err := json.MarshalIndent(res, "", " ")
	//fmt.Println(string(marshal))
	fmt.Println(res)

}

func TestGetOrderItems(t *testing.T) {
	var clientOptions = ClientOptions{
		AppKey: AppKey,
		Secret: Secret,
		Region: Region,
	}
	lc := NewClient(&clientOptions)
	p := paramsMap{
		"access_token": accessToken,
		"order_id":     "406948191627528",
	}
	res, err := lc.GetOrderItems(p)
	if err != nil {
		fmt.Printf("%s", err)
	}

	//marshal, err := json.MarshalIndent(res, "", " ")
	//fmt.Println(string(marshal))
	fmt.Println(res)

}
