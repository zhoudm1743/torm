# 贡献指南

感谢您对TORM项目的关注！我们欢迎各种形式的贡献，包括但不限于代码、文档、问题反馈和功能建议。

## 🤝 如何贡献

### 报告问题

1. 在 [GitHub Issues](https://github.com/zhoudm1743/torm/issues) 搜索是否已有相同问题
2. 如果没有，创建新的 Issue
3. 请提供：
   - 详细的问题描述
   - 复现步骤
   - 期望行为
   - 实际行为
   - 环境信息（Go版本、数据库版本等）

### 功能建议

1. 在 Issues 中创建功能请求
2. 描述新功能的用途和价值
3. 提供具体的使用场景
4. 考虑向后兼容性

### 代码贡献

1. Fork 项目
2. 创建功能分支：`git checkout -b feature/your-feature`
3. 编写代码和测试
4. 提交变更：`git commit -am 'Add some feature'`
5. 推送到分支：`git push origin feature/your-feature`
6. 创建 Pull Request

## 📝 开发规范

### 代码风格

- 遵循 Go 官方代码规范
- 使用 `gofmt` 格式化代码
- 使用 `golint` 检查代码
- 保持代码简洁和可读性

### 提交规范

使用以下格式的提交信息：

```
type(scope): subject

body

footer
```

类型：
- `feat`: 新功能
- `fix`: 修复bug
- `docs`: 文档更新
- `style`: 代码格式（不影响代码运行）
- `refactor`: 重构
- `test`: 测试相关
- `chore`: 构建过程或辅助工具的变动

示例：
```
feat(db): add PostgreSQL support

- Implement PostgreSQL connection driver
- Add PostgreSQL-specific query builder
- Update configuration to support PostgreSQL options

Closes #123
```

### 测试要求

- 所有新功能必须包含单元测试
- 测试覆盖率应保持在 90% 以上
- 运行 `go test ./...` 确保所有测试通过

## 🛠️ 开发环境

### 环境要求

- Go 1.19+
- Git
- 支持的数据库（MySQL、PostgreSQL、SQLite、MongoDB）

### 本地开发

```bash
# 克隆项目
git clone https://github.com/zhoudm1743/torm.git
cd torm

# 安装依赖
go mod tidy

# 运行测试
go test ./...

# 运行示例
go run examples/basic_usage.go
```

### 数据库设置

```bash
# MySQL
docker run -d --name mysql -p 3306:3306 -e MYSQL_ROOT_PASSWORD=password mysql:8.0

# PostgreSQL
docker run -d --name postgres -p 5432:5432 -e POSTGRES_PASSWORD=password postgres:15

# MongoDB
docker run -d --name mongodb -p 27017:27017 mongo:7.0
```

## 📋 Pull Request 流程

### 提交前检查

- [ ] 代码已格式化（`gofmt`）
- [ ] 通过 lint 检查（`golint`）
- [ ] 所有测试通过（`go test ./...`）
- [ ] 添加了必要的测试
- [ ] 更新了相关文档
- [ ] 提交信息符合规范

### PR 模板

```markdown
## 变更类型
- [ ] Bug 修复
- [ ] 新功能
- [ ] 重构
- [ ] 文档更新
- [ ] 其他

## 变更描述
简要描述本次变更的内容

## 相关 Issue
Closes #(issue number)

## 测试
- [ ] 已添加单元测试
- [ ] 已添加集成测试
- [ ] 手动测试通过

## 检查清单
- [ ] 代码已格式化
- [ ] 通过 lint 检查
- [ ] 所有测试通过
- [ ] 文档已更新
```

## 🎯 贡献方向

我们特别欢迎以下方面的贡献：

### 高优先级
- 性能优化
- 更多数据库驱动支持
- 查询构建器功能增强
- 文档和示例完善

### 中优先级
- 连接池优化
- 缓存系统增强
- 日志系统改进
- 错误处理优化

### 低优先级
- 代码重构
- 测试覆盖率提升
- 开发工具改进

## 👥 社区

### 沟通渠道

- **GitHub Issues**: 问题报告和功能请求
- **GitHub Discussions**: 技术讨论和问答
- **Email**: zhoudm1743@163.com

### 行为准则

我们致力于营造一个开放、友好的社区环境：

1. **尊重他人**: 尊重不同的观点和经验水平
2. **建设性反馈**: 提供有用的、建设性的反馈
3. **包容性**: 欢迎来自不同背景的贡献者
4. **专业性**: 保持专业和友善的沟通方式

## 🏆 贡献者认可

我们感谢所有贡献者的努力：

- 代码贡献者将被列入 Contributors 列表
- 重要贡献者将被邀请成为项目维护者
- 优秀贡献将在发布说明中特别提及

## 📚 资源

### 相关文档
- [快速开始](Quick-Start)
- [API 参考](API-Reference)
- [示例代码](Examples)

### 学习资源
- [Go 官方文档](https://golang.org/doc/)
- [数据库设计最佳实践](https://www.google.com/search?q=database+design+best+practices)
- [ORM 设计模式](https://www.google.com/search?q=orm+design+patterns)

## ❓ 常见问题

### Q: 我是Go新手，可以贡献吗？
A: 当然可以！我们欢迎各个水平的贡献者。您可以从文档、示例或简单的bug修复开始。

### Q: 如何选择要解决的Issue？
A: 查看标有 `good first issue` 或 `help wanted` 的Issue，这些通常适合新贡献者。

### Q: 我的PR被拒绝了怎么办？
A: 不要气馁！查看反馈意见，进行相应修改后重新提交。维护者会帮助您改进代码。

### Q: 可以添加新的数据库支持吗？
A: 可以！我们欢迎更多数据库驱动的贡献。请先创建Issue讨论实现方案。

---

**🙏 感谢您考虑为TORM项目做出贡献！每一个贡献都让项目变得更好。** 