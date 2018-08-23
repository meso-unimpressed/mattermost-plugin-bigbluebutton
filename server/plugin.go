/*
Copyright 2018 Blindside Networks

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

package main

import (
	"net/http"
	"strings"
	"fmt"
	"sync/atomic"

	bbbAPI "github.com/blindsidenetworks/mattermost-plugin-bigbluebutton/server/bigbluebuttonapiwrapper/api"
	"github.com/blindsidenetworks/mattermost-plugin-bigbluebutton/server/bigbluebuttonapiwrapper/dataStructs"
	BBBwh "github.com/blindsidenetworks/mattermost-plugin-bigbluebutton/server/bigbluebuttonapiwrapper/webhook"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
	"github.com/robfig/cron"
)
//test

type Plugin struct {

	plugin.MattermostPlugin

	// api                          plugin.API
	c                            *cron.Cron
	configuration                atomic.Value
	Meetings                     []dataStructs.MeetingRoom
	MeetingsWaitingforRecordings []dataStructs.MeetingRoom
	webhooks                     []*dataStructs.WebHook
	Hookid                       string
}

//OnActivate runs as soon as plugin activates
func (p *Plugin) OnActivate() error {
	// we save all the meetings infos that are stored on in our database upon deactivation
	// loads the details back so everything works
	p.LoadMeetingsFromStore()

	if err := p.OnConfigurationChange(); err != nil {
		return err
	}
	config := p.config()
	if err := config.IsValid(); err != nil {
		return err
	}

	BBBwh.SetWebhookAPI(config.BASE_URL+"/hooks/", config.SALT)
	bbbAPI.SetAPI(config.BASE_URL+"/", config.SALT)

	hookid := p.createEndMeetingWebook(config.CallBack_URL+"/plugins/bigbluebutton/webhookendmeeting", "")
	p.Hookid = hookid

	//every 2 minutes, look through active meetings and check if recordings are done
	p.c = cron.New()
	p.c.AddFunc("@every 2m", p.Loopthroughrecordings)
	p.c.Start()

	// register slash command '/bbb' to create a meeting
	return p.API.RegisterCommand(&model.Command{
		Trigger:          "bbb",
		AutoComplete:     true,
		AutoCompleteDesc: "Create a BigBlueButton meeting",
	})
}

func (p *Plugin) OnConfigurationChange() error {
	var configuration Configuration
	// loads configuration from our config ui page
	err := p.API.LoadPluginConfiguration(&configuration)
	//stores the config in an Atomic.Value place
	p.configuration.Store(&configuration)
	return err
}
func (p *Plugin) config() *Configuration {
	//returns the config file we had stored in Atomic.Value
	return p.configuration.Load().(*Configuration)
}

//following method is to create a meeting from '/bbb' slash command
func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {

	meetingpointer := new(dataStructs.MeetingRoom)
	p.PopulateMeeting(meetingpointer, nil, "")

	p.createStartMeetingPost(args.UserId, args.ChannelId, meetingpointer)
	p.Meetings = append(p.Meetings, *meetingpointer)
	return &model.CommandResponse{}, nil

}

//this is the router to handle our server calls
//methods are all in responsehandlers.go
func (p *Plugin) ServeHTTP(c *plugin.Context,w http.ResponseWriter, r *http.Request) {

	config := p.config()
	if err := config.IsValid(); err != nil {
		http.Error(w, "This plugin is not configured.", http.StatusNotImplemented)
		return
	}

	path := r.URL.Path
	if path == "/joinmeeting" {
		p.handleJoinMeeting(w, r)
	} else if strings.HasPrefix(path, "/endmeeting") {
		p.handleEndMeeting(w, r)
	} else if path == "/create" {
		p.handleCreateMeeting(w, r)
	} else if strings.HasPrefix(path, "/webhookendmeeting") {
		p.handleWebhookMeetingEnded(w, r)
	} else if strings.HasPrefix(path, "/recordingready") {
		p.handleRecordingReady(w, r)
	} else if path == "/getattendees" {
		p.handleGetAttendeesInfo(w, r)
	} else if path == "/publishrecordings" {
		p.handlePublishRecordings(w, r)
	} else if path == "/deleterecordings" {
		p.handleDeleteRecordings(w, r)
	} else if strings.HasPrefix(path, "/meetingendedcallback") {
		p.handleImmediateEndMeetingCallback(w, r, path)
	} else if path == "/ismeetingrunning" {
		p.handleIsMeetingRunning(w, r)
	} else if path == "/redirect"{
			fmt.Fprintf(w,`<!doctype html><html><head><script>
				 								window.onload = function load() {
													window.open('', '_self', '');
													window.close();
													};
											</script></head><body></body></html>`)
	}else {
		http.NotFound(w, r)
	}
	return
}

func (p *Plugin) OnDeactivate() error {
	//on deactivate, save meetings details, stop check recordings looper, destroy webhook
	p.SaveMeetingToStore()
	p.c.Stop()
	BBBwh.DestroyHook(p.Hookid)
	return nil
}

func main() {
	plugin.ClientMain(&Plugin{})
}
