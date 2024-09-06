package main

import "github.com/mattermost/mattermost/server/public/model"

// userHasRemovePermissionsToPost checks if the user has permissions to remove attachments from a post
// based on the user ID, channel, and post. It returns an error message if the user does not have permissions,
// The channel is used to get the team ID to check for team permissions.
func (p *Plugin) userHasRemovePermissionsToPost(userID string, channel *model.Channel, post *model.Post) string {
	// Check if the post exists
	post, appErr := p.API.GetPost(post.Id)
	if appErr != nil {
		return "Post does not exist"
	}

	// Check if the post has attachments
	if len(post.FileIds) == 0 {
		return "Post has no attachments"
	}

	// Check if the user is the post author or has permissions to edit others posts
	user, appErr := p.API.GetUser(userID)
	if appErr != nil {
		return "Internal error, check with your system administrator for assistance"
	}

	var permission *model.Permission
	if post.UserId == user.Id {
		permission = model.PermissionEditPost
	} else {
		permission = model.PermissionEditOthersPosts
	}

	if !p.API.HasPermissionToChannel(userID, channel.Id, permission) && !p.API.HasPermissionToTeam(userID, channel.TeamId, permission) {
		return "Not authorized"
	}

	// Check if the post is editable at this point in time
	config := p.API.GetConfig()
	if config.ServiceSettings.PostEditTimeLimit != nil && *config.ServiceSettings.PostEditTimeLimit > 0 && model.GetMillis() > post.CreateAt+int64(*config.ServiceSettings.PostEditTimeLimit*1000) {
		return "Post is too old to edit"
	}

	return ""
}
