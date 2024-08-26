import {Store, Action} from 'redux';

import {GlobalState} from 'mattermost-redux/types/store';
import {getPost} from 'mattermost-redux/selectors/entities/posts';

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
                return typeof post.file_ids?.length !== 'undefined' && post.file_ids?.length > 0;
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
