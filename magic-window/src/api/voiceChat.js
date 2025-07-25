import request from '../utils/request.js'


export const selectVoiceChatContentService = (assistantData) => {
    return request.post('/api/chat/selectVoiceChatContent',assistantData);
}



export const deleteVoiceChatContentService = (assistantData) => {
    return request.post('/api/chat/deleteVoiceChatContent', assistantData);
}