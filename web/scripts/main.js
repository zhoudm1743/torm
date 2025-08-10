// 创建TORM命名空间
window.TORM = window.TORM || {};

// 主要JavaScript功能
document.addEventListener('DOMContentLoaded', function() {
    initNavigation();
    initTabSwitching();
    initCopyFunction();
    initScrollEffects();
    initLanguageSelector();
});

// 导航功能
function initNavigation() {
    const navToggle = document.getElementById('nav-toggle');
    const navMenu = document.getElementById('nav-menu');
    
    if (navToggle && navMenu) {
        navToggle.addEventListener('click', function() {
            navMenu.classList.toggle('active');
            navToggle.classList.toggle('active');
        });
        
        // 点击导航链接时关闭菜单
        navMenu.addEventListener('click', function(e) {
            if (e.target.classList.contains('nav-link')) {
                navMenu.classList.remove('active');
                navToggle.classList.remove('active');
            }
        });
        
        // 点击外部区域时关闭菜单
        document.addEventListener('click', function(e) {
            if (!navToggle.contains(e.target) && !navMenu.contains(e.target)) {
                navMenu.classList.remove('active');
                navToggle.classList.remove('active');
            }
        });
    }
    
    // 滚动时导航栏透明度变化
    window.addEventListener('scroll', function() {
        const navbar = document.querySelector('.navbar');
        if (navbar) {
            if (window.scrollY > 50) {
                navbar.style.background = 'rgba(255, 255, 255, 0.95)';
                navbar.style.borderBottomColor = 'rgba(255, 255, 255, 0.3)';
            } else {
                navbar.style.background = 'rgba(255, 255, 255, 0.25)';
                navbar.style.borderBottomColor = 'rgba(255, 255, 255, 0.18)';
            }
        }
    });
}

// 标签切换功能
function initTabSwitching() {
    const tabButtons = document.querySelectorAll('.tab-btn');
    
    tabButtons.forEach(button => {
        button.addEventListener('click', function() {
            const targetTab = this.getAttribute('data-tab');
            if (targetTab) {
                showTab(targetTab);
            }
        });
    });
}

function showTab(tabName) {
    // 隐藏所有标签内容
    document.querySelectorAll('.tab-content').forEach(content => {
        content.classList.remove('active');
    });
    
    // 移除所有按钮的活动状态
    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.classList.remove('active');
    });
    
    // 显示目标标签内容
    const targetContent = document.getElementById(tabName);
    if (targetContent) {
        targetContent.classList.add('active');
    }
    
    // 激活对应按钮
    const targetButton = document.querySelector(`[data-tab="${tabName}"]`);
    if (targetButton) {
        targetButton.classList.add('active');
    }
}

// 添加到TORM命名空间
TORM.showTab = showTab;
window.showTab = showTab; // 保持向后兼容

// 复制功能
function initCopyFunction() {
    const copyButton = document.getElementById('copy-install-btn');
    if (copyButton) {
        copyButton.addEventListener('click', copyInstallCommand);
    }
}

function copyInstallCommand() {
    const command = document.querySelector('.install-command').textContent;
    
    if (navigator.clipboard) {
        navigator.clipboard.writeText(command).then(() => {
            showCopyFeedback();
        }).catch(err => {
            console.error('复制失败:', err);
            fallbackCopy(command);
        });
    } else {
        fallbackCopy(command);
    }
}

// 添加到TORM命名空间
TORM.copyInstallCommand = copyInstallCommand;
window.copyInstallCommand = copyInstallCommand; // 保持向后兼容

function fallbackCopy(text) {
    const textArea = document.createElement('textarea');
    textArea.value = text;
    textArea.style.position = 'fixed';
    textArea.style.left = '-999999px';
    textArea.style.top = '-999999px';
    document.body.appendChild(textArea);
    textArea.focus();
    textArea.select();
    
    try {
        document.execCommand('copy');
        showCopyFeedback();
    } catch (err) {
        console.error('复制失败:', err);
    }
    
    document.body.removeChild(textArea);
}

function showCopyFeedback() {
    const copyBtn = document.getElementById('copy-install-btn');
    if (!copyBtn) return;
    
    const originalText = copyBtn.textContent;
    
    copyBtn.textContent = '已复制!';
    copyBtn.style.background = 'var(--accent-green)';
    
    setTimeout(() => {
        copyBtn.textContent = originalText;
        copyBtn.style.background = 'var(--primary-green)';
    }, 2000);
}

// 滚动效果
function initScrollEffects() {
    // 平滑滚动到锚点
    document.querySelectorAll('a[href^="#"]').forEach(anchor => {
        anchor.addEventListener('click', function (e) {
            const href = this.getAttribute('href');
            // 跳过空锚点或只有#的链接
            if (!href || href === '#' || href.length <= 1) {
                return;
            }
            
            e.preventDefault();
            const target = document.querySelector(href);
            if (target) {
                target.scrollIntoView({
                    behavior: 'smooth',
                    block: 'start'
                });
            }
        });
    });
    
    // 滚动时的元素动画
    const observer = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                entry.target.style.opacity = '1';
                entry.target.style.transform = 'translateY(0)';
            }
        });
    }, {
        threshold: 0.1,
        rootMargin: '0px 0px -50px 0px'
    });
    
    // 观察需要动画的元素
    document.querySelectorAll('.feature-card, .sponsor-card').forEach(el => {
        el.style.opacity = '0';
        el.style.transform = 'translateY(30px)';
        el.style.transition = 'opacity 0.6s ease, transform 0.6s ease';
        observer.observe(el);
    });
}

// 语言选择器
function initLanguageSelector() {
    const languageSelect = document.getElementById('language-select');
    if (languageSelect) {
        languageSelect.addEventListener('change', function() {
            const selectedLang = this.value;
            // 这里可以添加语言切换逻辑
            console.log('切换语言到:', selectedLang);
            
            // 临时提示功能
            showNotification('语言切换功能正在开发中');
        });
    }
}

// 通知功能
function showNotification(message, type = 'info') {
    // 创建通知元素
    const notification = document.createElement('div');
    notification.className = `notification notification-${type}`;
    notification.textContent = message;
    
    // 样式
    notification.style.cssText = `
        position: fixed;
        top: 5rem;
        right: 1rem;
        background: var(--glass-backdrop);
        backdrop-filter: blur(16px);
        -webkit-backdrop-filter: blur(16px);
        border: 1px solid var(--glass-border);
        border-radius: 0.75rem;
        padding: 1rem 1.5rem;
        color: var(--text-primary);
        font-weight: 500;
        box-shadow: var(--shadow-lg);
        z-index: 1001;
        opacity: 0;
        transform: translateX(100%);
        transition: all 0.3s ease;
    `;
    
    document.body.appendChild(notification);
    
    // 显示动画
    requestAnimationFrame(() => {
        notification.style.opacity = '1';
        notification.style.transform = 'translateX(0)';
    });
    
    // 自动隐藏
    setTimeout(() => {
        notification.style.opacity = '0';
        notification.style.transform = 'translateX(100%)';
        setTimeout(() => {
            document.body.removeChild(notification);
        }, 300);
    }, 3000);
}

// 工具函数
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

function throttle(func, limit) {
    let inThrottle;
    return function() {
        const args = arguments;
        const context = this;
        if (!inThrottle) {
            func.apply(context, args);
            inThrottle = true;
            setTimeout(() => inThrottle = false, limit);
        }
    }
}

// 响应式样式已移至main.css文件中，无需动态添加

// GitHub统计获取（可选功能）
async function fetchGitHubStats() {
    try {
        const response = await fetch('https://api.github.com/repos/zhoudm1743/torm');
        const data = await response.json();
        
        // 更新统计数据
        const statItems = document.querySelectorAll('.stat-item');
        if (statItems.length >= 3) {
            statItems[0].querySelector('.stat-text').textContent = `${data.stargazers_count} stars`;
            statItems[1].querySelector('.stat-text').textContent = `${data.forks_count} forks`;
            // 贡献者数据需要额外请求
        }
    } catch (error) {
        console.log('GitHub统计获取失败:', error);
    }
}

// 页面加载完成后获取GitHub统计
window.addEventListener('load', () => {
    fetchGitHubStats();
});

// 键盘导航支持
document.addEventListener('keydown', function(e) {
    // ESC键关闭菜单
    if (e.key === 'Escape') {
        const navMenu = document.getElementById('nav-menu');
        const navToggle = document.getElementById('nav-toggle');
        if (navMenu && navToggle) {
            navMenu.classList.remove('active');
            navToggle.classList.remove('active');
        }
    }
    
    // Tab键切换代码示例
    if (e.key === 'Tab' && e.ctrlKey) {
        e.preventDefault();
        const activeTab = document.querySelector('.tab-btn.active');
        const allTabs = document.querySelectorAll('.tab-btn');
        const currentIndex = Array.from(allTabs).indexOf(activeTab);
        const nextIndex = (currentIndex + 1) % allTabs.length;
        allTabs[nextIndex].click();
    }
});

// 性能优化：节流滚动事件
window.addEventListener('scroll', throttle(function() {
    // 滚动相关的性能敏感操作
}, 16)); // 约60fps

// 标记脚本已加载
if (window.TORM) {
    TORM.markScriptLoaded('main.js');
} 