package main

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/xpzouying/xiaohongshu-mcp/browser"
	"github.com/xpzouying/xiaohongshu-mcp/cailianshe"
	"github.com/xpzouying/xiaohongshu-mcp/configs"
)

// CailiansheService 财联社服务
type CailiansheService struct {
	scheduler *cailianshe.NewsScheduler
}

// NewCailiansheService 创建财联社服务实例
func NewCailiansheService() *CailiansheService {
	return &CailiansheService{}
}

// FetchLatestNewsRequest 获取最新新闻请求
type FetchLatestNewsRequest struct {
	Limit int `json:"limit"`
}

// FetchLatestNewsResponse 获取最新新闻响应
type FetchLatestNewsResponse struct {
	News  []cailianshe.TelegraphNews `json:"news"`
	Count int                        `json:"count"`
}

// FetchLatestNews 获取最新新闻
func (s *CailiansheService) FetchLatestNews(ctx context.Context, limit int, fetchDetail bool) (*FetchLatestNewsResponse, error) {
	b := browser.NewBrowser(configs.IsHeadless(), browser.WithBinPath(configs.GetBinPath()))
	defer b.Close()

	page := b.NewPage()
	defer page.Close()

	action := cailianshe.NewNewsAction(page)

	// limit <= 0 表示获取所有新闻，不做限制
	// limit > 0 表示只获取指定数量

	newsList, err := action.FetchLatestNews(ctx, limit, fetchDetail)
	if err != nil {
		return nil, fmt.Errorf("获取新闻失败: %w", err)
	}

	return &FetchLatestNewsResponse{
		News:  newsList,
		Count: len(newsList),
	}, nil
}

// SearchNewsRequest 搜索新闻请求
type SearchNewsRequest struct {
	Keyword string `json:"keyword"`
	Limit   int    `json:"limit"`
}

// SearchNewsResponse 搜索新闻响应
type SearchNewsResponse struct {
	News    []cailianshe.TelegraphNews `json:"news"`
	Count   int                        `json:"count"`
	Keyword string                     `json:"keyword"`
}

// SearchNews 搜索新闻
func (s *CailiansheService) SearchNews(ctx context.Context, keyword string, limit int) (*SearchNewsResponse, error) {
	b := browser.NewBrowser(configs.IsHeadless(), browser.WithBinPath(configs.GetBinPath()))
	defer b.Close()

	page := b.NewPage()
	defer page.Close()

	action := cailianshe.NewNewsAction(page)

	// limit <= 0 表示获取所有搜索结果，不做限制

	newsList, err := action.SearchNews(ctx, keyword, limit)
	if err != nil {
		return nil, fmt.Errorf("搜索新闻失败: %w", err)
	}

	return &SearchNewsResponse{
		News:    newsList,
		Count:   len(newsList),
		Keyword: keyword,
	}, nil
}

// StartSchedulerRequest 启动定时任务请求
type StartSchedulerRequest struct {
	IntervalMinutes int `json:"interval_minutes"`
}

// StartSchedulerResponse 启动定时任务响应
type StartSchedulerResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Interval string `json:"interval"`
}

// StartScheduler 启动定时任务
func (s *CailiansheService) StartScheduler(ctx context.Context, intervalMinutes int) (*StartSchedulerResponse, error) {
	if s.scheduler != nil && s.scheduler.IsRunning() {
		return &StartSchedulerResponse{
			Success: true,
			Message: "定时任务已在运行中",
		}, nil
	}

	if intervalMinutes <= 0 {
		intervalMinutes = 5 // 默认5分钟
	}

	interval := time.Duration(intervalMinutes) * time.Minute

	b := browser.NewBrowser(configs.IsHeadless(), browser.WithBinPath(configs.GetBinPath()))
	page := b.NewPage()

	s.scheduler = cailianshe.NewNewsScheduler(page, interval)

	// 设置新新闻回调
	s.scheduler.SetNewNewsCallback(func(news []cailianshe.NewsWithAnalysis) {
		logrus.Infof("检测到 %d 条新新闻", len(news))
		for _, n := range news {
			logrus.Infof("新新闻: %s", n.News.Title)
			if n.Analysis != nil {
				logrus.Infof("  - 情感: %s, 行业: %v, 预测: %s",
					n.Analysis.Sentiment,
					n.Analysis.Industries,
					n.Analysis.Prediction)
			}
		}
	})

	if err := s.scheduler.Start(ctx); err != nil {
		return nil, fmt.Errorf("启动定时任务失败: %w", err)
	}

	return &StartSchedulerResponse{
		Success:  true,
		Message:  "定时任务启动成功",
		Interval: interval.String(),
	}, nil
}

// StopSchedulerResponse 停止定时任务响应
type StopSchedulerResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// StopScheduler 停止定时任务
func (s *CailiansheService) StopScheduler() (*StopSchedulerResponse, error) {
	if s.scheduler == nil || !s.scheduler.IsRunning() {
		return &StopSchedulerResponse{
			Success: true,
			Message: "定时任务未运行",
		}, nil
	}

	s.scheduler.Stop()

	return &StopSchedulerResponse{
		Success: true,
		Message: "定时任务已停止",
	}, nil
}

// GetSchedulerStatusResponse 获取定时任务状态响应
type GetSchedulerStatusResponse struct {
	Running       bool      `json:"running"`
	CachedCount   int       `json:"cached_count"`
	LastUpdate    time.Time `json:"last_update"`
	LastUpdateStr string    `json:"last_update_str"`
}

// GetSchedulerStatus 获取定时任务状态
func (s *CailiansheService) GetSchedulerStatus() (*GetSchedulerStatusResponse, error) {
	if s.scheduler == nil {
		return &GetSchedulerStatusResponse{
			Running: false,
		}, nil
	}

	count, lastUpdate := s.scheduler.GetCacheInfo()

	return &GetSchedulerStatusResponse{
		Running:       s.scheduler.IsRunning(),
		CachedCount:   count,
		LastUpdate:    lastUpdate,
		LastUpdateStr: lastUpdate.Format("2006-01-02 15:04:05"),
	}, nil
}

// GetCachedNewsResponse 获取缓存新闻响应
type GetCachedNewsResponse struct {
	News  []cailianshe.TelegraphNews `json:"news"`
	Count int                        `json:"count"`
}

// GetCachedNews 获取缓存的新闻（从全局缓存）
func (s *CailiansheService) GetCachedNews(limit int) (*GetCachedNewsResponse, error) {
	// limit <= 0 表示获取所有缓存新闻

	// 从全局缓存获取新闻
	news := cailianshe.GetGlobalCachedNews(limit)

	return &GetCachedNewsResponse{
		News:  news,
		Count: len(news),
	}, nil
}
