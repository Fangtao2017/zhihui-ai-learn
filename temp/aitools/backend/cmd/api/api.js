export const fetchWithAuth = async (url, options = {}) => {
    const token = localStorage.getItem('token');
    const headers = {
        'Content-Type': 'application/json',
        ...options.headers,
    };

    if (token) {
        headers.Authorization = `Bearer ${token}`;
    }

    const response = await fetch(url, {
        ...options,
        headers,
    });

    if (response.status === 401) {
        // Token 过期或无效
        localStorage.removeItem('token');
        localStorage.removeItem('username');
        window.location.href = '/login';
        return null;
    }

    return response;
};