import request from '../utils/request.js'

export const loginService = (userData) => {
    return request.post('/api/users/login',userData);
}

export const registerService = (userData) => {
    return request.post('/api/users/register', userData);
}
export const sendVerificationCode = (data)=>{
    return request.post('/api/users/sendEmailCode',data);
}
