package main

import "fmt"

const (
	MetaUrl            = "https://sdl-stickershop.line.naver.jp/stickershop/v1/product/%d/android/productInfo.meta"
	StickerUrl         = "https://stickershop.line-scdn.net/stickershop/v1/sticker/%d/android/sticker.png;compress=true"
	AnimatedStickerUrl = "https://sdl-stickershop.line.naver.jp/products/0/0/1/%d/android/animation/%d.png"
)

type (
	Response struct {
		PackageID    int  `json:"packageId"`
		OnSale       bool `json:"onSale"`
		ValidDays    int  `json:"validDays"`
		HasAnimation bool `json:"hasAnimation"`
		Title        struct {
			En   string `json:"en"`
			Es   string `json:"es"`
			In   string `json:"in"`
			Ja   string `json:"ja"`
			Ko   string `json:"ko"`
			Th   string `json:"th"`
			ZhCN string `json:"zh_CN"`
			ZhTW string `json:"zh_TW"`
		} `json:"title"`
		Author struct {
			En   string `json:"en"`
			Ja   string `json:"ja"`
			Ko   string `json:"ko"`
			ZhCN string `json:"zh_CN"`
			ZhTW string `json:"zh_TW"`
		} `json:"author"`
		Price []struct {
			Price    float64 `json:"price"`
			Symbol   string  `json:"symbol"`
			Currency string  `json:"currency"`
			Country  string  `json:"country"`
		} `json:"price"`
		Stickers []Sticker `json:"stickers"`
	}

	Sticker struct {
		ID     int `json:"id"`
		Height int `json:"height"`
		Width  int `json:"width"`
	}
)

func (r *Response) LocalizedTitle() string {
	if r.Title.En != "" {
		return r.Title.En
	}
	return r.Title.Ja
}

func (r *Response) LocalizedAuthor() string {
	if r.Author.En != "" {
		return r.Author.En
	}
	return r.Author.Ja
}

func (s *Sticker) AnimatedDownloadUrl(packageID int) string {
	return fmt.Sprintf(AnimatedStickerUrl, packageID, s.ID)
}

func (s *Sticker) DownloadUrl() string {
	return fmt.Sprintf(StickerUrl, s.ID)
}

func (s *Sticker) FileName() string {
	// Animated stickers are APNGs
	return fmt.Sprintf("%d.png", s.ID)
}
