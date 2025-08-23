# TORM 官方网站

这是 TORM (Go语言ORM库) 的官方网站源代码。

## 🌟 特性

- **绿色主题**: 以绿色为主题色的现代化设计
- **毛玻璃效果**: 使用 backdrop-filter 实现的毛玻璃效果
- **响应式设计**: 完美适配桌面端、平板和移动设备
- **文档系统**: 集成 wiki 文档的展示系统
- **交互式演示**: 代码示例的标签切换展示
- **现代化UI**: 参考现代ORM框架官网的设计理念
- **TORM 统一标签**: 展示最新的 `torm` 标签语法
- **大小写不敏感**: 支持各种大小写组合的标签语法

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

## 💡 TORM 标签特性

### 统一标签语法

TORM v1.1.6 引入了统一的 `torm` 标签，简化了模型定义：

```go
type User struct {
    model.BaseModel
    ID    uint   `torm:"primary_key,auto_increment,comment:用户ID"`
    Name  string `torm:"type:varchar,size:50,comment:用户名"`
    Email string `torm:"type:varchar,size:100,unique,comment:邮箱"`
    Price float64 `torm:"type:decimal,precision:10,scale:2,default:0.00"`
}
```

### 大小写不敏感

所有标签都支持大小写不敏感，以下写法完全等效：

```go
// 全小写（推荐）
`torm:"type:varchar,size:50,unique,comment:用户名"`

// 全大写
`torm:"TYPE:VARCHAR,SIZE:50,UNIQUE,COMMENT:用户名"`

// 首字母大写
`torm:"Type:VarChar,Size:50,Unique,Comment:用户名"`

// 混合大小写
`torm:"TYPE:varchar,SIZE:50,unique,COMMENT:用户名"`
```

### AutoMigrate 功能

支持自动迁移功能，根据模型结构体自动创建数据库表：

```go
func NewUser() *User {
    user := &User{}
    user.BaseModel = *model.NewBaseModelWithAutoDetect(user)
    user.SetTable("users")
    user.AutoMigrate() // 自动创建表结构
    return user
}
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

### TORM 特性展示

网站完整展示了 TORM v1.1.6 的核心特性：

- **统一标签语法**: 展示 `torm` 标签的使用方法
- **大小写不敏感**: 演示各种大小写组合
- **AutoMigrate**: 自动数据库迁移功能
- **查询构建器**: 强大的 SQL 查询构建器
- **模型系统**: Active Record 模式的模型操作
- **事务处理**: 完整的事务支持
- **多数据库**: MySQL、PostgreSQL、SQLite、MongoDB 支持

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

## 📈 更新历史

### v1.1.6 (最新)
- ✅ 引入统一的 `torm` 标签语法
- ✅ 支持大小写不敏感的标签解析
- ✅ 新增 AutoMigrate 自动迁移功能
- ✅ 增强的 WHERE 查询方法（NULL、BETWEEN、EXISTS 等）
- ✅ 高级排序功能（OrderRand、OrderField、FieldRaw）
- ✅ 完全移除 GORM 依赖
- ✅ 新增 `NewBaseModelWithAutoDetect` 方法
- ✅ 精确的类型长度和精度控制

---

**💚 感谢使用 TORM！** 