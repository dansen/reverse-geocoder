package rgeocoder

// DataProcessor 处理原始GeoNames数据 -> rg_cities1000.csv（占位）
type DataProcessor struct {
	config *Config
}

func NewDataProcessor(cfg *Config) *DataProcessor { return &DataProcessor{config: cfg} }

func (p *DataProcessor) ProcessGeoNamesData() error {
	// TODO: 实现数据转换逻辑
	return nil
}
