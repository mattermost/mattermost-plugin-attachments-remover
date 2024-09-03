import {Store, Action} from 'redux';

import {GlobalState} from 'mattermost-redux/types/store';
import {getPost} from 'mattermost-redux/selectors/entities/posts';
import {haveIChannelPermission} from 'mattermost-redux/selectors/entities/roles';
import {getCurrentUser} from 'mattermost-redux/selectors/entities/common';

import manifest from '@/manifest';

import {PluginRegistry} from '@/types/mattermost-webapp';

import {triggerRemoveAttachmentsCommand} from './actions';

export default class Plugin {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars, @typescript-eslint/no-empty-function
    public async initialize(registry: PluginRegistry, store: Store<GlobalState, Action<Record<string, unknown>>>) {
        // @see https://developers.mattermost.com/extend/plugins/webapp/reference/
        registry.registerPostDropdownMenuAction(
            'Remove attachments',
            async (postID) => {
                store.dispatch(triggerRemoveAttachmentsCommand(postID) as any);
            },
            (postID) => {
                const state = store.getState();
                const post = getPost(state, postID);

                // Don't show up if the post has no attachments. Permissions are checked server-side.
                if (!(typeof post.file_ids?.length !== 'undefined' && post.file_ids?.length > 0)) {
                    return false;
                }

                const user = getCurrentUser(state);

                // Check permissions for the user
                let permission = 'edit_post';
                if (post.user_id !== user.id) {
                    permission = 'delete_others_posts';
                }

                return haveIChannelPermission(state, {
                    channel: post.channel_id,
                    permission,
                });
            },
        );
    }
}

declare global {
    interface Window {
        registerPlugin(pluginId: string, plugin: Plugin): void
    }
}

window.registerPlugin(manifest.id, new Plugin());
