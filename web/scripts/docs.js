// 确保TORM命名空间存在
window.TORM = window.TORM || {};

// 文档系统
class DocsSystem {
    constructor() {
        this.currentDoc = 'home';
        this.docs = {
            'home': {
                title: '欢迎使用 TORM',
                file: '../wiki/Home.md',
                next: 'installation'
            },
            'installation': {
                title: '安装指南',
                file: '../wiki/Installation.md',
                prev: 'home',
                next: 'quick-start'
            },
            'quick-start': {
                title: '快速开始',
                file: '../wiki/Quick-Start.md',
                prev: 'installation',
                next: 'configuration'
            },
            'configuration': {
                title: '配置',
                file: '../wiki/Configuration.md',
                prev: 'quick-start',
                next: 'database-support'
            },
            'database-support': {
                title: '数据库支持',
                file: '../wiki/Database-Support.md',
                prev: 'configuration',
                next: 'query-builder'
            },
            'query-builder': {
                title: '查询构建器',
                file: '../wiki/Query-Builder.md',
                prev: 'database-support',
                next: 'model-system'
            },
            'model-system': {
                title: '模型系统',
                file: '../wiki/Model-System.md',
                prev: 'query-builder',
                next: 'relationships'
            },
            'relationships': {
                title: '关联关系',
                file: '../wiki/Relationships.md',
                prev: 'model-system',
                next: 'transactions'
            },
            'transactions': {
                title: '事务处理',
                file: '../wiki/Transactions.md',
                prev: 'relationships',
                next: 'migrations'
            },
            'migrations': {
                title: '数据迁移',
                file: '../wiki/Migrations.md',
                prev: 'transactions',
                next: 'caching'
            },
            'caching': {
                title: '缓存系统',
                file: '../wiki/Caching.md',
                prev: 'migrations',
                next: 'logging'
            },
            'logging': {
                title: '日志系统',
                file: '../wiki/Logging.md',
                prev: 'caching',
                next: 'performance'
            },
            'performance': {
                title: '性能优化',
                file: '../wiki/Performance.md',
                prev: 'logging',
                next: 'best-practices'
            },
            'best-practices': {
                title: '最佳实践',
                file: '../wiki/Best-Practices.md',
                prev: 'performance',
                next: 'examples'
            },
            'examples': {
                title: '示例代码',
                file: '../wiki/Examples.md',
                prev: 'best-practices',
                next: 'api-reference'
            },
            'api-reference': {
                title: 'API参考',
                file: '../wiki/API-Reference.md',
                prev: 'examples',
                next: 'contributing'
            },
            'contributing': {
                title: '贡献指南',
                file: '../wiki/Contributing.md',
                prev: 'api-reference',
                next: 'troubleshooting'
            },
            'troubleshooting': {
                title: '故障排除',
                file: '../wiki/Troubleshooting.md',
                prev: 'contributing',
                next: 'changelog'
            },
            'changelog': {
                title: '更新日志',
                file: '../wiki/Changelog.md',
                prev: 'troubleshooting'
            }
        };
        
        this.init();
    }
    
    init() {
        this.initEventListeners();
        this.initBackToTop();
        this.initSidebar();
        
        // 从URL获取文档名称
        const urlParams = new URLSearchParams(window.location.search);
        const docName = urlParams.get('doc') || 'home';
        this.loadDoc(docName);
    }
    
    initEventListeners() {
        // 侧边栏切换
        const sidebarToggle = document.getElementById('sidebar-toggle');
        if (sidebarToggle) {
            sidebarToggle.addEventListener('click', () => {
                this.toggleSidebar();
            });
        }
        
        // 文档导航点击事件
        document.addEventListener('click', (e) => {
            const docLink = e.target.closest('[data-doc]');
            if (docLink) {
                e.preventDefault();
                const docName = docLink.getAttribute('data-doc');
                this.loadDoc(docName);
            }
        });
        
        // 监听窗口大小变化
        window.addEventListener('resize', debounce(() => {
            this.handleResize();
        }, 250));
    }
    
    initBackToTop() {
        const backToTopBtn = document.getElementById('back-to-top');
        if (!backToTopBtn) return;
        
        // 监听滚动事件
        window.addEventListener('scroll', throttle(() => {
            const scrollTop = window.pageYOffset || document.documentElement.scrollTop;
            
            if (scrollTop > 500) {
                backToTopBtn.classList.add('visible');
            } else {
                backToTopBtn.classList.remove('visible');
            }
        }, 100));
        
        // 点击回到顶部
        backToTopBtn.addEventListener('click', () => {
            window.scrollTo({
                top: 0,
                behavior: 'smooth'
            });
        });
    }
    
    initSidebar() {
        // 移动端点击外部关闭侧边栏
        document.addEventListener('click', (e) => {
            const sidebar = document.getElementById('docs-sidebar');
            const sidebarToggle = document.getElementById('sidebar-toggle');
            
            if (window.innerWidth <= 768 && 
                sidebar && 
                sidebar.classList.contains('active') &&
                !sidebar.contains(e.target) &&
                !sidebarToggle.contains(e.target)) {
                this.closeSidebar();
            }
        });
    }
    
    toggleSidebar() {
        const sidebar = document.getElementById('docs-sidebar');
        if (sidebar) {
            sidebar.classList.toggle('active');
        }
    }
    
    closeSidebar() {
        const sidebar = document.getElementById('docs-sidebar');
        if (sidebar) {
            sidebar.classList.remove('active');
        }
    }
    
    handleResize() {
        // 大屏幕时自动关闭侧边栏
        if (window.innerWidth > 768) {
            this.closeSidebar();
        }
    }
    
    async loadDoc(docName) {
        if (!this.docs[docName]) {
            console.error('文档不存在:', docName);
            return;
        }
        
        try {
            this.showLoading();
            this.updateActiveNav(docName);
            this.updateURL(docName);
            
            const docInfo = this.docs[docName];
            const response = await fetch(docInfo.file);
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            const markdownContent = await response.text();
            const htmlContent = this.parseMarkdown(markdownContent);
            
            this.renderContent(htmlContent, docInfo.title);
            this.generateTOC();
            this.updatePagination(docName);
            this.highlightCode();
            this.currentDoc = docName;
            
            // 移动端自动关闭侧边栏
            if (window.innerWidth <= 768) {
                this.closeSidebar();
            }
            
        } catch (error) {
            console.error('加载文档失败:', error);
            this.showError(error.message);
        }
    }
    
    showLoading() {
        const content = document.getElementById('docs-content');
        if (content) {
            content.innerHTML = `
                <div class="loading-content">
                    <div class="loading-spinner"></div>
                    <p>正在加载文档...</p>
                </div>
            `;
        }
    }
    
    showError(message) {
        const content = document.getElementById('docs-content');
        if (content) {
            content.innerHTML = `
                <div class="error-content">
                    <h1>加载失败</h1>
                    <p>抱歉，无法加载文档内容。</p>
                    <p class="error-message">${message}</p>
                    <button onclick="location.reload()" class="retry-btn">重试</button>
                </div>
            `;
        }
    }
    
    parseMarkdown(markdown) {
        // 使用 marked.js 解析 Markdown
        if (typeof marked !== 'undefined') {
            // 配置 marked
            marked.setOptions({
                highlight: function(code, lang) {
                    if (typeof Prism !== 'undefined' && lang && Prism.languages[lang]) {
                        return Prism.highlight(code, Prism.languages[lang], lang);
                    }
                    return code;
                },
                breaks: true,
                gfm: true
            });
            
            return marked.parse(markdown);
        }
        
        // 简单的 Markdown 解析（备用）
        return this.simpleMarkdownParser(markdown);
    }
    
    simpleMarkdownParser(markdown) {
        return markdown
            .replace(/^# (.*$)/gim, '<h1>$1</h1>')
            .replace(/^## (.*$)/gim, '<h2>$1</h2>')
            .replace(/^### (.*$)/gim, '<h3>$1</h3>')
            .replace(/^#### (.*$)/gim, '<h4>$1</h4>')
            .replace(/\*\*(.*)\*\*/gim, '<strong>$1</strong>')
            .replace(/\*(.*)\*/gim, '<em>$1</em>')
            .replace(/`([^`]+)`/gim, '<code>$1</code>')
            .replace(/```([^`]+)```/gim, '<pre><code>$1</code></pre>')
            .replace(/^\* (.*$)/gim, '<li>$1</li>')
            .replace(/(<li>.*<\/li>)/s, '<ul>$1</ul>')
            .replace(/\n/gim, '<br>');
    }
    
    renderContent(html, title) {
        const content = document.getElementById('docs-content');
        if (content) {
            content.innerHTML = html;
            
            // 更新页面标题
            document.title = `${title} - TORM`;
        }
    }
    
    generateTOC() {
        const content = document.getElementById('docs-content');
        const tocNav = document.getElementById('toc-nav');
        
        if (!content || !tocNav) return;
        
        const headings = content.querySelectorAll('h1, h2, h3, h4');
        if (headings.length === 0) {
            tocNav.innerHTML = '<p class="no-toc">此页面没有目录</p>';
            return;
        }
        
        let tocHTML = '<ul>';
        
        headings.forEach((heading, index) => {
            const id = `heading-${index}`;
            heading.id = id;
            
            const level = heading.tagName.toLowerCase();
            const text = heading.textContent;
            const className = `toc-${level}`;
            
            tocHTML += `<li><a href="#${id}" class="${className}">${text}</a></li>`;
        });
        
        tocHTML += '</ul>';
        tocNav.innerHTML = tocHTML;
        
        // 添加平滑滚动和高亮
        this.initTOCInteractions();
    }
    
    initTOCInteractions() {
        const tocLinks = document.querySelectorAll('.toc-nav a');
        const headings = document.querySelectorAll('#docs-content h1, #docs-content h2, #docs-content h3, #docs-content h4');
        
        // 点击目录链接时的平滑滚动
        tocLinks.forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                const targetId = link.getAttribute('href').substring(1);
                const targetElement = document.getElementById(targetId);
                
                if (targetElement) {
                    const navHeight = document.querySelector('.navbar').offsetHeight;
                    const targetPosition = targetElement.offsetTop - navHeight - 20;
                    
                    window.scrollTo({
                        top: targetPosition,
                        behavior: 'smooth'
                    });
                }
            });
        });
        
        // 滚动时高亮当前章节
        const observerOptions = {
            rootMargin: '-80px 0px -80px 0px',
            threshold: 0.1
        };
        
        const observer = new IntersectionObserver((entries) => {
            entries.forEach(entry => {
                const id = entry.target.id;
                const tocLink = document.querySelector(`.toc-nav a[href="#${id}"]`);
                
                if (entry.isIntersecting) {
                    // 移除所有活动状态
                    document.querySelectorAll('.toc-nav a').forEach(link => {
                        link.classList.remove('active');
                    });
                    
                    // 添加当前章节的活动状态
                    if (tocLink) {
                        tocLink.classList.add('active');
                    }
                }
            });
        }, observerOptions);
        
        headings.forEach(heading => {
            observer.observe(heading);
        });
    }
    
    updateActiveNav(docName) {
        // 更新侧边栏导航的活动状态
        document.querySelectorAll('.nav-item').forEach(item => {
            item.classList.remove('active');
        });
        
        const activeItem = document.querySelector(`[data-doc="${docName}"]`);
        if (activeItem) {
            activeItem.classList.add('active');
        }
    }
    
    updateURL(docName) {
        const newURL = new URL(window.location);
        newURL.searchParams.set('doc', docName);
        window.history.pushState({}, '', newURL);
    }
    
    updatePagination(docName) {
        const pagination = document.getElementById('docs-pagination');
        const prevPage = document.getElementById('prev-page');
        const nextPage = document.getElementById('next-page');
        
        if (!pagination || !prevPage || !nextPage) return;
        
        const docInfo = this.docs[docName];
        
        // 更新上一页
        if (docInfo.prev) {
            const prevDoc = this.docs[docInfo.prev];
            prevPage.style.display = 'block';
            prevPage.onclick = () => this.loadDoc(docInfo.prev);
            prevPage.querySelector('.pagination-title').textContent = prevDoc.title;
        } else {
            prevPage.style.display = 'none';
        }
        
        // 更新下一页
        if (docInfo.next) {
            const nextDoc = this.docs[docInfo.next];
            nextPage.style.display = 'block';
            nextPage.onclick = () => this.loadDoc(docInfo.next);
            nextPage.querySelector('.pagination-title').textContent = nextDoc.title;
        } else {
            nextPage.style.display = 'none';
        }
        
        // 显示分页导航
        pagination.style.display = docInfo.prev || docInfo.next ? 'flex' : 'none';
    }
    
    highlightCode() {
        // 使用 Prism.js 高亮代码
        if (typeof Prism !== 'undefined') {
            Prism.highlightAll();
        }
    }
}

// 全局函数，供HTML调用
function loadDoc(docName) {
    if (window.docsSystem) {
        window.docsSystem.loadDoc(docName);
    }
}

// 添加到TORM命名空间
TORM.loadDoc = loadDoc;
window.loadDoc = loadDoc; // 保持向后兼容

// 工具函数（如果已存在则不重复定义）
if (typeof debounce === 'undefined') {
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
    TORM.debounce = debounce;
    window.debounce = debounce; // 保持向后兼容
}

if (typeof throttle === 'undefined') {
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
    TORM.throttle = throttle;
    window.throttle = throttle; // 保持向后兼容
}

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', function() {
    // 初始化文档系统
    window.docsSystem = new DocsSystem();
    
    // 添加搜索功能（简单版本）
    initDocSearch();
    
    // 添加键盘快捷键
    initKeyboardShortcuts();
});

// 搜索功能
function initDocSearch() {
    // 这里可以添加搜索功能
    // 现在先显示一个占位符
    const searchContainer = document.createElement('div');
    searchContainer.className = 'search-container';
    searchContainer.innerHTML = `
        <div class="search-box">
            <input type="text" placeholder="搜索文档..." class="search-input" disabled>
            <span class="search-hint">搜索功能开发中</span>
        </div>
    `;
    
    // 暂时不添加搜索框，等待后续开发
}

// 键盘快捷键
function initKeyboardShortcuts() {
    document.addEventListener('keydown', function(e) {
        // Ctrl/Cmd + K 打开搜索
        if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
            e.preventDefault();
            // 这里可以打开搜索对话框
            console.log('搜索功能待开发');
        }
        
        // 左右箭头键导航
        if (e.altKey) {
            if (e.key === 'ArrowLeft') {
                e.preventDefault();
                const prevBtn = document.getElementById('prev-page');
                if (prevBtn && prevBtn.style.display !== 'none') {
                    prevBtn.click();
                }
            } else if (e.key === 'ArrowRight') {
                e.preventDefault();
                const nextBtn = document.getElementById('next-page');
                if (nextBtn && nextBtn.style.display !== 'none') {
                    nextBtn.click();
                }
            }
        }
    });
}

// 添加错误处理样式
const errorStyles = `
    .error-content {
        text-align: center;
        padding: 4rem 2rem;
        color: var(--text-secondary);
    }
    
    .error-content h1 {
        color: var(--text-primary);
        margin-bottom: 1rem;
    }
    
    .error-message {
        font-family: var(--font-mono);
        background: var(--background-tertiary);
        padding: 1rem;
        border-radius: 0.5rem;
        margin: 1rem 0;
        color: #dc2626;
    }
    
    .retry-btn {
        background: var(--primary-green);
        color: var(--text-inverse);
        border: none;
        padding: 0.75rem 1.5rem;
        border-radius: 0.5rem;
        cursor: pointer;
        font-weight: 500;
        transition: all 0.3s ease;
    }
    
    .retry-btn:hover {
        background: var(--primary-green-dark);
        transform: translateY(-2px);
    }
    
    .no-toc {
        color: var(--text-tertiary);
        font-style: italic;
        text-align: center;
        padding: 1rem;
    }
    
    .search-container {
        padding: 1rem 1.5rem;
        border-bottom: 1px solid var(--border-light);
    }
    
    .search-box {
        position: relative;
    }
    
    .search-input {
        width: 100%;
        padding: 0.5rem 0.75rem;
        border: 1px solid var(--border);
        border-radius: 0.5rem;
        font-size: 0.875rem;
        background: var(--background);
        color: var(--text-secondary);
    }
    
    .search-hint {
        position: absolute;
        top: 100%;
        left: 0;
        font-size: 0.75rem;
        color: var(--text-tertiary);
        margin-top: 0.25rem;
    }
`;

// 动态添加样式
const docsStyleSheet = document.createElement('style');
docsStyleSheet.textContent = errorStyles;
document.head.appendChild(docsStyleSheet);

// 标记脚本已加载
if (window.TORM) {
    TORM.markScriptLoaded('docs.js');
} 