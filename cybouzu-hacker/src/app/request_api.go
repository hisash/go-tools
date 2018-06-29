package main

import (
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"text/template"
)

var (
	debug = flag.Bool("debug", false, "this option is able to output xml-body")
)

type SoapLayout struct {
	Users Users `xml:"Body>BaseGetUsersByLoginNameResponse>returns>user"`
}

type Users struct {
	LoginName           string `xml:"login_name,attr"`
	KanjiName           string `xml:"name,attr"`
	KanaName            string `xml:"reading,attr"`
	Email               string `xml:"email,attr"`
	Status              int    `xml:"status,attr"`
	PrimaryOrganization int    `xml:"primary_organization,attr"`
}

func main() {

	flag.Parse()

	var url = "https://cybozu.com/g/cbpapi/base/api.csp"

	parameters := make(map[string]string)
	parameters["userName"] = flag.Arg(0)
	parameters["password"] = flag.Arg(1)
	parameters["targetAccount"] = flag.Arg(2)

	if len(parameters) != len(flag.Args()) {
		fmt.Println("引数は" + strconv.Itoa(len(parameters)) + "つが必須です。")
		fmt.Println("サイボウズにログインするためのユーザー名、パスワード、サイボウズ上で調べたいユーザー名が必要です")
		fmt.Println("command [options] <userName> <password> <targetAccount>")
		os.Exit(1)
	}

	soapTemplate, err := ioutil.ReadFile("request_for_baase_get_users_by_login_name.xml")
	if err != nil {
		panic(err)
	}
	tmpl, err := template.New("request").Parse(string(soapTemplate))
	if err != nil {
		panic(err)
	}

	var doc bytes.Buffer
	if err := tmpl.Execute(&doc, parameters); err != nil {
		panic(err)
	}
	soapRequest := doc.String()

	response, err := http.Post(url, "application/xml; charset=utf-8", bytes.NewBufferString(soapRequest))
	if err != nil {
		panic(err)
	}

	responseBody, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		panic(err)
	}

	if *debug == true {
		fmt.Printf(string(responseBody))
		return
	}

	result := SoapLayout{}
	if err := xml.Unmarshal([]byte(responseBody), &result); err != nil {
		panic(err)
	}

	if result.Users.LoginName == "" {
		fmt.Println(parameters["targetAccount"] + ":nothing:nothing:nothing")
		return
	}

	roll := map[int]string{
		0: "enrolled",
		1: "retired",
	}

	if _, ok := roll[result.Users.Status]; ok == false {
		fmt.Println(result.Users.LoginName + "には予期せぬステータス:" + strconv.Itoa(result.Users.Status) + " がセットされているため、処理を中断しました")
		return
	}

	fmt.Println(result.Users.LoginName + ":" + result.Users.KanjiName + ":" + roll[result.Users.Status])
	return
}
