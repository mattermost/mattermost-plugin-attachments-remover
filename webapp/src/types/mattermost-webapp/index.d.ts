type PostMenuAction = (postID: string) => void;
type PostMenuFilter = (postID: string) => boolean;

export interface PluginRegistry {
    registerPostTypeComponent(typeName: string, component: React.ElementType)
    registerPostDropdownMenuAction(text: string, action?: PostMenuAction, filter?: PostMenuFilter)

    // Add more if needed from https://developers.mattermost.com/extend/plugins/webapp/reference
}
