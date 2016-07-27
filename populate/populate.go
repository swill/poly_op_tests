package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/nlopes/slack"
	"github.com/vharitonsky/iniflags"
	"io/ioutil"
	"os"
	"strings"
)

var (
	token   = flag.String("token", "", "Your Slack token")
	channel = flag.String("channel", "", "The Slack channel to post to (without the leading '#')")
)

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Points []Point

type TestCase struct {
	Subject Points `json:"subject"`
	Object  Points `json:"object"`
}

type TestCases []TestCase

func main() {
	iniflags.Parse()

	api := slack.New(*token)
	channels, err := api.GetChannels(true)
	if err != nil {
		fmt.Println("\nERROR: Could not get the Slack channels\n")
		fmt.Println(err)
		os.Exit(2)
	}
	var channel_id string
	for _, c := range channels {
		if c.Name == *channel {
			channel_id = c.ID
		}
	}
	if channel_id == "" {
		fmt.Println("\nERROR: Could not find the Slack channel specified.  Be sure NOT to include the '#' at the beginning.\n")
		os.Exit(2)
	}

	params := slack.NewHistoryParameters()
	params.Count = 1000
	history, err := api.GetChannelHistory(channel_id, params)

	var stab_points, switch_points string
	var testcases TestCases
	for _, m := range history.Messages {
		if strings.HasPrefix(m.Msg.Text, "`stab_points: ") {
			stab_points = m.Msg.Text[14 : len(m.Msg.Text)-1]
			stab_points = strings.Replace(stab_points, " ", ", ", -1)
			stab_points = strings.Replace(stab_points, "X:", "\"x\":", -1)
			stab_points = strings.Replace(stab_points, "Y:", "\"y\":", -1)
		}
		if strings.HasPrefix(m.Msg.Text, "`switch_points: ") {
			switch_points = m.Msg.Text[16 : len(m.Msg.Text)-1]
			switch_points = strings.Replace(switch_points, " ", ", ", -1)
			switch_points = strings.Replace(switch_points, "X:", "\"x\":", -1)
			switch_points = strings.Replace(switch_points, "Y:", "\"y\":", -1)
		}

		if stab_points != "" && switch_points != "" {
			var stab_var Points
			var switch_var Points
			var testcase TestCase

			err := json.Unmarshal([]byte(stab_points), &stab_var)
			if err != nil {
				panic(err)
			}
			err = json.Unmarshal([]byte(switch_points), &switch_var)
			if err != nil {
				panic(err)
			}

			testcase.Subject = switch_var
			testcase.Object = stab_var
			testcases = append(testcases, testcase)

			// reset for next round since we have processed this one...
			stab_points, switch_points = "", ""
		}
	}
	fmt.Printf("Length: %d\n", len(testcases))
	json_str, err := json.Marshal(testcases)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile("../test_cases.json", json_str, 0644)
	if err != nil {
		panic(err)
	}
}
