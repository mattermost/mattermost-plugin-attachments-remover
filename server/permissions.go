package main

import "github.com/mattermost/mattermost/server/public/model"

// userHasRemovePermissionsToPost checks if the user has permissions to remove attachments from a post
// based on the post ID, the user ID, and the channel ID.
// Returns an error message if the user does not have permissions, or an empty string if the user has permissions.
func (p *Plugin) userHasRemovePermissionsToPost(userID, channelID, postID string) string {
	// Check if the post exists
	post, appErr := p.API.GetPost(postID)
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

	if post.UserId != user.Id && !p.API.HasPermissionToChannel(userID, channelID, model.PermissionEditOthersPosts) {
		return "Not authorized"
	}

	// Check if the post is editable at this point in time
	config := p.API.GetConfig()
	if config.ServiceSettings.PostEditTimeLimit != nil && *config.ServiceSettings.PostEditTimeLimit > 0 && model.GetMillis() > post.CreateAt+int64(*config.ServiceSettings.PostEditTimeLimit*1000) {
		return "Post is too old to edit"
	}

	return ""
}
