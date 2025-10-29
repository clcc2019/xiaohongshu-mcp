package cailianshe

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/sirupsen/logrus"
)

// 全局新闻缓存
var (
	globalNewsCache      = make(map[string]TelegraphNews) // ID -> 新闻
	globalNewsCacheMutex sync.RWMutex
	globalNewsCacheTime  time.Time
)

// NewsAction 财联社新闻操作
type NewsAction struct {
	page         *rod.Page
	contentCache map[string]string // URL -> 完整内容的缓存
}

// NewNewsAction 创建新闻操作实例
func NewNewsAction(page *rod.Page) *NewsAction {
	pp := page.Timeout(60 * time.Second)
	return &NewsAction{
		page:         pp,
		contentCache: make(map[string]string),
	}
}

// FetchLatestNews 获取最新的电报新闻
func (n *NewsAction) FetchLatestNews(ctx context.Context, limit int, fetchDetail bool) ([]TelegraphNews, error) {
	page := n.page.Context(ctx)

	logrus.Info("正在访问财联社电报页面获取最新新闻...")

	// 访问财联社电报页面
	if err := page.Navigate("https://www.cls.cn/telegraph"); err != nil {
		return nil, fmt.Errorf("访问财联社电报页面失败: %w", err)
	}

	// 等待页面加载完成
	page.MustWaitLoad()
	time.Sleep(5 * time.Second) // 增加等待时间，确保动态内容加载

	logrus.Info("正在分析页面结构...")

	// 先检查页面结构
	debugInfo := page.MustEval(`() => {
		const info = {
			title: document.title,
			hasInitialState: !!window.__INITIAL_STATE__,
			bodyClasses: document.body.className || '',
			allClasses: []
		};
		
		// 收集所有class名称
		const elements = document.querySelectorAll('[class]');
		const classSet = new Set();
		elements.forEach(el => {
			const className = el.className;
			if (className && typeof className === 'string') {
				className.split(' ').forEach(cls => {
					if (cls.trim()) classSet.add(cls.trim());
				});
			} else if (className && className.baseVal) {
				// SVG元素的className是对象
				const classes = className.baseVal.split(' ');
				classes.forEach(cls => {
					if (cls.trim()) classSet.add(cls.trim());
				});
			}
		});
		info.allClasses = Array.from(classSet).slice(0, 50); // 只取前50个
		
		return JSON.stringify(info);
	}`).String()

	logrus.Infof("页面调试信息: %s", debugInfo)

	logrus.Info("正在提取新闻数据...")

	// 执行JavaScript提取新闻数据
	result := page.MustEval(`() => {
		const newsList = [];
		
		// 策略1: 从 __NEXT_DATA__ 获取（财联社使用Next.js）
		try {
			const nextDataScript = document.getElementById('__NEXT_DATA__');
			if (nextDataScript) {
				const nextData = JSON.parse(nextDataScript.textContent);
				const telegraphList = nextData?.props?.initialState?.telegraph?.telegraphList;
				
				if (telegraphList && Array.isArray(telegraphList) && telegraphList.length > 0) {
					console.log('从 __NEXT_DATA__ 找到数据，数量:', telegraphList.length);
					
					// 转换为我们需要的格式
					telegraphList.forEach((item, index) => {
						const news = {
							id: item.id ? String(item.id) : 'cls_' + Date.now() + '_' + index,
							title: item.title || '',
							content: item.content || item.brief || '',
							brief: item.brief || (item.content ? item.content.substring(0, 200) : ''),
							publish_time: item.ctime ? new Date(item.ctime * 1000).toISOString() : new Date().toISOString(),
							source: '财联社',
							tags: item.subjects ? item.subjects.map(s => s.subject_name) : [],
							url: item.shareurl || 'https://www.cls.cn/detail/' + item.id
						};
						newsList.push(news);
					});
					
					return JSON.stringify(newsList);
				}
			}
		} catch (e) {
			console.log('从 __NEXT_DATA__ 获取失败:', e);
		}
		
		// 策略2: 尝试从全局变量获取
		if (window.__INITIAL_STATE__) {
			try {
				const state = window.__INITIAL_STATE__;
				console.log('找到 __INITIAL_STATE__');
				
				// 尝试多种可能的数据路径
				const possiblePaths = [
					state.telegraph?.telegraphList,
					state.telegraph?.list,
					state.telegraph?.data?.list,
					state.data?.telegraph?.list,
					state.list,
					state.data?.list
				];
				
				for (const data of possiblePaths) {
					if (data && Array.isArray(data) && data.length > 0) {
						console.log('从全局变量找到数据，数量:', data.length);
						return JSON.stringify(data);
					}
				}
			} catch (e) {
				console.log('从全局变量获取失败:', e);
			}
		}
		
		// 策略3: 尝试多种DOM选择器
		const selectors = [
			'li[class*="telegraph"]',
			'div[class*="telegraph"]',
			'li[class*="item"]',
			'div[class*="item"]',
			'li[class*="list"]',
			'div[class*="list"]',
			'.list-item',
			'.item',
			'article',
			'[class*="news"]'
		];
		
		let newsItems = [];
		for (const selector of selectors) {
			newsItems = document.querySelectorAll(selector);
			if (newsItems.length > 0) {
				console.log('使用选择器找到元素:', selector, '数量:', newsItems.length);
				break;
			}
		}
		
		if (newsItems.length === 0) {
			console.log('未找到新闻元素');
			return JSON.stringify([]);
		}
		
		// 策略4: 从DOM元素中提取
		newsItems.forEach((item, index) => {
			try {
				// 尝试多种方式提取标题
				let title = '';
				const titleSelectors = [
					'.title',
					'[class*="title"]',
					'h1', 'h2', 'h3', 'h4',
					'a',
					'span[class*="title"]',
					'div[class*="title"]'
				];
				
				for (const sel of titleSelectors) {
					const el = item.querySelector(sel);
					if (el && el.textContent.trim()) {
						title = el.textContent.trim();
						break;
					}
				}
				
				// 如果还是没有标题，尝试直接从item获取
				if (!title) {
					const allText = item.textContent.trim();
					if (allText) {
						title = allText.split('\n')[0].trim();
					}
				}
				
				if (!title || title.length < 5) {
					return; // 跳过无效项
				}
				
				// 提取内容
				let content = '';
				const contentSelectors = [
					'.content',
					'[class*="content"]',
					'.brief',
					'[class*="brief"]',
					'p',
					'.desc',
					'[class*="desc"]'
				];
				
				for (const sel of contentSelectors) {
					const el = item.querySelector(sel);
					if (el && el.textContent.trim()) {
						content = el.textContent.trim();
						break;
					}
				}
				
				if (!content) {
					content = item.textContent.trim();
				}
				
				// 提取时间
				let time = '';
				const timeSelectors = [
					'.time',
					'[class*="time"]',
					'time',
					'[class*="date"]',
					'.date'
				];
				
				for (const sel of timeSelectors) {
					const el = item.querySelector(sel);
					if (el && el.textContent.trim()) {
						time = el.textContent.trim();
						break;
					}
				}
				
				// 提取链接
				let url = '';
				const linkEl = item.querySelector('a');
				if (linkEl && linkEl.href) {
					url = linkEl.href;
				}
				
				const news = {
					id: 'cls_' + Date.now() + '_' + index,
					title: title,
					content: content,
					brief: content.substring(0, 200),
					publish_time: time || new Date().toISOString(),
					source: '财联社',
					tags: [],
					url: url || window.location.href
				};
				
				newsList.push(news);
			} catch (e) {
				console.log('提取新闻项失败:', e);
			}
		});
		
		console.log('最终提取到新闻数量:', newsList.length);
		return JSON.stringify(newsList);
	}`).String()

	if result == "" || result == "[]" {
		logrus.Warn("未能从页面提取到新闻数据")

		// 保存页面HTML用于调试
		html := page.MustHTML()
		logrus.Debugf("页面HTML长度: %d", len(html))

		return nil, fmt.Errorf("页面未返回新闻数据，请检查页面结构")
	}

	var newsList []TelegraphNews
	if err := json.Unmarshal([]byte(result), &newsList); err != nil {
		return nil, fmt.Errorf("解析新闻数据失败: %w", err)
	}

	logrus.Infof("从页面提取到 %d 条新闻", len(newsList))

	// 通过ID去重
	newsList = deduplicateNews(newsList)
	logrus.Infof("去重后剩余 %d 条新闻", len(newsList))

	// 处理时间格式
	for i := range newsList {
		if newsList[i].PublishTime.IsZero() {
			newsList[i].PublishTime = time.Now()
		}
		if newsList[i].Source == "" {
			newsList[i].Source = "财联社"
		}
	}

	// 更新全局缓存（合并新旧新闻）
	n.updateGlobalCache(newsList)

	// 从缓存中获取最新的新闻（包含刚刚添加的和之前的）
	finalNewsList := n.getLatestNewsFromCache(limit)

	logrus.Infof("返回 %d 条新闻（从缓存中获取最新的）", len(finalNewsList))

	// 根据fetchDetail参数决定是否获取详细内容
	if !fetchDetail {
		logrus.Info("跳过获取详细内容（快速模式）")
		return finalNewsList, nil
	}

	// 使用finalNewsList替换newsList
	newsList = finalNewsList

	// 获取每条新闻的完整内容（串行处理，避免并发问题）
	logrus.Info("开始获取新闻详细内容...")
	successCount := 0
	cachedCount := 0
	for i := range newsList {
		if newsList[i].URL != "" {
			// 检查缓存
			if cachedContent, exists := n.contentCache[newsList[i].URL]; exists {
				logrus.Debugf("第 %d/%d 条新闻从缓存获取: %s", i+1, len(newsList), newsList[i].Title)
				newsList[i].Content = cachedContent
				cachedCount++
				successCount++
				continue
			}

			logrus.Debugf("正在获取第 %d/%d 条新闻详情: %s", i+1, len(newsList), newsList[i].Title)
			fullContent, err := n.FetchNewsDetail(ctx, newsList[i].URL)
			if err != nil {
				logrus.Warnf("获取新闻详情失败 [%d/%d] %s: %v", i+1, len(newsList), newsList[i].URL, err)
				// 继续处理其他新闻，不中断
				continue
			}
			// 更新新闻内容为完整内容
			if fullContent != "" {
				newsList[i].Content = fullContent
				// 保存到缓存
				n.contentCache[newsList[i].URL] = fullContent
				successCount++
				logrus.Debugf("成功获取第 %d/%d 条新闻详情，内容长度: %d", i+1, len(newsList), len(fullContent))
			}
		}
	}

	logrus.Infof("成功获取 %d/%d 条新闻的详细内容（缓存命中: %d 条）", successCount, len(newsList), cachedCount)
	return newsList, nil
}

// FetchNewsDetail 获取新闻详细内容
func (n *NewsAction) FetchNewsDetail(ctx context.Context, url string) (string, error) {
	// 随机延迟1-2秒，模拟人类行为（减少等待时间）
	randomDelay := 1 + rand.IntN(2) // 1到2秒之间的随机数
	logrus.Debugf("随机延迟 %d 秒后访问详情页", randomDelay)
	time.Sleep(time.Duration(randomDelay) * time.Second)

	// 访问详情页
	if err := n.page.Navigate(url); err != nil {
		return "", fmt.Errorf("访问详情页失败: %w", err)
	}

	// 等待页面加载
	if err := n.page.WaitLoad(); err != nil {
		return "", fmt.Errorf("等待页面加载失败: %w", err)
	}

	// 等待内容加载（减少等待时间）
	time.Sleep(2 * time.Second)

	// 提取详细内容
	content := n.page.MustEval(`() => {
		// 策略1: 查找 telegraph-content 类（电报新闻）
		let contentEl = document.querySelector('.telegraph-content');
		if (contentEl && contentEl.textContent.trim()) {
			console.log('使用策略1: .telegraph-content');
			return contentEl.textContent.trim();
		}
		
		// 策略2: 查找 content-box 下的 content
		contentEl = document.querySelector('.content-box .content');
		if (contentEl && contentEl.textContent.trim()) {
			console.log('使用策略2: .content-box .content');
			return contentEl.textContent.trim();
		}
		
		// 策略3: 直接查找 .content
		contentEl = document.querySelector('section.content-box .content');
		if (contentEl && contentEl.textContent.trim()) {
			console.log('使用策略3: section.content-box .content');
			return contentEl.textContent.trim();
		}
		
		// 策略4: 查找任何包含 telegraph 的内容
		contentEl = document.querySelector('[class*="telegraph-content"]');
		if (contentEl && contentEl.textContent.trim()) {
			console.log('使用策略4: [class*="telegraph-content"]');
			return contentEl.textContent.trim();
		}
		
		// 策略5: 遍历多个可能的选择器
		const contentSelectors = [
			'div.telegraph-content',
			'div.content',
			'[class*="article-content"]',
			'[class*="detail-content"]',
			'article',
			'main',
			'[class*="content"]'
		];
		
		for (const selector of contentSelectors) {
			const elements = document.querySelectorAll(selector);
			for (const el of elements) {
				const text = el.textContent.trim();
				// 内容长度要合理（至少50个字符）
				if (text.length > 50 && text.length < 50000) {
					console.log('使用策略5:', selector, '长度:', text.length);
					return text;
				}
			}
		}
		
		console.log('所有策略都失败');
		return '';
	}`).String()

	if content == "" {
		return "", fmt.Errorf("未能提取到详细内容")
	}

	// 清理内容（去除多余空白）
	content = strings.TrimSpace(content)

	return content, nil
}

// SearchNews 搜索新闻
func (n *NewsAction) SearchNews(ctx context.Context, keyword string, limit int) ([]TelegraphNews, error) {
	// 先获取所有新闻（搜索不需要详细内容，使用快速模式）
	allNews, err := n.FetchLatestNews(ctx, 0, false)
	if err != nil {
		return nil, err
	}

	// 过滤包含关键词的新闻
	var filteredNews []TelegraphNews
	keyword = strings.ToLower(keyword)

	for _, news := range allNews {
		titleLower := strings.ToLower(news.Title)
		contentLower := strings.ToLower(news.Content)

		if strings.Contains(titleLower, keyword) || strings.Contains(contentLower, keyword) {
			filteredNews = append(filteredNews, news)

			if limit > 0 && len(filteredNews) >= limit {
				break
			}
		}
	}

	logrus.Infof("搜索关键词 '%s' 找到 %d 条相关新闻", keyword, len(filteredNews))
	return filteredNews, nil
}

// GetNewsByID 根据ID获取新闻详情
func (n *NewsAction) GetNewsByID(ctx context.Context, newsID string) (*TelegraphNews, error) {
	page := n.page.Context(ctx)

	url := fmt.Sprintf("https://www.cls.cn/telegraph/%s", newsID)
	if err := page.Navigate(url); err != nil {
		return nil, fmt.Errorf("访问新闻详情页失败: %w", err)
	}

	page.MustWaitLoad()
	time.Sleep(2 * time.Second)

	// 提取新闻详情
	result := page.MustEval(`() => {
		const titleEl = document.querySelector('.detail-title, h1, [class*="title"]');
		const contentEl = document.querySelector('.detail-content, .content, [class*="content"]');
		const timeEl = document.querySelector('.detail-time, time, [class*="time"]');
		
		return JSON.stringify({
			title: titleEl ? titleEl.textContent.trim() : '',
			content: contentEl ? contentEl.textContent.trim() : '',
			publish_time: timeEl ? timeEl.textContent.trim() : new Date().toISOString()
		});
	}`).String()

	var newsDetail struct {
		Title       string `json:"title"`
		Content     string `json:"content"`
		PublishTime string `json:"publish_time"`
	}

	if err := json.Unmarshal([]byte(result), &newsDetail); err != nil {
		return nil, fmt.Errorf("解析新闻详情失败: %w", err)
	}

	publishTime, _ := time.Parse("2006-01-02 15:04:05", newsDetail.PublishTime)
	if publishTime.IsZero() {
		publishTime = time.Now()
	}

	news := &TelegraphNews{
		ID:          newsID,
		Title:       newsDetail.Title,
		Content:     newsDetail.Content,
		Brief:       newsDetail.Content[:min(200, len(newsDetail.Content))],
		PublishTime: publishTime,
		Source:      "财联社",
		URL:         url,
	}

	return news, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// deduplicateNews 通过ID去重新闻列表
func deduplicateNews(newsList []TelegraphNews) []TelegraphNews {
	if len(newsList) == 0 {
		return newsList
	}

	// 使用map记录已见过的ID
	seen := make(map[string]bool)
	result := make([]TelegraphNews, 0, len(newsList))

	for _, news := range newsList {
		// 如果ID为空，跳过（保留，因为无法判断是否重复）
		if news.ID == "" {
			result = append(result, news)
			continue
		}

		// 如果ID未见过，添加到结果
		if !seen[news.ID] {
			seen[news.ID] = true
			result = append(result, news)
		} else {
			logrus.Debugf("发现重复新闻，ID: %s, 标题: %s", news.ID, news.Title)
		}
	}

	return result
}

// getLatestNewsFromCache 从缓存获取最新的新闻
func (n *NewsAction) getLatestNewsFromCache(limit int) []TelegraphNews {
	return GetGlobalCachedNews(limit)
}

// GetGlobalCachedNews 从全局缓存获取新闻（公共函数）
func GetGlobalCachedNews(limit int) []TelegraphNews {
	globalNewsCacheMutex.RLock()
	defer globalNewsCacheMutex.RUnlock()

	// 将缓存转换为列表
	newsList := make([]TelegraphNews, 0, len(globalNewsCache))
	for _, news := range globalNewsCache {
		newsList = append(newsList, news)
	}

	// 按发布时间排序（最新的在前）
	for i := 0; i < len(newsList)-1; i++ {
		for j := i + 1; j < len(newsList); j++ {
			if newsList[i].PublishTime.Before(newsList[j].PublishTime) {
				newsList[i], newsList[j] = newsList[j], newsList[i]
			}
		}
	}

	// 应用limit限制
	if limit > 0 && len(newsList) > limit {
		newsList = newsList[:limit]
	}

	return newsList
}

// updateGlobalCache 更新全局新闻缓存（合并新旧新闻）
func (n *NewsAction) updateGlobalCache(newsList []TelegraphNews) {
	globalNewsCacheMutex.Lock()
	defer globalNewsCacheMutex.Unlock()

	oldCount := len(globalNewsCache)
	newCount := 0

	// 合并新新闻到缓存（保留旧新闻，添加新新闻）
	for _, news := range newsList {
		if news.ID != "" {
			if _, exists := globalNewsCache[news.ID]; !exists {
				newCount++
			}
			globalNewsCache[news.ID] = news
		}
	}

	// 清理过期新闻（保留最近100条）
	if len(globalNewsCache) > 100 {
		// 转换为列表并按时间排序
		allNews := make([]TelegraphNews, 0, len(globalNewsCache))
		for _, news := range globalNewsCache {
			allNews = append(allNews, news)
		}

		// 简单排序（最新的在前）
		for i := 0; i < len(allNews)-1; i++ {
			for j := i + 1; j < len(allNews); j++ {
				if allNews[i].PublishTime.Before(allNews[j].PublishTime) {
					allNews[i], allNews[j] = allNews[j], allNews[i]
				}
			}
		}

		// 只保留最新的100条
		globalNewsCache = make(map[string]TelegraphNews)
		for i := 0; i < 100 && i < len(allNews); i++ {
			globalNewsCache[allNews[i].ID] = allNews[i]
		}
	}

	// 更新缓存时间
	globalNewsCacheTime = time.Now()

	logrus.Infof("已更新全局新闻缓存，总计 %d 条（新增 %d 条，原有 %d 条）", len(globalNewsCache), newCount, oldCount)
}

// getNewsFromCache 从缓存获取新闻
func (n *NewsAction) getNewsFromCache(limit int, fetchDetail bool) ([]TelegraphNews, error) {
	globalNewsCacheMutex.RLock()
	defer globalNewsCacheMutex.RUnlock()

	// 将缓存转换为列表
	newsList := make([]TelegraphNews, 0, len(globalNewsCache))
	for _, news := range globalNewsCache {
		newsList = append(newsList, news)
	}

	// 按发布时间排序（最新的在前）
	for i := 0; i < len(newsList)-1; i++ {
		for j := i + 1; j < len(newsList); j++ {
			if newsList[i].PublishTime.Before(newsList[j].PublishTime) {
				newsList[i], newsList[j] = newsList[j], newsList[i]
			}
		}
	}

	// 应用limit限制
	if limit > 0 && len(newsList) > limit {
		newsList = newsList[:limit]
	}

	logrus.Infof("从缓存返回 %d 条新闻", len(newsList))

	// 如果需要详细内容，获取详情
	if fetchDetail {
		logrus.Info("开始获取新闻详细内容...")
		successCount := 0
		cachedCount := 0

		// 创建临时context用于获取详情
		ctx := context.Background()

		for i := range newsList {
			if newsList[i].URL != "" {
				// 检查内容缓存
				if cachedContent, exists := n.contentCache[newsList[i].URL]; exists {
					logrus.Debugf("第 %d/%d 条新闻从内容缓存获取: %s", i+1, len(newsList), newsList[i].Title)
					newsList[i].Content = cachedContent
					cachedCount++
					successCount++
					continue
				}

				logrus.Debugf("正在获取第 %d/%d 条新闻详情: %s", i+1, len(newsList), newsList[i].Title)
				fullContent, err := n.FetchNewsDetail(ctx, newsList[i].URL)
				if err != nil {
					logrus.Warnf("获取新闻详情失败 [%d/%d] %s: %v", i+1, len(newsList), newsList[i].URL, err)
					continue
				}

				if fullContent != "" {
					newsList[i].Content = fullContent
					n.contentCache[newsList[i].URL] = fullContent
					successCount++
					logrus.Debugf("成功获取第 %d/%d 条新闻详情，内容长度: %d", i+1, len(newsList), len(fullContent))
				}
			}
		}

		logrus.Infof("成功获取 %d/%d 条新闻的详细内容（内容缓存命中: %d 条）", successCount, len(newsList), cachedCount)
	}

	return newsList, nil
}
