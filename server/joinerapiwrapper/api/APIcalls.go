/*
Copyright 2020 Blindside Networks

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

package api

import (
	"github.com/blindsidenetworks/mattermost-plugin-bigbluebutton/server/bigbluebuttonapiwrapper/dataStructs"
	"github.com/blindsidenetworks/mattermost-plugin-bigbluebutton/server/joinerapiwrapper/helpers"
	"github.com/blindsidenetworks/mattermost-plugin-bigbluebutton/server/mattermost"
	"github.com/pkg/errors"
	"fmt"
)

type Request struct {
	MeetingID string `json:"meeting_id"`
	Password  string `json:"password"`
}

type Response struct {
	LinkID string `json:"id"`
}

//url of the joiner service
var joinBaseUrl string

//access token for the joiner service
var joinToken string

//Sets the BaseUrl and secret
func SetAPI(url string, tokenParam string) {
	joinBaseUrl = url
	joinToken = tokenParam
}

//CreateJoinLink creates A join link to a meeting
func CreateJoinLink(meetingRoom *dataStructs.MeetingRoom) (string, error) {
	if meetingRoom.MeetingID_ == "" {
		return "", errors.New("meeting ID cannot be empty")
	}

	if meetingRoom.AttendeePW_ == "" {
		return "", errors.New("attendee PW cannot be empty")
	}

	request := Request{meetingRoom.MeetingID_, meetingRoom.AttendeePW_}
	var response Response
	err := helpers.HttpPost(joinBaseUrl + "link", request, &response, joinToken)

	if err != nil {
		mattermost.API.LogError(fmt.Sprintf("ERROR: HTTP ERROR: %v", err))
		return "", err
	}

	return joinBaseUrl + response.LinkID, nil
}
