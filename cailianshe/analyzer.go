package cailianshe

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// NewsAnalyzer 新闻分析器
type NewsAnalyzer struct {
	// 可以集成外部AI服务，如OpenAI、Claude等
	// 这里先实现基础的关键词分析
}

// NewNewsAnalyzer 创建新闻分析器
func NewNewsAnalyzer() *NewsAnalyzer {
	return &NewsAnalyzer{}
}

// AnalyzeNews 分析新闻内容
func (a *NewsAnalyzer) AnalyzeNews(ctx context.Context, news *TelegraphNews) (*NewsAnalysis, error) {
	logrus.Infof("开始分析新闻: %s", news.Title)

	analysis := &NewsAnalysis{
		NewsID:     news.ID,
		AnalyzedAt: time.Now(),
	}

	// 1. 提取关键词
	analysis.Keywords = a.extractKeywords(news.Title + " " + news.Content)

	// 2. 识别相关行业
	analysis.Industries = a.identifyIndustries(news.Title + " " + news.Content)

	// 3. 识别相关股票
	analysis.Stocks = a.identifyStocks(news.Title + " " + news.Content)

	// 4. 情感分析
	analysis.Sentiment = a.analyzeSentiment(news.Title + " " + news.Content)

	// 5. 生成摘要
	analysis.Summary = a.generateSummary(news)

	// 6. 影响分析
	analysis.Impact = a.analyzeImpact(news, analysis)

	// 7. 预测建议
	analysis.Prediction = a.generatePrediction(news, analysis)

	// 8. 计算置信度
	analysis.Confidence = a.calculateConfidence(analysis)

	logrus.Infof("新闻分析完成: %s, 情感: %s, 置信度: %.2f", news.Title, analysis.Sentiment, analysis.Confidence)

	return analysis, nil
}

// extractKeywords 提取关键词
func (a *NewsAnalyzer) extractKeywords(text string) []string {
	keywords := make(map[string]bool)

	// 金融相关关键词列表
	financialKeywords := []string{
		"央行", "货币政策", "利率", "降息", "加息", "存款准备金",
		"GDP", "CPI", "PPI", "PMI", "通胀", "通缩",
		"股市", "A股", "港股", "美股", "指数", "涨停", "跌停",
		"IPO", "重组", "并购", "增发", "配股",
		"业绩", "财报", "营收", "利润", "亏损",
		"监管", "政策", "改革", "开放",
		"科技", "芯片", "新能源", "人工智能", "5G",
		"房地产", "地产", "楼市",
		"消费", "零售", "电商",
		"金融", "银行", "保险", "证券",
		"制造", "工业", "基建",
	}

	textLower := strings.ToLower(text)
	for _, keyword := range financialKeywords {
		if strings.Contains(textLower, strings.ToLower(keyword)) {
			keywords[keyword] = true
		}
	}

	// 转换为切片
	result := make([]string, 0, len(keywords))
	for k := range keywords {
		result = append(result, k)
	}

	return result
}

// identifyIndustries 识别相关行业
func (a *NewsAnalyzer) identifyIndustries(text string) []string {
	industries := make(map[string]bool)

	industryKeywords := map[string][]string{
		"科技":  {"科技", "互联网", "软件", "芯片", "半导体", "人工智能", "AI", "5G", "云计算"},
		"金融":  {"银行", "保险", "证券", "基金", "信托", "金融"},
		"地产":  {"房地产", "地产", "楼市", "住宅", "商业地产"},
		"消费":  {"消费", "零售", "电商", "商超", "百货"},
		"医药":  {"医药", "生物", "制药", "医疗", "器械"},
		"能源":  {"能源", "石油", "天然气", "煤炭", "电力"},
		"新能源": {"新能源", "光伏", "风电", "锂电", "电池", "充电桩"},
		"汽车":  {"汽车", "车企", "新能源车", "电动车"},
		"制造":  {"制造", "工业", "机械", "装备"},
		"基建":  {"基建", "建筑", "工程", "铁路", "公路"},
		"农业":  {"农业", "种植", "养殖", "农产品"},
		"传媒":  {"传媒", "影视", "游戏", "广告"},
	}

	textLower := strings.ToLower(text)
	for industry, keywords := range industryKeywords {
		for _, keyword := range keywords {
			if strings.Contains(textLower, strings.ToLower(keyword)) {
				industries[industry] = true
				break
			}
		}
	}

	result := make([]string, 0, len(industries))
	for k := range industries {
		result = append(result, k)
	}

	return result
}

// identifyStocks 识别相关股票代码
func (a *NewsAnalyzer) identifyStocks(text string) []string {
	stocks := make(map[string]bool)

	// 匹配股票代码格式：6位数字（A股）或5位数字（港股）
	stockPattern := regexp.MustCompile(`\b([036]\d{5}|[0-9]{5})\b`)
	matches := stockPattern.FindAllString(text, -1)

	for _, match := range matches {
		stocks[match] = true
	}

	// 匹配带前缀的股票代码
	prefixPattern := regexp.MustCompile(`(SH|SZ|HK)?[036]\d{5}`)
	prefixMatches := prefixPattern.FindAllString(text, -1)

	for _, match := range prefixMatches {
		stocks[match] = true
	}

	result := make([]string, 0, len(stocks))
	for k := range stocks {
		result = append(result, k)
	}

	return result
}

// analyzeSentiment 情感分析
func (a *NewsAnalyzer) analyzeSentiment(text string) string {
	positiveWords := []string{
		"上涨", "增长", "提升", "改善", "利好", "突破", "创新高", "超预期",
		"强劲", "积极", "乐观", "看好", "机会", "受益", "推动", "支持",
	}

	negativeWords := []string{
		"下跌", "下降", "下滑", "恶化", "利空", "跌破", "创新低", "不及预期",
		"疲软", "悲观", "担忧", "风险", "压力", "冲击", "拖累", "制约",
	}

	textLower := strings.ToLower(text)
	positiveCount := 0
	negativeCount := 0

	for _, word := range positiveWords {
		positiveCount += strings.Count(textLower, strings.ToLower(word))
	}

	for _, word := range negativeWords {
		negativeCount += strings.Count(textLower, strings.ToLower(word))
	}

	if positiveCount > negativeCount {
		return "positive"
	} else if negativeCount > positiveCount {
		return "negative"
	}
	return "neutral"
}

// generateSummary 生成摘要
func (a *NewsAnalyzer) generateSummary(news *TelegraphNews) string {
	// 简单实现：取标题和内容前200字
	summary := news.Title
	if len(news.Content) > 0 {
		contentPreview := news.Content
		if len(contentPreview) > 200 {
			contentPreview = contentPreview[:200] + "..."
		}
		summary = fmt.Sprintf("%s - %s", news.Title, contentPreview)
	}
	return summary
}

// analyzeImpact 分析影响
func (a *NewsAnalyzer) analyzeImpact(news *TelegraphNews, analysis *NewsAnalysis) string {
	var impacts []string

	// 根据情感和行业生成影响分析
	sentimentDesc := map[string]string{
		"positive": "利好",
		"negative": "利空",
		"neutral":  "中性",
	}

	if len(analysis.Industries) > 0 {
		impacts = append(impacts, fmt.Sprintf("该新闻对%s行业呈%s影响",
			strings.Join(analysis.Industries, "、"),
			sentimentDesc[analysis.Sentiment]))
	}

	if len(analysis.Stocks) > 0 {
		impacts = append(impacts, fmt.Sprintf("可能影响股票: %s",
			strings.Join(analysis.Stocks, "、")))
	}

	// 根据关键词添加具体影响
	text := news.Title + " " + news.Content
	if strings.Contains(text, "政策") || strings.Contains(text, "监管") {
		impacts = append(impacts, "政策面影响较大，需关注后续政策落地情况")
	}
	if strings.Contains(text, "业绩") || strings.Contains(text, "财报") {
		impacts = append(impacts, "业绩相关消息，可能影响相关公司估值")
	}
	if strings.Contains(text, "央行") || strings.Contains(text, "货币政策") {
		impacts = append(impacts, "宏观政策影响，可能波及整体市场")
	}

	if len(impacts) == 0 {
		return "影响程度有限，建议持续关注"
	}

	return strings.Join(impacts, "；")
}

// generatePrediction 生成预测建议
func (a *NewsAnalyzer) generatePrediction(news *TelegraphNews, analysis *NewsAnalysis) string {
	var predictions []string

	switch analysis.Sentiment {
	case "positive":
		if len(analysis.Industries) > 0 {
			predictions = append(predictions, fmt.Sprintf("建议关注%s板块的投资机会",
				strings.Join(analysis.Industries, "、")))
		}
		predictions = append(predictions, "短期内相关标的可能有上涨动能")

	case "negative":
		predictions = append(predictions, "建议谨慎对待相关板块，注意风险控制")
		if len(analysis.Industries) > 0 {
			predictions = append(predictions, fmt.Sprintf("%s板块可能面临压力",
				strings.Join(analysis.Industries, "、")))
		}

	case "neutral":
		predictions = append(predictions, "建议保持观望，等待更多信息")
		predictions = append(predictions, "可适当关注后续发展")
	}

	// 根据置信度调整建议
	if analysis.Confidence < 0.5 {
		predictions = append(predictions, "注意：分析置信度较低，建议结合更多信息综合判断")
	}

	return strings.Join(predictions, "；")
}

// calculateConfidence 计算置信度
func (a *NewsAnalyzer) calculateConfidence(analysis *NewsAnalysis) float64 {
	confidence := 0.3 // 基础置信度

	// 有关键词加分
	if len(analysis.Keywords) > 0 {
		confidence += 0.2
	}

	// 有行业识别加分
	if len(analysis.Industries) > 0 {
		confidence += 0.2
	}

	// 有股票代码加分
	if len(analysis.Stocks) > 0 {
		confidence += 0.15
	}

	// 情感明确加分
	if analysis.Sentiment != "neutral" {
		confidence += 0.15
	}

	// 确保在0-1之间
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// BatchAnalyzeNews 批量分析新闻
func (a *NewsAnalyzer) BatchAnalyzeNews(ctx context.Context, newsList []TelegraphNews) ([]NewsWithAnalysis, error) {
	results := make([]NewsWithAnalysis, 0, len(newsList))

	for i := range newsList {
		analysis, err := a.AnalyzeNews(ctx, &newsList[i])
		if err != nil {
			logrus.Warnf("分析新闻失败: %s, 错误: %v", newsList[i].Title, err)
			// 即使分析失败也返回新闻，只是没有分析结果
			results = append(results, NewsWithAnalysis{
				News:     newsList[i],
				Analysis: nil,
			})
			continue
		}

		results = append(results, NewsWithAnalysis{
			News:     newsList[i],
			Analysis: analysis,
		})
	}

	return results, nil
}

// ExportAnalysisToJSON 导出分析结果为JSON
func (a *NewsAnalyzer) ExportAnalysisToJSON(analysis *NewsAnalysis) (string, error) {
	data, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		return "", fmt.Errorf("导出JSON失败: %w", err)
	}
	return string(data), nil
}
