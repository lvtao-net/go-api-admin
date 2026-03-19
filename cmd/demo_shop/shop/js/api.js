// API 配置 - 使用相对路径，与当前服务同端口
const API_BASE_URL = '/api';

// API 请求封装
const api = {
    // 基础请求方法
    async request(url, options = {}) {
        const token = localStorage.getItem('token');
        const headers = {
            'Content-Type': 'application/json',
            ...options.headers
        };

        if (token) {
            headers['Authorization'] = `Bearer ${token}`;
        }

        try {
            const response = await fetch(`${API_BASE_URL}${url}`, {
                ...options,
                headers
            });

            const data = await response.json();

            if (data.code !== 0) {
                throw new Error(data.message || '请求失败');
            }

            return data;
        } catch (error) {
            console.error('API Error:', error);
            throw error;
        }
    },

    // GET 请求
    get(url, params = {}) {
        const queryString = new URLSearchParams(params).toString();
        const fullUrl = queryString ? `${url}?${queryString}` : url;
        return this.request(fullUrl, { method: 'GET' });
    },

    // POST 请求
    post(url, data = {}) {
        return this.request(url, {
            method: 'POST',
            body: JSON.stringify(data)
        });
    },

    // PATCH 请求
    patch(url, data = {}) {
        return this.request(url, {
            method: 'PATCH',
            body: JSON.stringify(data)
        });
    },

    // DELETE 请求
    delete(url) {
        return this.request(url, { method: 'DELETE' });
    }
};

// 集合 API
const collectionAPI = {
    // 获取记录列表
    list(collection, params = {}) {
        return api.get(`/collections/${collection}/records`, params);
    },

    // 获取单条记录
    get(collection, id) {
        return api.get(`/collections/${collection}/records/${id}`);
    },

    // 创建记录
    create(collection, data) {
        return api.post(`/collections/${collection}/records`, data);
    },

    // 更新记录
    update(collection, id, data) {
        return api.patch(`/collections/${collection}/records/${id}`, data);
    },

    // 删除记录
    delete(collection, id) {
        return api.delete(`/collections/${collection}/records/${id}`);
    }
};

// 认证 API
const authAPI = {
    // 注册
    register(collection, data) {
        return api.post(`/collections/${collection}/register`, data);
    },

    // 登录
    login(collection, identity, password) {
        return api.post(`/collections/${collection}/auth-with-password`, {
            identity,
            password
        });
    },

    // 刷新 Token
    refreshToken(collection, refreshToken) {
        return api.post(`/collections/${collection}/auth-refresh`, {
            refreshToken
        });
    },

    // 请求验证码
    requestOTP(collection, identity, type) {
        return api.post(`/collections/${collection}/request-otp`, {
            identity,
            type
        });
    },

    // 重置密码
    resetPassword(collection, identity, code, password) {
        return api.post(`/collections/${collection}/reset-password`, {
            identity,
            code,
            password
        });
    }
};

// 事务 API - 用于原子性多步骤操作
// 事务现在是集合的一种类型，通过集合名称执行
const transactionAPI = {
    // 执行事务（事务集合名称，参数）
    execute(collection, params = {}) {
        return api.post(`/collections/${collection}/execute`, {
            params
        });
    }
};
