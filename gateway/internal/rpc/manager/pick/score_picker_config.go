package pick

import "time"

// ScorePickerConfig 是 ScorePicker 的评分配置
type ScorePickerConfig struct {
	ActiveWeight      float64       // 活跃请求数权重，值越大，活跃请求越多时惩罚越重
	LatencyWeight     float64       // 平均耗时权重，值越大，平均耗时越高时惩罚越重
	FailureRateWeight float64       // 失败率权重，值越大，失败率越高时惩罚越重
	LatencyBase       time.Duration // 平均耗时归一化基准，平均耗时 / LatencyBase 用于转成评分项
	WarmupPenalty     float64       // warm-up 惩罚系数，处于 warm-up 的节点会额外增加该惩罚
	MinSampleCount    uint64        // 最低样本数，小于该样本数时降低失败率惩罚影响，避免冷启动误判
}

// setDefault 设置默认值
func (c *ScorePickerConfig) SetDefault() {
	if c.ActiveWeight <= 0 {
		c.ActiveWeight = 10
	}
	if c.LatencyWeight <= 0 {
		c.LatencyWeight = 20
	}
	if c.FailureRateWeight <= 0 {
		c.FailureRateWeight = 80
	}
	if c.WarmupPenalty <= 0 {
		c.WarmupPenalty = 15
	}
	if c.LatencyBase <= 0 {
		c.LatencyBase = 300 * time.Millisecond
	}
	if c.MinSampleCount == 0 {
		c.MinSampleCount = 20
	}
}
