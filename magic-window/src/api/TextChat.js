import request from '../utils/request.js'


export const selectTextChatContentService = (assistantData) => {
    return request.post('/api/chat/selectTextChatContent',assistantData);
}

export const addTextChatContentService = (assistantData) => {
    return request.post('/api/chat/addTextChatContent', assistantData);
}


export const deleteTextChatContentService = (assistantData) => {
    return request.post('/api/chat/deleteTextChatContent', assistantData);
}