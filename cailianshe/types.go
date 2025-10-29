package cailianshe

import "time"

// TelegraphNews 财联社电报新闻
type TelegraphNews struct {
	ID          string    `json:"id"`           // 新闻ID
	Title       string    `json:"title"`        // 新闻标题
	Content     string    `json:"content"`      // 新闻内容
	Brief       string    `json:"brief"`        // 新闻摘要
	PublishTime time.Time `json:"publish_time"` // 发布时间
	Source      string    `json:"source"`       // 来源
	Tags        []string  `json:"tags"`         // 标签
	URL         string    `json:"url"`          // 新闻链接
}

// NewsListResponse 新闻列表响应
type NewsListResponse struct {
	News  []TelegraphNews `json:"news"`
	Count int             `json:"count"`
	Total int             `json:"total"`
}

// NewsAnalysis 新闻分析结果
type NewsAnalysis struct {
	NewsID     string    `json:"news_id"`     // 新闻ID
	Summary    string    `json:"summary"`     // 内容摘要
	Sentiment  string    `json:"sentiment"`   // 情感倾向（positive/negative/neutral）
	Keywords   []string  `json:"keywords"`    // 关键词
	Industries []string  `json:"industries"`  // 相关行业
	Stocks     []string  `json:"stocks"`      // 相关股票代码
	Impact     string    `json:"impact"`      // 影响分析
	Prediction string    `json:"prediction"`  // 预测建议
	Confidence float64   `json:"confidence"`  // 置信度 (0-1)
	AnalyzedAt time.Time `json:"analyzed_at"` // 分析时间
}

// NewsWithAnalysis 带分析的新闻
type NewsWithAnalysis struct {
	News     TelegraphNews `json:"news"`
	Analysis *NewsAnalysis `json:"analysis,omitempty"`
}

// FetchNewsRequest 获取新闻请求
type FetchNewsRequest struct {
	Limit     int    `json:"limit"`      // 获取数量限制
	Keyword   string `json:"keyword"`    // 关键词筛选（可选）
	StartTime string `json:"start_time"` // 开始时间（可选）
	EndTime   string `json:"end_time"`   // 结束时间（可选）
}

// AnalyzeNewsRequest 分析新闻请求
type AnalyzeNewsRequest struct {
	NewsID  string `json:"news_id"` // 新闻ID
	Content string `json:"content"` // 新闻内容
}
