# TORM 官方网站

这是 TORM (Go语言ORM库) 的官方网站源代码。

## 🌟 特性

- **绿色主题**: 以绿色为主题色的现代化设计
- **毛玻璃效果**: 使用 backdrop-filter 实现的毛玻璃效果
- **响应式设计**: 完美适配桌面端、平板和移动设备
- **文档系统**: 集成 wiki 文档的展示系统
- **交互式演示**: 代码示例的标签切换展示
- **现代化UI**: 参考 GORM 官网的设计理念

## 📁 文件结构

```
web/
├── index.html          # 主页
├── docs.html           # 文档页面
├── styles/
│   ├── main.css        # 主要样式文件
│   └── docs.css        # 文档页面样式
├── scripts/
│   ├── main.js         # 主页JavaScript
│   └── docs.js         # 文档页面JavaScript
├── assets/
│   └── logo.svg        # LOGO文件
└── README.md           # 本文件
```

## 🚀 使用方法

### 本地运行

1. 启动一个简单的HTTP服务器：

```bash
# 使用Python
python -m http.server 8000

# 使用Node.js
npx serve .

# 使用Go
go run -m httptest.NewServer
```

2. 在浏览器中访问 `http://localhost:8000`

### 部署到生产环境

可以将整个 `web` 目录部署到任何静态网站托管服务：

- **Netlify**: 直接拖拽文件夹上传
- **Vercel**: 连接 Git 仓库自动部署
- **GitHub Pages**: 推送到 gh-pages 分支
- **传统服务器**: 上传到 Web 服务器目录

## 🎨 设计特色

### 颜色系统

- **主绿色**: `#22c55e` - 用于主要按钮、链接等
- **深绿色**: `#16a34a` - 用于悬停状态
- **浅绿色**: `#4ade80` - 用于装饰元素
- **绿色背景**: `#dcfce7` - 用于高亮区域

### 毛玻璃效果

使用 CSS `backdrop-filter` 属性实现：

```css
.glass {
    background: rgba(255, 255, 255, 0.25);
    backdrop-filter: blur(16px);
    -webkit-backdrop-filter: blur(16px);
    border: 1px solid rgba(255, 255, 255, 0.18);
}
```

### 响应式断点

- **桌面端**: > 1024px
- **平板端**: 768px - 1024px  
- **移动端**: < 768px

## 🛠️ 技术栈

- **HTML5**: 语义化标签
- **CSS3**: 现代CSS特性（Grid, Flexbox, backdrop-filter）
- **Vanilla JavaScript**: 无框架依赖
- **第三方库**:
  - Prism.js: 代码高亮
  - Marked.js: Markdown解析
  - Inter字体: 现代化字体

## 📚 文档系统

文档系统会自动从 `../wiki/` 目录加载 Markdown 文件：

- 支持实时 Markdown 解析
- 自动生成页面目录
- 代码语法高亮
- 页面间导航
- 移动端适配

### 添加新文档

1. 在 `wiki/` 目录添加 `.md` 文件
2. 在 `scripts/docs.js` 的 `docs` 对象中添加配置：

```javascript
'new-doc': {
    title: '新文档标题',
    file: '../wiki/New-Doc.md',
    prev: 'previous-doc',
    next: 'next-doc'
}
```

## 🔧 自定义配置

### 修改主题色

在 `styles/main.css` 中修改 CSS 变量：

```css
:root {
    --primary-green: #your-color;
    --primary-green-dark: #your-dark-color;
    --primary-green-light: #your-light-color;
}
```

### 添加新功能

- **搜索功能**: 在 `docs.js` 中实现 `initDocSearch` 函数
- **多语言**: 扩展语言选择器功能
- **评论系统**: 集成第三方评论系统

## 🎯 性能优化

- **图片优化**: 使用 SVG 格式的 LOGO
- **字体优化**: 使用 Google Fonts 的 font-display: swap
- **代码分割**: JavaScript 按页面分离
- **缓存策略**: 适合设置较长的缓存时间

## 📱 移动端优化

- **触摸友好**: 按钮尺寸适合手指点击
- **手势支持**: 滑动关闭侧边栏
- **性能优化**: 节流滚动事件处理
- **可访问性**: 支持键盘导航

## 🔍 SEO 优化

- **语义化HTML**: 正确使用标题层级
- **Meta标签**: 适当的描述和关键词
- **Open Graph**: 社交媒体分享优化
- **结构化数据**: 可添加 JSON-LD 数据

## 🤝 贡献指南

1. Fork 本仓库
2. 创建特性分支
3. 提交更改
4. 发起 Pull Request

## 📄 许可证

本项目采用与 TORM 项目相同的许可证。

---

**💚 感谢使用 TORM！** 