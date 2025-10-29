package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
)

// 财联社MCP工具处理函数

// handleFetchLatestNews 处理获取最新新闻
func (s *AppServer) handleFetchLatestNews(ctx context.Context, args map[string]interface{}) *MCPToolResult {
	logrus.Info("MCP: 获取财联社最新新闻")
	logrus.Infof("MCP参数: %+v", args)

	limit := 0 // 默认0表示获取所有新闻
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
		logrus.Infof("设置limit参数: %d", limit)
	} else {
		logrus.Infof("未设置limit参数，获取所有新闻")
	}

	fetchDetail := false
	if fd, ok := args["fetch_detail"].(bool); ok {
		fetchDetail = fd
		logrus.Infof("设置fetch_detail参数: %v", fetchDetail)
	}

	logrus.Infof("调用FetchLatestNews: limit=%d, fetchDetail=%v", limit, fetchDetail)
	result, err := s.cailiansheService.FetchLatestNews(ctx, limit, fetchDetail)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "获取新闻失败: " + err.Error(),
			}},
			IsError: true,
		}
	}

	// 格式化输出
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: fmt.Sprintf("获取新闻成功，但序列化失败: %v", err),
			}},
			IsError: true,
		}
	}

	return &MCPToolResult{
		Content: []MCPContent{{
			Type: "text",
			Text: string(jsonData),
		}},
	}
}

// handleSearchCailiansheNews 处理搜索新闻
func (s *AppServer) handleSearchCailiansheNews(ctx context.Context, args map[string]interface{}) *MCPToolResult {
	logrus.Info("MCP: 搜索财联社新闻")

	keyword, ok := args["keyword"].(string)
	if !ok || keyword == "" {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "搜索失败: 缺少关键词参数",
			}},
			IsError: true,
		}
	}

	limit := 0 // 默认0表示返回所有搜索结果
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	result, err := s.cailiansheService.SearchNews(ctx, keyword, limit)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "搜索新闻失败: " + err.Error(),
			}},
			IsError: true,
		}
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: fmt.Sprintf("搜索成功，但序列化失败: %v", err),
			}},
			IsError: true,
		}
	}

	return &MCPToolResult{
		Content: []MCPContent{{
			Type: "text",
			Text: string(jsonData),
		}},
	}
}

// handleStartCailiansheScheduler 处理启动定时任务
func (s *AppServer) handleStartCailiansheScheduler(ctx context.Context, args map[string]interface{}) *MCPToolResult {
	logrus.Info("MCP: 启动财联社定时任务")

	intervalMinutes := 5
	if i, ok := args["interval_minutes"].(float64); ok {
		intervalMinutes = int(i)
	}

	result, err := s.cailiansheService.StartScheduler(ctx, intervalMinutes)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "启动定时任务失败: " + err.Error(),
			}},
			IsError: true,
		}
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: fmt.Sprintf("启动成功，但序列化失败: %v", err),
			}},
			IsError: true,
		}
	}

	return &MCPToolResult{
		Content: []MCPContent{{
			Type: "text",
			Text: string(jsonData),
		}},
	}
}

// handleStopCailiansheScheduler 处理停止定时任务
func (s *AppServer) handleStopCailiansheScheduler(ctx context.Context) *MCPToolResult {
	logrus.Info("MCP: 停止财联社定时任务")

	result, err := s.cailiansheService.StopScheduler()
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "停止定时任务失败: " + err.Error(),
			}},
			IsError: true,
		}
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: fmt.Sprintf("停止成功，但序列化失败: %v", err),
			}},
			IsError: true,
		}
	}

	return &MCPToolResult{
		Content: []MCPContent{{
			Type: "text",
			Text: string(jsonData),
		}},
	}
}

// handleGetCailiansheSchedulerStatus 处理获取定时任务状态
func (s *AppServer) handleGetCailiansheSchedulerStatus(ctx context.Context) *MCPToolResult {
	logrus.Info("MCP: 获取财联社定时任务状态")

	result, err := s.cailiansheService.GetSchedulerStatus()
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "获取状态失败: " + err.Error(),
			}},
			IsError: true,
		}
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: fmt.Sprintf("获取成功，但序列化失败: %v", err),
			}},
			IsError: true,
		}
	}

	return &MCPToolResult{
		Content: []MCPContent{{
			Type: "text",
			Text: string(jsonData),
		}},
	}
}

// handleGetCachedCailiansheNews 处理获取缓存新闻
func (s *AppServer) handleGetCachedCailiansheNews(ctx context.Context, args map[string]interface{}) *MCPToolResult {
	logrus.Info("MCP: 获取缓存的财联社新闻")

	limit := 0 // 默认0表示返回所有缓存新闻
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	result, err := s.cailiansheService.GetCachedNews(limit)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "获取缓存新闻失败: " + err.Error(),
			}},
			IsError: true,
		}
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: fmt.Sprintf("获取成功，但序列化失败: %v", err),
			}},
			IsError: true,
		}
	}

	return &MCPToolResult{
		Content: []MCPContent{{
			Type: "text",
			Text: string(jsonData),
		}},
	}
}
