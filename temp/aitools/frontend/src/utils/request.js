const request = async (url, options = {}) => {
    const token = localStorage.getItem('token');
    
    // 设置请求超时时间（毫秒）
    const TIMEOUT = 30000; // 30秒
    
    // Basic configuration
    const defaultOptions = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': token ? `Bearer ${token}` : '',
        },
        // Not using credentials: 'include', because we're using Bearer token
    };

    // Merge configurations
    const finalOptions = {
        ...defaultOptions,
        ...options,
        headers: {
            ...defaultOptions.headers,
            ...options.headers,
        },
    };

    try {
        // 创建 AbortController 用于请求超时控制
        const controller = new AbortController();
        const signal = controller.signal;
        finalOptions.signal = signal;
        
        // 设置超时
        const timeoutId = setTimeout(() => {
            controller.abort();
        }, options.timeout || TIMEOUT);
        
        const response = await fetch(url, finalOptions);
        
        // 清除超时
        clearTimeout(timeoutId);
        
        // 对于登录请求，不自动重定向，而是返回错误
        if (response.status === 401) {
            if (url.includes('/api/login')) {
                // 如果是登录接口的 401 错误，创建一个包含状态码的错误对象
                const error = new Error('Invalid email or password');
                error.status = 401;
                throw error;
            } else {
                // 其他接口的 401 错误，按原来的逻辑处理
                localStorage.removeItem('token'); // Clear invalid token
                window.location.href = '/login'; // Redirect to login page
                throw new Error('请先登录');
            }
        }

        // Handle other errors
        if (!response.ok) {
            const errorText = await response.text();
            const error = new Error(errorText || response.statusText);
            error.status = response.status;
            throw error;
        }

        return await response.json();
    } catch (error) {
        console.error('Request failed:', error);
        
        // 处理不同类型的错误
        if (error.name === 'AbortError') {
            const abortError = new Error('请求超时，请重试');
            abortError.status = 408; // Request Timeout
            throw abortError;
        } else if (error.message.includes('NetworkError') || error.message.includes('Failed to fetch')) {
            const networkError = new Error('网络错误，请检查连接');
            networkError.status = 0; // Network Error
            throw networkError;
        }
        
        throw error;
    }
};

// Add convenient HTTP methods
request.get = (url, options = {}) => {
    return request(url, { ...options, method: 'GET' });
};

request.post = (url, data, options = {}) => {
    return request(url, {
        ...options,
        method: 'POST',
        body: JSON.stringify(data),
    });
};

request.put = (url, data, options = {}) => {
    return request(url, {
        ...options,
        method: 'PUT',
        body: JSON.stringify(data),
    });
};

request.delete = (url, options = {}) => {
    return request(url, { ...options, method: 'DELETE' });
};

export default request; 