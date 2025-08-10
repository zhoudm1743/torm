// TORM Web 配置文件
// 确保这个文件最先加载

// 创建全局命名空间
window.TORM = window.TORM || {
    version: '1.0.0',
    debug: false,
    scripts: {
        loaded: []
    }
};

// 日志函数
TORM.log = function(message, type = 'info') {
    if (TORM.debug) {
        const prefix = `[TORM-${type.toUpperCase()}]`;
        console.log(prefix, message);
    }
};

// 脚本加载追踪
TORM.markScriptLoaded = function(scriptName) {
    TORM.scripts.loaded.push(scriptName);
    TORM.log(`Script loaded: ${scriptName}`);
};

// 检查脚本是否已加载
TORM.isScriptLoaded = function(scriptName) {
    return TORM.scripts.loaded.includes(scriptName);
};

// 工具函数检查
TORM.checkUtilities = function() {
    const utilities = ['debounce', 'throttle'];
    utilities.forEach(util => {
        if (typeof window[util] !== 'function') {
            TORM.log(`Utility function ${util} not found`, 'warn');
        }
    });
};

// 页面准备就绪时的回调队列
TORM.readyCallbacks = [];
TORM.isReady = false;

TORM.ready = function(callback) {
    if (TORM.isReady) {
        callback();
    } else {
        TORM.readyCallbacks.push(callback);
    }
};

// 标记页面已准备就绪
TORM.markReady = function() {
    TORM.isReady = true;
    TORM.readyCallbacks.forEach(callback => callback());
    TORM.readyCallbacks = [];
    TORM.log('TORM system ready');
};

// 错误处理
TORM.handleError = function(error, context = 'unknown') {
    console.error(`[TORM-ERROR] ${context}:`, error);
    if (TORM.debug) {
        console.trace();
    }
};

// 初始化
TORM.log('TORM namespace initialized');

// 检测是否在移动设备上
TORM.isMobile = function() {
    return window.innerWidth <= 768;
};

// 检测是否支持现代特性
TORM.features = {
    backdropFilter: CSS.supports('backdrop-filter', 'blur(10px)'),
    grid: CSS.supports('display', 'grid'),
    flexbox: CSS.supports('display', 'flex'),
    intersectionObserver: 'IntersectionObserver' in window,
    clipboard: 'clipboard' in navigator
};

TORM.log('Feature detection complete', 'info');
TORM.log(TORM.features, 'debug');

// DOM ready 检测
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', TORM.markReady);
} else {
    // DOM already loaded
    setTimeout(TORM.markReady, 0);
} 