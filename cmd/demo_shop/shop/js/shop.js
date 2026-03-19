// 商城功能

// 当前商品数据
let currentProduct = null;
let currentQuantity = 1;

// 当前筛选条件
let currentFilters = {
    category: '',
    search: '',
    status: '',
    page: 1,
    perPage: 12
};

// ==================== 商品相关 ====================

// 加载商品分类
async function loadCategories() {
    try {
        const result = await collectionAPI.list('categories', {
            filter: "status = 'active'",
            sort: 'sort',
            perPage: 100
        });
        
        const categoryList = document.getElementById('categoryList');
        if (categoryList && result.data.items) {
            categoryList.innerHTML = result.data.items.map(cat => 
                `<button class="category-btn" onclick="filterByCategory('${cat.id}')">${cat.name}</button>`
            ).join('');
        }
    } catch (error) {
        console.error('加载分类失败:', error);
    }
}

// 加载商品列表
async function loadProducts() {
    const productsList = document.getElementById('productsList');
    if (!productsList) return;
    
    productsList.innerHTML = '<div class="loading">加载中...</div>';
    
    try {
        const params = {
            sort: '-created',
            perPage: currentFilters.perPage,
            page: currentFilters.page
        };
        
        // 构建过滤条件
        let filters = ["status = 'active'", 'stock > 0'];
        if (currentFilters.category) {
            filters.push(`category = '${currentFilters.category}'`);
        }
        if (currentFilters.search) {
            filters.push(`name ~ '${currentFilters.search}'`);
        }
        params.filter = filters.join(' && ');
        
        const result = await collectionAPI.list('products', params);
        
        if (result.data.items && result.data.items.length > 0) {
            productsList.innerHTML = result.data.items.map(product => `
                <div class="product-card" onclick="goToProduct(${product.id})">
                    <div class="product-image">
                        ${product.image ? `<img src="${product.image}" alt="${product.name}">` : '📦'}
                    </div>
                    <div class="product-info">
                        <div class="product-name">${product.name}</div>
                        <div class="product-price">
                            ${formatPrice(product.price)}
                            ${product.originalPrice ? `<small>${formatPrice(product.originalPrice)}</small>` : ''}
                        </div>
                        <div class="product-stock">库存: ${product.stock}</div>
                    </div>
                </div>
            `).join('');
            
            renderPagination(result.data.totalPages, result.data.page);
        } else {
            productsList.innerHTML = '<div class="empty-state"><p>暂无商品</p></div>';
        }
    } catch (error) {
        console.error('加载商品失败:', error);
        productsList.innerHTML = '<div class="empty-state"><p>加载失败，请稍后重试</p></div>';
    }
}

// 搜索商品
function searchProducts() {
    const searchInput = document.getElementById('searchInput');
    if (searchInput) {
        currentFilters.search = searchInput.value.trim();
        currentFilters.page = 1;
        loadProducts();
    }
}

// 按分类筛选
function filterByCategory(categoryId) {
    currentFilters.category = categoryId;
    currentFilters.page = 1;
    
    // 更新按钮状态
    document.querySelectorAll('.category-btn').forEach(btn => {
        btn.classList.remove('active');
    });
    event.target.classList.add('active');
    
    loadProducts();
}

// 跳转到商品详情
function goToProduct(id) {
    window.location.href = `/shop/product.html?id=${id}`;
}

// 加载商品详情
async function loadProductDetail() {
    const productId = getUrlParam('id');
    if (!productId) {
        window.location.href = '/shop';
        return;
    }
    
    const productDetail = document.getElementById('productDetail');
    const productActions = document.getElementById('productActions');
    
    try {
        const result = await collectionAPI.get('products', productId);
        currentProduct = result.data;
        
        productDetail.innerHTML = `
            <div class="product-detail-grid">
                <div class="product-detail-image">
                    ${currentProduct.image ? `<img src="${currentProduct.image}" alt="${currentProduct.name}">` : '📦'}
                </div>
                <div class="product-detail-info">
                    <h1>${currentProduct.name}</h1>
                    <div class="product-detail-price">${formatPrice(currentProduct.price)}</div>
                    <div class="product-detail-desc">${currentProduct.description || '暂无描述'}</div>
                    <div class="product-meta">
                        <span>库存: ${currentProduct.stock}</span>
                        <span>销量: ${currentProduct.sales || 0}</span>
                    </div>
                </div>
            </div>
        `;
        
        document.getElementById('stockInfo').textContent = `库存: ${currentProduct.stock}`;
        updateTotal();
        
        productActions.style.display = 'flex';
    } catch (error) {
        console.error('加载商品详情失败:', error);
        productDetail.innerHTML = '<div class="empty-state"><p>商品不存在或已下架</p></div>';
    }
}

// 修改数量
function changeQuantity(delta) {
    const quantityInput = document.getElementById('quantity');
    let newQuantity = parseInt(quantityInput.value) + delta;
    
    if (newQuantity < 1) newQuantity = 1;
    if (currentProduct && newQuantity > currentProduct.stock) {
        newQuantity = currentProduct.stock;
    }
    
    quantityInput.value = newQuantity;
    currentQuantity = newQuantity;
    updateTotal();
}

// 更新总价
function updateTotal() {
    if (!currentProduct) return;
    
    const quantityInput = document.getElementById('quantity');
    currentQuantity = parseInt(quantityInput.value) || 1;
    
    const total = currentProduct.price * currentQuantity;
    document.getElementById('totalPrice').textContent = formatPrice(total);
}

// 立即购买
function buyNow() {
    if (!isLoggedIn()) {
        window.location.href = `/shop/login.html?from=${encodeURIComponent(window.location.href)}`;
        return;
    }
    
    if (!currentProduct) return;
    
    const orderPreview = document.getElementById('orderPreview');
    const total = currentProduct.price * currentQuantity;
    
    orderPreview.innerHTML = `
        <div class="order-item">
            <div class="order-item-image">
                ${currentProduct.image ? `<img src="${currentProduct.image}">` : '📦'}
            </div>
            <div class="order-item-info">
                <div class="order-item-name">${currentProduct.name}</div>
                <div class="order-item-price">${formatPrice(currentProduct.price)} x ${currentQuantity}</div>
            </div>
        </div>
        <div style="text-align: right; margin-top: 15px; font-size: 16px;">
            合计: <span style="color: #ff4d4f; font-weight: bold;">${formatPrice(total)}</span>
        </div>
    `;
    
    // 填充默认地址
    const user = getLocalUser();
    if (user && user.address) {
        document.getElementById('orderAddress').value = user.address;
    }
    
    document.getElementById('orderModal').classList.add('show');
}

// 关闭订单弹窗
function closeOrderModal() {
    document.getElementById('orderModal').classList.remove('show');
}

// 确认下单
async function confirmOrder() {
    if (!currentProduct || !isLoggedIn()) return;
    
    const address = document.getElementById('orderAddress').value.trim();
    const remark = document.getElementById('orderRemark').value.trim();
    
    if (!address) {
        alert('请填写收货地址');
        return;
    }
    
    const user = getLocalUser();
    const total = currentProduct.price * currentQuantity;
    
    try {
        // 创建订单
        const orderData = {
            orderNo: generateOrderNo(),
            memberId: user.id,
            productId: currentProduct.id,
            productName: currentProduct.name,
            productImage: currentProduct.image || '',
            productPrice: currentProduct.price,
            quantity: currentQuantity,
            totalAmount: total,
            status: 'pending',
            address: address,
            remark: remark
        };
        
        await collectionAPI.create('orders', orderData);
        
        // 减少库存
        await collectionAPI.update('products', currentProduct.id, {
            stock: currentProduct.stock - currentQuantity,
            sales: (currentProduct.sales || 0) + currentQuantity
        });
        
        closeOrderModal();
        alert('下单成功！请前往订单页面支付');
        window.location.href = '/shop/orders.html';
    } catch (error) {
        alert(error.message || '下单失败，请重试');
    }
}

// ==================== 订单相关 ====================

// 加载订单列表
async function loadOrders() {
    const ordersList = document.getElementById('ordersList');
    if (!ordersList) return;
    
    ordersList.innerHTML = '<div class="loading">加载中...</div>';
    
    const user = getLocalUser();
    if (!user) return;
    
    try {
        const params = {
            filter: `memberId = '${user.id}'`,
            sort: '-created',
            perPage: 20
        };
        
        if (currentFilters.status) {
            params.filter += ` && status = '${currentFilters.status}'`;
        }
        
        const result = await collectionAPI.list('orders', params);
        
        if (result.data.items && result.data.items.length > 0) {
            ordersList.innerHTML = result.data.items.map(order => `
                <div class="order-card">
                    <div class="order-header">
                        <div class="order-number">订单号: ${order.orderNo}</div>
                        <div class="order-status ${getOrderStatusClass(order.status)}">${getOrderStatusText(order.status)}</div>
                    </div>
                    <div class="order-items">
                        <div class="order-item">
                            <div class="order-item-image">
                                ${order.productImage ? `<img src="${order.productImage}">` : '📦'}
                            </div>
                            <div class="order-item-info">
                                <div class="order-item-name">${order.productName}</div>
                                <div class="order-item-price">${formatPrice(order.productPrice)} x ${order.quantity}</div>
                            </div>
                        </div>
                    </div>
                    <div class="order-footer">
                        <div class="order-total">合计: <span>${formatPrice(order.totalAmount)}</span></div>
                        <div class="order-actions">
                            ${getOrderActions(order)}
                        </div>
                    </div>
                </div>
            `).join('');
        } else {
            ordersList.innerHTML = '<div class="empty-state"><p>暂无订单</p><a href="index.html" class="btn-primary">去购物</a></div>';
        }
    } catch (error) {
        console.error('加载订单失败:', error);
        ordersList.innerHTML = '<div class="empty-state"><p>加载失败，请稍后重试</p></div>';
    }
}

// 获取订单操作按钮
function getOrderActions(order) {
    let actions = `<button class="btn-secondary" onclick="viewOrderDetail(${order.id})">详情</button>`;
    
    switch (order.status) {
        case 'pending':
            actions += `<button class="btn-primary" onclick="payOrder(${order.id})">支付</button>`;
            actions += `<button class="btn-danger" onclick="cancelOrder(${order.id})">取消</button>`;
            break;
        case 'shipped':
            actions += `<button class="btn-success" onclick="confirmReceive(${order.id})">确认收货</button>`;
            break;
    }
    
    return actions;
}

// 筛选订单
function filterOrders(status) {
    currentFilters.status = status;
    
    // 更新按钮状态
    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.classList.remove('active');
    });
    event.target.classList.add('active');
    
    loadOrders();
}

// 查看订单详情
async function viewOrderDetail(orderId) {
    try {
        const result = await collectionAPI.get('orders', orderId);
        const order = result.data;
        
        const content = document.getElementById('orderDetailContent');
        content.innerHTML = `
            <div class="order-detail-section">
                <p><strong>订单号:</strong> ${order.orderNo}</p>
                <p><strong>状态:</strong> <span class="order-status ${getOrderStatusClass(order.status)}">${getOrderStatusText(order.status)}</span></p>
                <p><strong>创建时间:</strong> ${formatDateTime(order.created)}</p>
            </div>
            <div class="order-detail-section">
                <h4>商品信息</h4>
                <div class="order-item">
                    <div class="order-item-image">
                        ${order.productImage ? `<img src="${order.productImage}">` : '📦'}
                    </div>
                    <div class="order-item-info">
                        <div class="order-item-name">${order.productName}</div>
                        <div class="order-item-price">${formatPrice(order.productPrice)} x ${order.quantity}</div>
                    </div>
                </div>
            </div>
            <div class="order-detail-section">
                <h4>收货信息</h4>
                <p><strong>地址:</strong> ${order.address}</p>
                ${order.remark ? `<p><strong>备注:</strong> ${order.remark}</p>` : ''}
            </div>
            <div class="order-detail-section">
                <p style="font-size: 18px;"><strong>订单金额:</strong> <span style="color: #ff4d4f;">${formatPrice(order.totalAmount)}</span></p>
            </div>
        `;
        
        document.getElementById('orderDetailModal').classList.add('show');
    } catch (error) {
        alert('获取订单详情失败');
    }
}

// 关闭订单详情弹窗
function closeOrderDetailModal() {
    document.getElementById('orderDetailModal').classList.remove('show');
}

// 支付订单
async function payOrder(orderId) {
    if (!confirm('确认使用余额支付？')) return;
    
    try {
        // 使用事务 API 进行支付（原子操作）
        const result = await transactionAPI.execute('payment', {
            orderId: orderId
        });
        
        alert(result.message || '支付成功！');
        loadOrders();
    } catch (error) {
        alert(error.message || '支付失败');
    }
}

// 取消订单
async function cancelOrder(orderId) {
    if (!confirm('确认取消订单？')) return;
    
    try {
        // 获取订单信息
        const orderResult = await collectionAPI.get('orders', orderId);
        const order = orderResult.data;
        
        // 恢复库存
        const productResult = await collectionAPI.get('products', order.productId);
        const product = productResult.data;
        
        await collectionAPI.update('products', order.productId, {
            stock: product.stock + order.quantity,
            sales: Math.max(0, (product.sales || 0) - order.quantity)
        });
        
        // 更新订单状态
        await collectionAPI.update('orders', orderId, {
            status: 'cancelled'
        });
        
        alert('订单已取消');
        loadOrders();
    } catch (error) {
        alert(error.message || '取消失败');
    }
}

// 确认收货
async function confirmReceive(orderId) {
    if (!confirm('确认已收到商品？')) return;
    
    try {
        await collectionAPI.update('orders', orderId, {
            status: 'completed'
        });
        
        alert('已确认收货');
        loadOrders();
    } catch (error) {
        alert(error.message || '操作失败');
    }
}

// ==================== 钱包相关 ====================

// 加载钱包
async function loadWallet() {
    const balanceEl = document.getElementById('balanceAmount');
    if (!balanceEl) return;
    
    const user = getLocalUser();
    if (!user) return;
    
    try {
        const result = await collectionAPI.list('wallets', {
            filter: `memberId = '${user.id}'`
        });
        
        if (result.data.items && result.data.items.length > 0) {
            balanceEl.textContent = formatPrice(result.data.items[0].balance);
        } else {
            // 创建钱包
            await collectionAPI.create('wallets', {
                memberId: user.id,
                balance: 0
            });
            balanceEl.textContent = formatPrice(0);
        }
    } catch (error) {
        console.error('加载钱包失败:', error);
    }
}

// 加载交易记录
async function loadTransactions() {
    const transactionsList = document.getElementById('transactionsList');
    if (!transactionsList) return;
    
    transactionsList.innerHTML = '<div class="loading">加载中...</div>';
    
    const user = getLocalUser();
    if (!user) return;
    
    try {
        const params = {
            filter: `memberId = '${user.id}'`,
            sort: '-created',
            perPage: 50
        };
        
        if (currentFilters.transactionType) {
            params.filter += ` && type = '${currentFilters.transactionType}'`;
        }
        
        const result = await collectionAPI.list('transactions', params);
        
        if (result.data.items && result.data.items.length > 0) {
            transactionsList.innerHTML = result.data.items.map(tran => `
                <div class="transaction-card">
                    <div class="transaction-info">
                        <div class="transaction-type">${getTransactionTypeText(tran.type)}</div>
                        <div class="transaction-time">${formatDateTime(tran.created)}</div>
                        <div style="font-size: 12px; color: #999;">${tran.description || ''}</div>
                    </div>
                    <div class="transaction-amount ${tran.amount >= 0 ? 'amount-positive' : 'amount-negative'}">
                        ${tran.amount >= 0 ? '+' : ''}${formatPrice(tran.amount)}
                    </div>
                </div>
            `).join('');
        } else {
            transactionsList.innerHTML = '<div class="empty-state"><p>暂无交易记录</p></div>';
        }
    } catch (error) {
        console.error('加载交易记录失败:', error);
        transactionsList.innerHTML = '<div class="empty-state"><p>加载失败</p></div>';
    }
}

// 筛选交易记录
function filterTransactions(type) {
    currentFilters.transactionType = type;
    
    document.querySelectorAll('.transaction-tabs .tab-btn').forEach(btn => {
        btn.classList.remove('active');
    });
    event.target.classList.add('active');
    
    loadTransactions();
}

// 显示充值弹窗
function showRechargeModal() {
    document.getElementById('rechargeModal').classList.add('show');
}

// 关闭充值弹窗
function closeRechargeModal() {
    document.getElementById('rechargeModal').classList.remove('show');
    document.getElementById('customAmount').value = '';
}

// 选择充值金额
function selectAmount(amount) {
    document.getElementById('customAmount').value = amount;
    
    document.querySelectorAll('.amount-btn').forEach(btn => {
        btn.classList.remove('active');
    });
    event.target.classList.add('active');
}

// 确认充值
async function confirmRecharge() {
    const customAmount = document.getElementById('customAmount').value;
    const amount = parseFloat(customAmount);
    
    if (!amount || amount <= 0) {
        alert('请输入有效的充值金额');
        return;
    }
    
    try {
        // 使用事务 API 进行充值（原子操作）
        const result = await transactionAPI.execute('recharge', {
            amount: amount
        });
        
        closeRechargeModal();
        alert(result.message || '充值成功！');
        loadWallet();
        loadTransactions();
    } catch (error) {
        alert(error.message || '充值失败');
    }
}

// ==================== 个人中心 ====================

// 加载个人信息
async function loadProfile() {
    const user = getLocalUser();
    if (!user) return;
    
    try {
        // 获取最新用户信息
        const result = await collectionAPI.get('members', user.id);
        const member = result.data;
        
        // 更新本地存储
        localStorage.setItem('user', JSON.stringify(member));
        
        // 更新页面显示
        document.getElementById('profileNickname').textContent = member.nickname || member.email || '用户';
        document.getElementById('profileEmail').textContent = member.email || member.mobile || '';
        
        // 填充表单
        document.getElementById('nickname').value = member.nickname || '';
        document.getElementById('phone').value = member.phone || '';
        document.getElementById('address').value = member.address || '';
    } catch (error) {
        console.error('加载个人信息失败:', error);
    }
}

// 更新个人信息
async function updateProfile(event) {
    event.preventDefault();
    
    const user = getLocalUser();
    if (!user) return;
    
    const nickname = document.getElementById('nickname').value.trim();
    const phone = document.getElementById('phone').value.trim();
    const address = document.getElementById('address').value.trim();
    
    try {
        await collectionAPI.update('members', user.id, {
            nickname,
            phone,
            address
        });
        
        // 更新本地存储
        const updatedUser = { ...user, nickname, phone, address };
        localStorage.setItem('user', JSON.stringify(updatedUser));
        
        showMessage('message', '保存成功', 'success');
        loadProfile();
    } catch (error) {
        showMessage('message', error.message || '保存失败', 'error');
    }
}

// 修改密码
async function changePassword(event) {
    event.preventDefault();
    
    const oldPassword = document.getElementById('oldPassword').value;
    const newPassword = document.getElementById('newPassword').value;
    const confirmNewPassword = document.getElementById('confirmNewPassword').value;
    
    if (newPassword.length < 6) {
        showMessage('message', '新密码至少6位', 'error');
        return;
    }
    
    if (newPassword !== confirmNewPassword) {
        showMessage('message', '两次密码不一致', 'error');
        return;
    }
    
    // 这里需要后端支持修改密码的接口
    // 暂时提示
    showMessage('message', '密码修改功能需要后端支持', 'error');
}

// ==================== 分页 ====================

// 渲染分页
function renderPagination(totalPages, currentPage) {
    const pagination = document.getElementById('pagination');
    if (!pagination || totalPages <= 1) {
        if (pagination) pagination.innerHTML = '';
        return;
    }
    
    let html = '';
    
    // 上一页
    html += `<button ${currentPage <= 1 ? 'disabled' : ''} onclick="goToPage(${currentPage - 1})">上一页</button>`;
    
    // 页码
    for (let i = 1; i <= totalPages; i++) {
        if (i === 1 || i === totalPages || (i >= currentPage - 2 && i <= currentPage + 2)) {
            html += `<button class="${i === currentPage ? 'active' : ''}" onclick="goToPage(${i})">${i}</button>`;
        } else if (i === currentPage - 3 || i === currentPage + 3) {
            html += '<button disabled>...</button>';
        }
    }
    
    // 下一页
    html += `<button ${currentPage >= totalPages ? 'disabled' : ''} onclick="goToPage(${currentPage + 1})">下一页</button>`;
    
    pagination.innerHTML = html;
}

// 跳转页面
function goToPage(page) {
    currentFilters.page = page;
    loadProducts();
    window.scrollTo(0, 0);
}
