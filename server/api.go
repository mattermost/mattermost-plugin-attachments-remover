package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost/server/public/model"
)

type API struct {
	plugin *Plugin
	router *mux.Router
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.router.ServeHTTP(w, r)
}

func (a *API) handlerRemoveAttachments(w http.ResponseWriter, r *http.Request) {
	// Get post_id from the query parameters
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	postID := r.FormValue("post_id")
	if !model.IsValidId(postID) {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Check if the user is the post author or a system admin
	userID := r.Header.Get("Mattermost-User-ID")
	if !model.IsValidId(userID) {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	post, appErr := a.plugin.API.GetPost(postID)
	if appErr != nil {
		http.Error(w, appErr.Error(), appErr.StatusCode)
		return
	}

	if errReason := a.plugin.userHasRemovePermissionsToPost(userID, post.ChannelId, postID); errReason != "" {
		http.Error(w, errReason, http.StatusForbidden)
		return
	}

	// Otherwise, update the post without attachments and then soft-delete the attachments from the channel
	originalFileIDs := post.FileIds

	post.FileIds = []string{}
	newPost, appErr := a.plugin.API.UpdatePost(post)
	if appErr != nil {
		http.Error(w, appErr.Error(), appErr.StatusCode)
		return
	}

	// Soft-delete the attachments from channel
	for _, fileID := range post.FileIds {
		if err := a.plugin.SQLStore.DetachAttachmentFromChannel(fileID); err != nil {
			a.plugin.API.LogError("error detaching attachment from channel", "err", err)
			http.Error(w, "Internal server error, check logs", http.StatusInternalServerError)
			return
		}
	}

	// Attach the original file IDs to the original post so history is not lost
	if err := a.plugin.SQLStore.AttachFileIDsToPost(newPost.OriginalId, originalFileIDs); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// setupAPI sets up the API for the plugin.
func setupAPI(plugin *Plugin) (*API, error) {
	api := &API{
		plugin: plugin,
		router: mux.NewRouter(),
	}

	group := api.router.PathPrefix("/api/v1").Subrouter()
	group.Use(authorizationRequiredMiddleware)
	group.HandleFunc("/remove_attachments", api.handlerRemoveAttachments).Methods(http.MethodPost)

	return api, nil
}

func authorizationRequiredMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("Mattermost-User-ID")
		if userID != "" {
			next.ServeHTTP(w, r)
			return
		}

		http.Error(w, "Not authorized", http.StatusUnauthorized)
	})
}
