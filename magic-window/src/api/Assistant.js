import request from '../utils/request.js'

export const addAssistantService = (assistantData) => {
    return request.post('/api/robots/addAssistant', assistantData);
}

export const selectAllService = () => {
    const email = sessionStorage.getItem('user');
    return request.post('/api/robots/getAssistant',{email:email});
}

export const editAssistantService = (assistantData) => {
    return request.post('/api/robots/updateAssistant', assistantData);
}

export const deleteAssistantService = (assistantData) => {
    return request.post('/api/robots/deleteAssistant', assistantData);
}