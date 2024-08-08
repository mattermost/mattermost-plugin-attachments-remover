import {AnyAction, Dispatch} from 'redux';
import {getCurrentChannel} from 'mattermost-redux/selectors/entities/channels';
import {getCurrentTeamId} from 'mattermost-redux/selectors/entities/teams';
import {GetStateFunc} from 'mattermost-redux/types/actions';
import {Client4} from 'mattermost-redux/client';
import {IntegrationTypes} from 'mattermost-redux/action_types';

export function setTriggerId(triggerId: string) {
    return {
        type: IntegrationTypes.RECEIVED_DIALOG_TRIGGER_ID,
        data: triggerId,
    };
}

export function triggerRemoveAttachmentsCommand(postID: string) {
    return (dispatch: Dispatch<AnyAction>, getState: GetStateFunc) => {
        const command = '/removeattachments ' + postID;
        clientExecuteCommand(dispatch, getState, command);

        return {data: true};
    };
}

export async function clientExecuteCommand(dispatch: Dispatch<AnyAction>, getState: GetStateFunc, command: string) {
    let currentChannel = getCurrentChannel(getState());
    const currentTeamId = getCurrentTeamId(getState());

    // Default to town square if there is no current channel (i.e., if Mattermost has not yet loaded)
    if (!currentChannel) {
        currentChannel = await Client4.getChannelByName(currentTeamId, 'town-square');
    }

    const args = {
        channel_id: currentChannel?.id,
        team_id: currentTeamId,
    };

    try {
        //@ts-ignore Typing in mattermost-redux is wrong
        const data = await Client4.executeCommand(command, args);
        dispatch(setTriggerId(data?.trigger_id));
    } catch (error) {
        console.error(error); //eslint-disable-line no-console
    }
}
