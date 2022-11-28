package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func SetupApiRouter() {
	e := echo.New()
	// 本番環境にするときはドメインを指定する
	e.Use(middleware.CORS())
	// api
	api := e.Group("/api")
	// v1
	v1 := api.Group("/v1")

	// 新着チーム募集を5件返すapi
	// TODO: よく表示される内容なのでキャッシュにしておくLRUなど
	v1.GET("/newteams", func(c echo.Context) error {
		teams, err := GetNewTeams()
		if err != nil {
			return c.String(http.StatusOK, "")
		}
		fmt.Println(teams)
		return c.JSON(http.StatusOK, teams)
	})

	// teamの詳細情報を取得するapi
	v1.GET("/detail/:id", func(c echo.Context) error {
		fmt.Println(c.Param("id"))
		team, err := GetTeamDetail(c.Param("id"))
		if err != nil {
			c.String(http.StatusOK, c.Param("teamId"))
		}
		return c.JSON(http.StatusOK, team)
	})

	// team募集を登録するapi
	v1.POST("/team", func(c echo.Context) error {
		var teamName = c.FormValue("teamName")
		var teamDescribe = c.FormValue("teamDescribe")
		var teamImage = c.FormValue("teamImage")
		var teamUrl = c.FormValue("teamUrl")
		keywords, err := httpPost("http://127.0.0.1:8000/split", teamDescribe)
		if keywords == nil || err != nil {
			return c.String(http.StatusOK, c.FormValue("teamName"))
		}
		CreateTeam(teamName, teamDescribe, teamImage, teamUrl, keywords)
		return c.String(http.StatusOK, c.FormValue("teamName"))
	})

	// team募集を更新するapi
	// team募集を削除するapi
	// 該当するteamを検索するapi
	v1.GET("/find", func(c echo.Context) error {
		q := c.QueryParam("q")
		fmt.Println("search query :", q)
		teams, err := FindTeam(q)
		fmt.Println(err)
		return c.JSON(http.StatusOK, teams)
	})
	e.Logger.Fatal(e.Start(":5050"))
}

type Res struct {
	Result []string `json:"result"`
}

func httpPost(url, text string) (*Res, error) {
	jsonStr := `{"text":"` + text + `"}`

	req, err := http.NewRequest(
		"POST",
		url,
		bytes.NewBuffer([]byte(jsonStr)),
	)
	if err != nil {
		return nil, err
	}

	// Content-Type 設定
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	fmt.Println("response", resp)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(b))
	var res *Res
	err = json.Unmarshal([]byte(string(b)), &res)
	if err != nil {
		fmt.Println("json error", err)
		return nil, err
	}
	fmt.Println("json", res)
	return res, nil
}
