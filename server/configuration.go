package main

import (
	"fmt"

	"github.com/mattermost/mattermost-server/model"
)

// OnConfigurationChange is invoked when configuration changes may have been made.
//
// This demo implementation ensures the configured demo user and channel are created for use
// by the plugin.
func (p *Plugin) OnConfigurationChange() error {
	// Leverage the default implementation on the embedded plugin.Mattermost. This
	// automatically attempts to unmarshal the plugin config block of the server's
	// configuration onto the public members of Plugin, such as Username and ChannelName.
	//
	// Feel free to skip this and implement your own handler if you have more complex needs.
	if err := p.MattermostPlugin.OnConfigurationChange(); err != nil {
		p.API.LogError(err.Error())
		return err
	}

	if err := p.ensureDemoUser(); err != nil {
		p.API.LogError(err.Error())
		return err
	}

	if err := p.ensureDemoChannels(); err != nil {
		p.API.LogError(err.Error())
		return err
	}

	return nil
}

func (p *Plugin) ensureDemoUser() *model.AppError {
	var err *model.AppError

	// Check for the configured user. Ignore any error, since it's hard to distinguish runtime
	// errors from a user simply not existing.
	user, _ := p.API.GetUserByUsername(p.Username)

	// Ensure the configured user exists.
	if user == nil {
		user, err = p.API.CreateUser(&model.User{
			Username:  p.Username,
			Password:  "password",
			Email:     fmt.Sprintf("%s@example.com", p.Username),
			Nickname:  "Demo Day",
			FirstName: "Demo",
			LastName:  "Plugin User",
			Position:  "Bot",
		})

		if err != nil {
			return err
		}
	}

	teams, err := p.API.GetTeams()
	if err != nil {
		return err
	}

	for _, team := range teams {
		// Ignore any error.
		p.API.CreateTeamMember(team.Id, p.demoUserId)
	}

	// Save the id for later use.
	p.demoUserId = user.Id

	return nil
}

func (p *Plugin) ensureDemoChannels() *model.AppError {
	teams, err := p.API.GetTeams()
	if err != nil {
		return err
	}

	p.demoChannelIds = make(map[string]string)
	for _, team := range teams {
		// Check for the configured channel. Ignore any error, since it's hard to
		// distinguish runtime errors from a channel simply not existing.
		channel, _ := p.API.GetChannelByNameForTeamName(team.Name, p.ChannelName, false)

		// Ensure the configured channel exists.
		if channel == nil {
			channel, err = p.API.CreateChannel(&model.Channel{
				TeamId:      team.Id,
				Type:        model.CHANNEL_OPEN,
				DisplayName: "Demo Plugin",
				Name:        p.ChannelName,
				Header:      "The channel used by the demo plugin.",
				Purpose:     "This channel was created by a plugin for testing.",
			})

			if err != nil {
				return err
			}
		}

		// Save the ids for later use.
		p.demoChannelIds[team.Id] = channel.Id
	}

	return nil
}
