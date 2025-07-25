const baseURL = '/api';

const request = {
  async get(url) {
    try {
      const response = await fetch(baseURL + url, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json'
        }
      });
      
      return handleResponse(response);
    } catch (err) {
      return handleError(err);
    }
  },
  
  async post(url, data) {
    try {
      const response = await fetch(baseURL + url, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(data)
      });
      

      return handleResponse(response);
    } catch (err) {
      return handleError(err);
    }
  }
};

async function handleResponse(response) {
  if (!response.ok) {
    alert('服务异常');
    throw new Error(`HTTP error! status: ${response.status}`);
  }
  
  const result = await response.json();//
  return result;
}

function handleError(err) {
  alert('服务异常');
  console.error('Request failed:', err);
  return Promise.reject(err);
}

export default request;