package rgeocoder

import (
	"fmt"
)

// Downloader 下载GeoNames数据（占位）
type Downloader struct {
	config *Config
}

func NewDownloader(cfg *Config) *Downloader { return &Downloader{config: cfg} }

func (d *Downloader) DownloadRequired() error {
	// TODO: 实现下载逻辑
	if d.config.Verbose {
		fmt.Println("DownloadRequired placeholder")
	}
	return nil
}
