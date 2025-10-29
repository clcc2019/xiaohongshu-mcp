package cailianshe

import (
	"context"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/sirupsen/logrus"
)

// NewsScheduler 新闻定时任务调度器
type NewsScheduler struct {
	interval   time.Duration
	newsAction *NewsAction
	analyzer   *NewsAnalyzer
	cache      *NewsCache
	running    bool
	stopChan   chan struct{}
	mu         sync.RWMutex
	onNewNews  func([]NewsWithAnalysis) // 新新闻回调
}

// NewsCache 新闻缓存
type NewsCache struct {
	news       []NewsWithAnalysis
	lastUpdate time.Time
	mu         sync.RWMutex
}

// NewNewsCache 创建新闻缓存
func NewNewsCache() *NewsCache {
	return &NewsCache{
		news: make([]NewsWithAnalysis, 0),
	}
}

// Set 设置缓存
func (c *NewsCache) Set(news []NewsWithAnalysis) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.news = news
	c.lastUpdate = time.Now()
}

// Get 获取缓存
func (c *NewsCache) Get() ([]NewsWithAnalysis, time.Time) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.news, c.lastUpdate
}

// GetLatest 获取最新N条
func (c *NewsCache) GetLatest(n int) []NewsWithAnalysis {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if n <= 0 || n > len(c.news) {
		return c.news
	}
	return c.news[:n]
}

// NewNewsScheduler 创建新闻调度器
func NewNewsScheduler(page *rod.Page, interval time.Duration) *NewsScheduler {
	return &NewsScheduler{
		interval:   interval,
		newsAction: NewNewsAction(page),
		analyzer:   NewNewsAnalyzer(),
		cache:      NewNewsCache(),
		stopChan:   make(chan struct{}),
	}
}

// Start 启动定时任务
func (s *NewsScheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = true
	s.mu.Unlock()

	logrus.Infof("启动财联社新闻定时任务，间隔: %v", s.interval)

	// 立即执行一次（使用传入的ctx）
	if err := s.fetchAndAnalyze(ctx); err != nil {
		logrus.Errorf("初始获取新闻失败: %v", err)
	}

	// 启动定时任务（使用独立的context，不受启动请求影响）
	go s.run(context.Background())

	return nil
}

// Stop 停止定时任务
func (s *NewsScheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	logrus.Info("停止财联社新闻定时任务")
	s.running = false
	close(s.stopChan)
}

// IsRunning 检查是否运行中
func (s *NewsScheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// run 运行定时任务
func (s *NewsScheduler) run(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 每次执行时创建新的context，设置超时
			fetchCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
			if err := s.fetchAndAnalyze(fetchCtx); err != nil {
				logrus.Errorf("定时获取新闻失败: %v", err)
			}
			cancel()
		case <-s.stopChan:
			logrus.Info("定时任务已停止")
			return
		case <-ctx.Done():
			logrus.Info("上下文已取消，定时任务退出")
			return
		}
	}
}

// fetchAndAnalyze 获取新闻（不再分析）
func (s *NewsScheduler) fetchAndAnalyze(ctx context.Context) error {
	logrus.Info("开始获取最新财联社新闻...")

	// 定时任务使用快速模式，只获取摘要，避免超时
	newsList, err := s.newsAction.FetchLatestNews(ctx, 20, false)
	if err != nil {
		return err
	}

	if len(newsList) == 0 {
		logrus.Warn("未获取到新闻")
		return nil
	}

	logrus.Infof("获取到 %d 条新闻", len(newsList))

	// 转换为NewsWithAnalysis格式（不包含分析数据）
	newsWithAnalysis := make([]NewsWithAnalysis, len(newsList))
	for i := range newsList {
		newsWithAnalysis[i] = NewsWithAnalysis{
			News: newsList[i],
		}
	}

	// 检查是否有新新闻
	oldNews, _ := s.cache.Get()
	newNewsItems := s.findNewNews(newsWithAnalysis, oldNews)

	// 更新缓存
	s.cache.Set(newsWithAnalysis)

	logrus.Infof("新闻获取完成，共 %d 条，其中新增 %d 条", len(newsWithAnalysis), len(newNewsItems))

	// 触发回调
	if len(newNewsItems) > 0 && s.onNewNews != nil {
		s.onNewNews(newNewsItems)
	}

	return nil
}

// findNewNews 找出新增的新闻
func (s *NewsScheduler) findNewNews(current, old []NewsWithAnalysis) []NewsWithAnalysis {
	if len(old) == 0 {
		return current
	}

	oldIDs := make(map[string]bool)
	for _, news := range old {
		oldIDs[news.News.ID] = true
	}

	var newNews []NewsWithAnalysis
	for _, news := range current {
		if !oldIDs[news.News.ID] {
			newNews = append(newNews, news)
		}
	}

	return newNews
}

// GetCachedNews 获取缓存的新闻
func (s *NewsScheduler) GetCachedNews(limit int) []NewsWithAnalysis {
	return s.cache.GetLatest(limit)
}

// GetCacheInfo 获取缓存信息
func (s *NewsScheduler) GetCacheInfo() (count int, lastUpdate time.Time) {
	news, lastUpdate := s.cache.Get()
	return len(news), lastUpdate
}

// SetNewNewsCallback 设置新新闻回调
func (s *NewsScheduler) SetNewNewsCallback(callback func([]NewsWithAnalysis)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onNewNews = callback
}

// ForceUpdate 强制更新
func (s *NewsScheduler) ForceUpdate(ctx context.Context) error {
	logrus.Info("强制更新新闻...")
	return s.fetchAndAnalyze(ctx)
}
