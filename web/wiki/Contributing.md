# 贡献指南

感谢你对 TORM 项目的关注！我们欢迎各种形式的贡献。

## 🤝 如何贡献

### 报告问题

1. 在 [GitHub Issues](https://github.com/zhoudm1743/torm/issues) 搜索是否已有相同问题
2. 如果没有，创建新的 Issue
3. 请提供：
   - 详细的问题描述
   - 复现步骤
   - 期望行为 vs 实际行为
   - 环境信息（Go 版本、数据库版本等）

### 功能建议

1. 在 Issues 中创建功能请求
2. 描述新功能的用途和价值
3. 提供具体的使用场景

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
```

类型：
- `feat`: 新功能
- `fix`: 修复 bug
- `docs`: 文档更新
- `style`: 代码格式
- `refactor`: 重构
- `test`: 测试相关
- `chore`: 构建/工具变动

示例：
```
feat(query): add WhereExists method

- Implement WhereExists for subquery support
- Add WhereNotExists method
- Update documentation

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
- 支持的数据库（MySQL、PostgreSQL、SQLite）

### 本地开发

```bash
# 克隆项目
git clone https://github.com/zhoudm1743/torm.git
cd torm

# 安装依赖
go mod tidy

# 运行测试
go test ./...
```

### 数据库设置

```bash
# MySQL
docker run -d --name mysql -p 3306:3306 -e MYSQL_ROOT_PASSWORD=password mysql:8.0

# PostgreSQL
docker run -d --name postgres -p 5432:5432 -e POSTGRES_PASSWORD=password postgres:15
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

## 变更描述
简要描述本次变更的内容

## 相关 Issue
Closes #(issue number)

## 测试
- [ ] 已添加单元测试
- [ ] 手动测试通过
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

### 低优先级
- 代码重构
- 测试覆盖率提升

## 👥 社区

### 沟通渠道

- **GitHub Issues**: 问题报告和功能请求
- **GitHub Discussions**: 技术讨论和问答
- **Email**: zhoudm1743@163.com

### 行为准则

1. **尊重他人**: 尊重不同的观点和经验水平
2. **建设性反馈**: 提供有用的、建设性的反馈
3. **包容性**: 欢迎来自不同背景的贡献者
4. **专业性**: 保持专业和友善的沟通方式

## 🏆 贡献者认可

- 代码贡献者将被列入 Contributors 列表
- 重要贡献者将被邀请成为项目维护者
- 优秀贡献将在发布说明中特别提及

## ❓ 常见问题

### Q: 我是 Go 新手，可以贡献吗？
A: 当然可以！你可以从文档、示例或简单的 bug 修复开始。

### Q: 如何选择要解决的 Issue？
A: 查看标有 `good first issue` 或 `help wanted` 的 Issue。

### Q: 我的 PR 被拒绝了怎么办？
A: 不要气馁！查看反馈意见，进行相应修改后重新提交。

---

**🙏 感谢你考虑为 TORM 项目做出贡献！**
