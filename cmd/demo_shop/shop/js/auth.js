// 认证相关功能

// 检查登录状态
function checkAuth(requireAuth = false) {
    const user = getLocalUser();
    const token = localStorage.getItem('token');
    
    const userNameEl = document.getElementById('userName');
    const loginBtn = document.getElementById('loginBtn');
    const logoutBtn = document.getElementById('logoutBtn');
    
    if (token && user) {
        if (userNameEl) userNameEl.textContent = user.nickname || user.email || '用户';
        if (loginBtn) loginBtn.style.display = 'none';
        if (logoutBtn) logoutBtn.style.display = 'inline';
    } else {
        if (userNameEl) userNameEl.textContent = '未登录';
        if (loginBtn) loginBtn.style.display = 'inline';
        if (logoutBtn) logoutBtn.style.display = 'none';
        
        if (requireAuth) {
            window.location.href = '/shop/login.html';
        }
    }
}

// 退出登录
function logout() {
    clearUserInfo();
    window.location.href = '/shop';
}

// 处理登录
async function handleLogin(event) {
    event.preventDefault();
    
    const identity = document.getElementById('identity').value.trim();
    const password = document.getElementById('password').value;
    
    if (!identity || !password) {
        showMessage('message', '请填写完整信息', 'error');
        return;
    }
    
    try {
        const result = await authAPI.login('members', identity, password);
        saveUserInfo(result.data);
        
        showMessage('message', '登录成功！', 'success');
        
        setTimeout(() => {
            // 检查是否有来源页面
            const from = getUrlParam('from');
            if (from) {
                window.location.href = decodeURIComponent(from);
            } else {
                window.location.href = '/shop';
            }
        }, 1000);
    } catch (error) {
        showMessage('message', error.message || '登录失败', 'error');
    }
}

// 处理注册
async function handleRegister(event) {
    event.preventDefault();
    
    const identity = document.getElementById('identity').value.trim();
    const code = document.getElementById('code').value.trim();
    const password = document.getElementById('password').value;
    const confirmPassword = document.getElementById('confirmPassword').value;
    const nickname = document.getElementById('nickname').value.trim();
    
    if (!identity || !password) {
        showMessage('message', '请填写完整信息', 'error');
        return;
    }
    
    if (password.length < 6) {
        showMessage('message', '密码至少6位', 'error');
        return;
    }
    
    if (password !== confirmPassword) {
        showMessage('message', '两次密码不一致', 'error');
        return;
    }
    
    try {
        const data = {
            identity,
            password
        };
        
        if (code) data.code = code;
        if (nickname) data.nickname = nickname;
        
        const result = await authAPI.register('members', data);
        
        showMessage('message', '注册成功！请登录', 'success');
        
        setTimeout(() => {
            window.location.href = '/shop/login.html';
        }, 1500);
    } catch (error) {
        showMessage('message', error.message || '注册失败', 'error');
    }
}

// 发送注册验证码
async function sendRegisterCode() {
    const identity = document.getElementById('identity').value.trim();
    
    if (!identity) {
        showMessage('message', '请先输入邮箱或手机号', 'error');
        return;
    }
    
    try {
        await authAPI.requestOTP('members', identity, 'register');
        showMessage('message', '验证码已发送，请查收', 'success');
        startCountdown('sendCodeBtn', 60);
    } catch (error) {
        showMessage('message', error.message || '发送失败', 'error');
    }
}

// 发送重置密码验证码
async function sendResetCode() {
    const identity = document.getElementById('identity').value.trim();
    
    if (!identity) {
        showMessage('message', '请先输入邮箱或手机号', 'error');
        return;
    }
    
    try {
        await authAPI.requestOTP('members', identity, 'password-reset');
        showMessage('message', '验证码已发送，请查收', 'success');
        startCountdown('sendCodeBtn', 60);
    } catch (error) {
        showMessage('message', error.message || '发送失败', 'error');
    }
}

// 处理重置密码
async function handleResetPassword(event) {
    event.preventDefault();
    
    const identity = document.getElementById('identity').value.trim();
    const code = document.getElementById('code').value.trim();
    const password = document.getElementById('password').value;
    const confirmPassword = document.getElementById('confirmPassword').value;
    
    if (!identity || !code || !password) {
        showMessage('message', '请填写完整信息', 'error');
        return;
    }
    
    if (password.length < 6) {
        showMessage('message', '密码至少6位', 'error');
        return;
    }
    
    if (password !== confirmPassword) {
        showMessage('message', '两次密码不一致', 'error');
        return;
    }
    
    try {
        await authAPI.resetPassword('members', identity, code, password);
        
        showMessage('message', '密码重置成功！请登录', 'success');
        
        setTimeout(() => {
            window.location.href = '/shop/login.html';
        }, 1500);
    } catch (error) {
        showMessage('message', error.message || '重置失败', 'error');
    }
}
