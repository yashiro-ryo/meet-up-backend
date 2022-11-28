package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/xid"
)

var db *sql.DB
var err error

type Team struct {
	TeamId        int    `json:"teamId"`
	TeamName      string `json:"teamName"`
	TeamDescribe  string `json:"teamDescribe"`
	TeamUrl       string `json:"teamUrl"`
	TeamImage     string `json:"teamImage"`
	TeamCreateAt  string `json:"teamCreateAt"`
	TeamCreatedBy int    `json:"teamCreatedBy"`
	TeamAddress   string `json:"teamAddress"`
}

func SetupDB() {
	fmt.Println("setup database")
	db, err = sql.Open("mysql", "root:LTDEXPuzushio22@@tcp(localhost:3306)/meet-up?parseTime=true")
	if err != nil {
		// ここではエラーを返さない
		log.Fatal(err)
	}
}

func GetTeamDetail(idStr string) (Team, error) {
	fmt.Println("get team detail id=", idStr)
	rows, _ := db.Query("select team_id, team_name, team_describe, team_image, team_url from team where team_id = " + idStr)
	team := Team{}
	for rows.Next() {
		rows.Scan(&team.TeamId, &team.TeamName, &team.TeamDescribe, &team.TeamImage, &team.TeamUrl)
		fmt.Println("result", team)
	}
	return team, nil
}

func CreateTeam(teamName string, teamDescribe string, teamImage string, teamUrl string, keywords *Res) {
	fmt.Println(teamName, teamDescribe, teamImage, teamUrl)
	t := time.Now()
	t = t.Add(time.Duration(9) * time.Hour)
	if db == nil {
		fmt.Println("db is not found")
		return
	}
	fmt.Println("create talk room")
	// nounテーブルを作成するための一意なIDを発行
	guid := xid.New()
	uniqueId := guid.String()
	// insert team
	ins, err := db.Prepare("INSERT INTO team (team_id, team_name, team_describe, team_image, team_url, team_create_at, team_keyword_id) VALUES(null, ?,?,?,?,?,?)")
	if err != nil {
		log.Fatal(err)
	}
	ins.Exec(teamName, teamDescribe, teamImage, teamUrl, t, uniqueId)
	// insert noun team_keyword table
	nouns := (*keywords).Result
	for i := range nouns {
		ins, err = db.Prepare("INSERT INTO team_keyword (keyword_id, keyword_team_id, keyword_text) VALUES(null, ?, ?)")
		if err != nil {
			log.Fatal(err)
		}
		ins.Exec(uniqueId, nouns[i])
	}
}

func GetNewTeams() ([]Team, error) {
	if db == nil {
		fmt.Println("db is not found")
		return nil, errors.New("db is not found")
	}
	rows, err := db.Query("SELECT team_id, team_name, team_describe, team_url, team_image, team_create_at, team_created_by FROM team ORDER BY team_create_at DESC LIMIT 5")
	if err != nil {
		log.Fatal(err)
		return nil, errors.New("error")
	}
	var teams []Team
	for rows.Next() {
		var teamId int
		var teamName string
		var teamDescribe string
		var teamUrl string
		var teamImage string
		var teamCreateAt string
		var teamCreatedBy int
		rows.Scan(&teamId, &teamName, &teamDescribe, &teamUrl, &teamImage, &teamCreateAt, &teamCreatedBy)
		fmt.Println("team", teamId, teamName, teamDescribe, teamUrl, teamImage, teamCreateAt, teamCreatedBy)
		team := Team{
			TeamId:        teamId,
			TeamName:      teamName,
			TeamDescribe:  teamDescribe,
			TeamUrl:       teamUrl,
			TeamImage:     teamImage,
			TeamCreateAt:  teamCreateAt,
			TeamCreatedBy: teamCreatedBy,
		}
		teams = append(teams, team)
	}
	return teams, nil
}

func FindTeam(query string) ([]Team, error) {
	// queryを空白ごとに分割する
	splitedQuery := strings.Split(query, " ")
	// 単語ごとに検索をかける
	fmt.Println(splitedQuery)
	var idsStr []string
	for i := range splitedQuery {
		rows, err := db.Query("select keyword_team_id from team_keyword where keyword_text = '" + splitedQuery[i] + "'")
		if err != nil {
			fmt.Println(err)
			return nil, nil
		}
		for rows.Next() {
			var teamIdStr string
			rows.Scan(&teamIdStr)
			fmt.Println("result", teamIdStr)
			idsStr = append(idsStr, teamIdStr)
		}
		fmt.Println(idsStr)
	}

	// idsStrの重複を削除する
	idsStr = sliceUnique(idsStr)

	var teams []Team
	for i := range idsStr {
		// get team info
		rows, err := db.Query("SELECT team_id, team_name, team_describe, team_url, team_image, team_create_at FROM team where team_keyword_id = '" + idsStr[i] + "'")
		if err != nil {
			log.Fatal(err)
			return nil, errors.New("error")
		}
		for rows.Next() {
			var teamId int
			var teamName string
			var teamDescribe string
			var teamUrl string
			var teamImage string
			var teamCreateAt string
			rows.Scan(&teamId, &teamName, &teamDescribe, &teamUrl, &teamImage, &teamCreateAt)
			fmt.Println("team", teamId, teamName, teamDescribe, teamUrl, teamImage, teamCreateAt)
			team := Team{
				TeamId:       teamId,
				TeamName:     teamName,
				TeamDescribe: teamDescribe,
				TeamUrl:      teamUrl,
				TeamImage:    teamImage,
				TeamCreateAt: teamCreateAt,
			}
			teams = append(teams, team)
		}
	}
	return teams, nil
}

func CloseDB() {
	fmt.Println("clear database")
	defer db.Close()
}

// lib
func sliceUnique(target []string) (unique []string) {
	m := map[string]bool{}

	for _, v := range target {
		if !m[v] {
			m[v] = true
			unique = append(unique, v)
		}
	}

	return unique
}
