// Package catalogue orchestrates YouTube channel metadata ingestion via yt-dlp.
package catalogue

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/coldcanuk/catfu/internal/backends"
	"github.com/coldcanuk/catfu/internal/store"
)

// Service implements backends.Cataloguer.
type Service struct {
	Store  *store.Store
	YTDLP  *YTDLP
	Logger *slog.Logger
}

// CatalogueChannel resolves a channel and streams video metadata into the store.
func (s *Service) CatalogueChannel(ctx context.Context, channelURL string, opts backends.CatalogueOpts) error {
	_, _, err := s.CatalogueChannelWithProgress(ctx, channelURL, opts)
	return err
}

// CatalogueChannelWithProgress catalogues a channel and reports progress.
func (s *Service) CatalogueChannelWithProgress(ctx context.Context, channelURL string, opts backends.CatalogueOpts) (channelID string, count int, err error) {
	if s.Store == nil || s.YTDLP == nil {
		return "", 0, fmt.Errorf("catalogue service not configured")
	}
	url := NormalizeChannelURL(channelURL)
	log := s.Logger
	if log == nil {
		log = slog.Default()
	}

	meta, err := s.YTDLP.ResolveChannel(ctx, url)
	if err != nil {
		return "", 0, err
	}
	ch := store.Channel{
		ID:           meta.ID,
		Title:        meta.Title,
		Description:  meta.Description,
		CustomURL:    meta.CustomURL,
		URL:          meta.URL,
		ThumbnailURL: meta.ThumbnailURL,
	}
	if ch.URL == "" {
		ch.URL = url
	}
	if err := s.Store.UpsertChannel(ctx, ch); err != nil {
		return meta.ID, 0, fmt.Errorf("upsert channel: %w", err)
	}

	dump := DumpOpts{
		FullMetadata: opts.FullMetadata,
		Limit:        opts.Limit,
		SleepRequest: opts.SleepRequest,
		SleepMin:     opts.SleepMin,
		SleepMax:     opts.SleepMax,
		DateAfter:    opts.DateAfter,
	}
	count = 0
	_, err = s.YTDLP.StreamEntries(ctx, url, dump, func(entry RawEntry) error {
		vid, title, desc, upload, thumb, page, chID, _, duration, views, likes := MapEntry(entry, meta.ID)
		if vid == "" || title == "" {
			return nil
		}
		if chID == "" {
			chID = meta.ID
		}
		v := store.Video{
			ID:           vid,
			ChannelID:    chID,
			Title:        title,
			Description:  desc,
			UploadDate:   upload,
			Duration:     duration,
			ViewCount:    views,
			LikeCount:    likes,
			ThumbnailURL: thumb,
			WebpageURL:   page,
		}
		if err := s.Store.UpsertVideo(ctx, v); err != nil {
			return err
		}
		count++
		if opts.Progress != nil {
			opts.Progress(count, vid, title)
		}
		return nil
	})
	_ = s.Store.TouchChannelCatalogued(ctx, meta.ID, time.Now())
	if err != nil {
		return meta.ID, count, err
	}
	log.Info("catalogue complete", "channel_id", meta.ID, "title", meta.Title, "videos", count)
	return meta.ID, count, nil
}

// Ensure Service implements Cataloguer.
var _ backends.Cataloguer = (*Service)(nil)
