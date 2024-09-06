package main

import (
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

func (p *Plugin) getCommand() *model.Command {
	return &model.Command{
		Trigger:      "removeattachments",
		DisplayName:  "Remove Attachments",
		Description:  "Remove all attachments from a post",
		AutoComplete: false, // Hide from autocomplete
	}
}

func (p *Plugin) createCommandResponse(message string) *model.CommandResponse {
	return &model.CommandResponse{
		Text: message,
	}
}

func (p *Plugin) createErrorCommandResponse(errorMessage string) *model.CommandResponse {
	return &model.CommandResponse{
		Text: "Can't remove attachments: " + errorMessage,
	}
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	commandSplit := strings.Split(args.Command, " ")

	// Do not provide command output since it's going to be triggered from the frontend
	if len(commandSplit) != 2 {
		return p.createErrorCommandResponse("Invalid number of arguments, use `/removeattachments [postID]`."), nil
	}

	postID := commandSplit[1]

	// Check if the post ID is a valid ID
	if !model.IsValidId(postID) {
		return p.createErrorCommandResponse("Invalid post ID"), nil
	}

	// Check if the post exists
	post, appErr := p.API.GetPost(postID)
	if appErr != nil {
		return p.createErrorCommandResponse(appErr.Error()), nil
	}

	channel, appErr := p.API.GetChannel(post.ChannelId)
	if appErr != nil {
		return p.createErrorCommandResponse(appErr.Error()), nil
	}

	// Check if the user has permissions to remove attachments from the post
	if errReason := p.userHasRemovePermissionsToPost(args.UserId, channel, post); errReason != "" {
		return p.createErrorCommandResponse(errReason), nil
	}

	// Create an interactive dialog to confirm the action
	if err := p.API.OpenInteractiveDialog(model.OpenDialogRequest{
		TriggerId: args.TriggerId,
		URL:       "/plugins/" + manifest.Id + "/api/v1/remove_attachments?post_id=" + postID,
		Dialog: model.Dialog{
			Title:            "Remove Attachments",
			IntroductionText: "Are you sure you want to remove all attachments from this post?",
			SubmitLabel:      "Remove",
		},
	}); err != nil {
		return p.createCommandResponse(err.Error()), nil
	}

	// Return nothing, let the dialog/api handle the response
	return &model.CommandResponse{}, nil
}
