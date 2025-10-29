# 财联社MCP快速开始指南

## 5分钟快速上手

### 第一步：启动服务 (1分钟)

```bash
# 进入项目目录
cd /home/narwal/workdir/devops/xiaohongshu-mcp

# 启动服务
go run .
```

看到以下输出表示启动成功：
```
INFO MCP Server initialized with official SDK (Cailianshe + Xiaohongshu)
INFO 启动 HTTP 服务器: :18060
INFO Registered 11 MCP tools
INFO Registered 8 Cailianshe MCP tools
```

### 第二步：配置MCP客户端 (2分钟)

#### 使用Cursor

1. 在项目根目录创建 `.cursor/mcp.json`：

```json
{
  "mcpServers": {
    "cailianshe-mcp": {
      "url": "http://localhost:18060/mcp",
      "description": "财联社新闻获取与分析"
    }
  }
}
```

2. 重启Cursor

#### 使用Claude Code CLI

```bash
claude mcp add --transport http cailianshe-mcp http://localhost:18060/mcp
```

### 第三步：开始使用 (2分钟)

在MCP客户端中尝试以下对话：

#### 示例1：获取最新新闻
```
你: 获取最新的10条财联社新闻

AI: [调用 fetch_cailianshe_news]
获取到10条最新新闻...
```

#### 示例2：分析新闻
```
你: 分析这条新闻：央行宣布降息

AI: [调用 analyze_cailianshe_news]
分析结果：
- 情感: 利好
- 行业: 金融、地产
- 预测: 建议关注金融板块...
```

#### 示例3：启动定时监控
```
你: 启动定时任务，每5分钟获取一次新闻

AI: [调用 start_cailianshe_scheduler]
定时任务已启动，间隔5分钟
```

## 常用命令速查

### 获取新闻
```
获取最新的财联社新闻
搜索关于"人工智能"的新闻
获取并分析最新5条新闻
```

### 分析新闻
```
分析这条新闻：[标题和内容]
分析新能源行业的最新动态
```

### 定时任务
```
启动定时任务，每5分钟更新一次
查看定时任务状态
停止定时任务
获取缓存的新闻
```

## 验证安装

运行测试脚本：
```bash
./test_cailianshe.sh
```

看到以下输出表示安装成功：
```
✓ MCP服务正在运行
✓ MCP初始化成功
✓ 财联社工具已注册
  总共注册了 19 个工具
```

## 下一步

- 📖 阅读 [完整文档](README_CAILIANSHE.md)
- 💡 查看 [使用示例](USAGE_EXAMPLES.md)
- 📊 了解 [项目架构](PROJECT_SUMMARY.md)

## 常见问题

**Q: 服务启动失败？**  
A: 检查端口18060是否被占用，可以使用 `-port` 参数指定其他端口

**Q: 获取新闻失败？**  
A: 检查网络连接，确保能访问 www.cls.cn

**Q: MCP客户端连接不上？**  
A: 确认服务正在运行，检查配置文件中的URL是否正确

## 需要帮助？

- 查看 [常见问题](README_CAILIANSHE.md#常见问题)
- 提交 [GitHub Issue](https://github.com/xpzouying/xiaohongshu-mcp/issues)

---

**提示**: 首次运行会自动下载Chrome浏览器（约150MB），请耐心等待。

