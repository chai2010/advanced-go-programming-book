// Copyright 2017 <chaishushan{AT}gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

// https://api.github.com/repos/{user}/{repo}/contributors
// https://api.github.com/repos/chai2010/advanced-go-programming-book/contributors

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
)

var (
	flagFile = flag.String("file", "contributors.json", "set contributors file")
)

type User struct {
	Login     string `json:"login"`
	AvatarUrl string `json:"avatar_url"`
}

func main() {
	flag.Parse()

	content, err := ioutil.ReadFile(*flagFile)
	if err != nil {
		log.Fatal(err)
	}

	var users []User
	if err := json.Unmarshal(content, &users); err != nil {
		log.Fatal(err)
	}

	// skip chai2010 and cch123
	fmt.Println(genTableHeader(users[2:]))
	fmt.Println(genTableHeaderSepLine())
	fmt.Println(genTableElemLines(users[2:]))
}

func genTableHeader(users []User) string {
	var s = "|"
	for i := 0; i < len(users) && i < 7; i++ {
		s += fmt.Sprintf(
			` [<img src="%s" width="100px;"/><br /><sub><b>%s</b></sub>](https://github.com/%s) |`,
			users[i].AvatarUrl,
			users[i].Login,
			users[i].Login,
		)
	}
	return s
}
func genTableHeaderSepLine() string {
	return "| :---: | :---: | :---: | :---: | :---: | :---: | :---: |"
}
func genTableElemLines(users []User) string {
	var s string
	for i := 7; i < len(users); i += 7 {
		s += genTableHeader(users[i:]) + "\n"
	}
	return s
}
