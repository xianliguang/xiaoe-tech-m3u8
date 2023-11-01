package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

const VideoInfoDetailUrl string = "https://appsve3sa4d7572.pc.xiaoe-tech.com/xe.course.business.video.detail_info.get/2.0.0"
const ShopConfigUrl string = "https://appsve3sa4d7572.pc.xiaoe-tech.com/xe.course.business.avoidlogin.e_course.shop_config.get/1.0.0"

var (
	resourceId = "v_630f1b58e4b0c942648f1790"
	oprSys     = "MacIntel"
	productId  = "p_63101d44e4b0a51fef14c62e"
	pcUserKey  = "1edf46000134a8a6ff7dbad83c209761"
	r          = regexp.MustCompile("(.+\\.ts\\?.+)")
)

func main() {
	flag.StringVar(&resourceId, "resource_id", "", "resource_id")
	flag.StringVar(&productId, "product_id", "", "product_id")
	flag.StringVar(&pcUserKey, "user_key", "", "user_key")

	flag.Parse()

	if resourceId == "" || productId == "" || pcUserKey == "" {
		fmt.Println("请输入resource_id,product_id以及user_key")
		return
	}

	uid := fetchUserId()
	name, urls := fetchVideoInfo()
	fmt.Println(uid, name, urls)
	for _, url := range urls {
		fetchM3u8(name, url)
	}
}

type VideoUrl struct {
	DefinitionName string `json:"definition_name"`
	Url            string `json:"url"`
	Ext            struct {
		Host string `json:"host"`
		Path string `json:"path"`
		Parm string `json:"param"`
	} `json:"ext"`
}

func fetchM3u8(name string, url VideoUrl) string {
	resp, _ := http.Get(url.Url)
	body, _ := io.ReadAll(resp.Body)

	f, _ := os.Create(name + ".m3u8")

	f.WriteString(r.ReplaceAllString(string(body), url.Ext.Host+"/"+url.Ext.Path+"/$1"))
	return f.Name()
}

func parseVideoUrls(s string) []VideoUrl {
	s = strings.Replace(s, "__ba", "", -1)
	s = strings.Replace(s, "@", "1", -1)
	s = strings.Replace(s, "#", "2", -1)
	s = strings.Replace(s, "$", "3", -1)
	s = strings.Replace(s, "%", "4", -1)
	s = strings.Replace(s, "_", "+", -1)
	s = strings.Replace(s, "-", "+", -1)

	data, _ := base64.StdEncoding.DecodeString(s)
	fmt.Println(string(data))

	var vus []VideoUrl
	json.Unmarshal(data, &vus)
	return vus
}

func fetchVideoInfo() (string, []VideoUrl) {
	data := make(url.Values)
	data.Set("resource_id", resourceId)
	data.Set("opr_sys", oprSys)
	data.Set("product_id", productId)

	req, _ := http.NewRequest("POST", VideoInfoDetailUrl, strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", fmt.Sprintf("pc_user_key=%s;", pcUserKey))

	var hc http.Client
	resp, _ := hc.Do(req)
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
	var video map[string]interface{}
	json.Unmarshal(body, &video)
	d := video["data"].(map[string]interface{})
	vus := parseVideoUrls(d["video_urls"].(string))
	vi := d["video_info"].(map[string]interface{})
	vn := vi["file_name"].(string)

	return vn, vus
}

func fetchUserId() string {
	req, _ := http.NewRequest("POST", ShopConfigUrl, nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", fmt.Sprintf("pc_user_key=%s;", pcUserKey))

	var hc http.Client
	resp, _ := hc.Do(req)
	body, _ := io.ReadAll(resp.Body)

	var video map[string]interface{}
	json.Unmarshal(body, &video)
	d := video["data"].(map[string]interface{})
	ui := d["user_info"].(map[string]interface{})
	return ui["user_id"].(string)
}
