// 工具函数

// 显示消息
function showMessage(elementId, message, type = 'error') {
    const element = document.getElementById(elementId);
    if (element) {
        element.textContent = message;
        element.className = `message ${type}`;
        element.style.display = 'block';
        
        // 3秒后自动隐藏
        setTimeout(() => {
            element.style.display = 'none';
        }, 3000);
    }
}

// 格式化价格
function formatPrice(price) {
    return '¥' + parseFloat(price || 0).toFixed(2);
}

// 格式化日期时间
function formatDateTime(dateStr) {
    if (!dateStr) return '-';
    const date = new Date(dateStr);
    return date.toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit'
    });
}

// 格式化日期
function formatDate(dateStr) {
    if (!dateStr) return '-';
    const date = new Date(dateStr);
    return date.toLocaleDateString('zh-CN');
}

// 获取 URL 参数
function getUrlParam(name) {
    const urlParams = new URLSearchParams(window.location.search);
    return urlParams.get(name);
}

// 生成订单号
function generateOrderNo() {
    const now = new Date();
    const year = now.getFullYear();
    const month = String(now.getMonth() + 1).padStart(2, '0');
    const day = String(now.getDate()).padStart(2, '0');
    const random = Math.floor(Math.random() * 10000).toString().padStart(4, '0');
    return `${year}${month}${day}${random}`;
}

// 防抖函数
function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

// 获取订单状态文本
function getOrderStatusText(status) {
    const statusMap = {
        'pending': '待支付',
        'paid': '待发货',
        'shipped': '待收货',
        'completed': '已完成',
        'cancelled': '已取消'
    };
    return statusMap[status] || status;
}

// 获取订单状态样式类
function getOrderStatusClass(status) {
    return `status-${status}`;
}

// 获取交易类型文本
function getTransactionTypeText(type) {
    const typeMap = {
        'recharge': '充值',
        'payment': '消费',
        'refund': '退款'
    };
    return typeMap[type] || type;
}

// 本地存储用户信息
function saveUserInfo(data) {
    localStorage.setItem('token', data.token);
    localStorage.setItem('refreshToken', data.refreshToken);
    localStorage.setItem('user', JSON.stringify(data.record));
}

// 获取本地用户信息
function getLocalUser() {
    const userStr = localStorage.getItem('user');
    return userStr ? JSON.parse(userStr) : null;
}

// 清除用户信息
function clearUserInfo() {
    localStorage.removeItem('token');
    localStorage.removeItem('refreshToken');
    localStorage.removeItem('user');
}

// 检查是否登录
function isLoggedIn() {
    return !!localStorage.getItem('token');
}

// 倒计时
function startCountdown(elementId, seconds, callback) {
    const element = document.getElementById(elementId);
    let remaining = seconds;
    
    element.disabled = true;
    element.textContent = `${remaining}秒后重试`;
    
    const timer = setInterval(() => {
        remaining--;
        if (remaining <= 0) {
            clearInterval(timer);
            element.disabled = false;
            element.textContent = '发送验证码';
            if (callback) callback();
        } else {
            element.textContent = `${remaining}秒后重试`;
        }
    }, 1000);
}
