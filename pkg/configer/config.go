package configer

import (
	"os"

	"go.yaml.in/yaml/v3"
)

// Load 加载并合并多个 YAML 配置文件
//
// 参数:
//
//	defaultCfg: 默认配置（会被后续文件覆盖）
//	newRaw:     返回一个新的空 raw 结构体（必须是指针，且字段全为指针类型）
//	merge:      将 raw 中的非 nil 字段合并到 dst
//	paths:      配置文件路径，后面的覆盖前面的
func Load[T any, R any](
	defaultCfg *T,
	newRaw func() *R,
	merge func(dst *T, src *R),
	paths ...string,
) (*T, error) {
	cfg := *defaultCfg // 复制一份默认值

	for _, path := range paths {
		if path == "" {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue // 文件不存在时跳过
			}
			return nil, err
		}

		raw := newRaw() // 创建 raw 实例
		if err := yaml.Unmarshal(data, raw); err != nil {
			return nil, err
		}
		merge(&cfg, raw) // 合并到最终配置
	}
	return &cfg, nil
}
