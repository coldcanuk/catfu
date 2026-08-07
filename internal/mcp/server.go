// Package mcp implements the catfu Model Context Protocol server (stdio).
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/coldcanuk/catfu/internal/backends"
	"github.com/coldcanuk/catfu/internal/backends/brave"
	"github.com/coldcanuk/catfu/internal/catalogue"
	"github.com/coldcanuk/catfu/internal/config"
	localsearch "github.com/coldcanuk/catfu/internal/search"
	"github.com/coldcanuk/catfu/internal/store"
	"github.com/coldcanuk/catfu/pkg/version"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Options configures the MCP server.
type Options struct {
	DBPath      string
	YTDLP       string
	BraveAPIKey string
	SleepReq    float64
	SleepMin    float64
	SleepMax    float64
	Logger      *slog.Logger
}

// Run starts the stdio MCP server and blocks until the client disconnects.
func Run(ctx context.Context, opts Options) error {
	log := opts.Logger
	if log == nil {
		log = slog.Default()
	}
	if err := config.EnsureDBDir(opts.DBPath); err != nil {
		return err
	}
	st, err := store.Open(ctx, opts.DBPath)
	if err != nil {
		return err
	}
	defer st.Close()

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "catfu",
		Version: version.Version,
	}, nil)

	type catalogueArgs struct {
		ChannelURL string `json:"channel_url" jsonschema:"YouTube channel URL, @handle, or UC id"`
		Full       bool   `json:"full,omitempty" jsonschema:"fetch full metadata (slower)"`
		Limit      int    `json:"limit,omitempty" jsonschema:"max videos (0=all)"`
	}
	type catalogueOut struct {
		ChannelID string `json:"channel_id"`
		Videos    int    `json:"videos"`
		Status    string `json:"status"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "catalogue_channel",
		Description: "Catalogue YouTube channel video metadata into the local SQLite store (no download)",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args catalogueArgs) (*mcp.CallToolResult, catalogueOut, error) {
		yt := &catalogue.YTDLP{Bin: opts.YTDLP, Logger: log}
		svc := &catalogue.Service{Store: st, YTDLP: yt, Logger: log}
		id, n, err := svc.CatalogueChannelWithProgress(ctx, args.ChannelURL, backends.CatalogueOpts{
			FullMetadata: args.Full,
			Limit:        args.Limit,
			SleepRequest: opts.SleepReq,
			SleepMin:     opts.SleepMin,
			SleepMax:     opts.SleepMax,
		})
		if err != nil {
			return nil, catalogueOut{}, err
		}
		out := catalogueOut{ChannelID: id, Videos: n, Status: "ok"}
		b, _ := json.Marshal(out)
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}, out, nil
	})

	type searchArgs struct {
		Query     string `json:"query" jsonschema:"full-text query (optional if filtering by channel/date)"`
		ChannelID string `json:"channel_id,omitempty" jsonschema:"optional channel id or @handle"`
		After     string `json:"after,omitempty" jsonschema:"YYYY-MM-DD"`
		Before    string `json:"before,omitempty" jsonschema:"YYYY-MM-DD"`
		Limit     int    `json:"limit,omitempty"`
		Offset    int    `json:"offset,omitempty"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_catalogue",
		Description: "Search the local YouTube metadata catalogue with FTS and date filters",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args searchArgs) (*mcp.CallToolResult, any, error) {
		sq := backends.SearchQuery{Query: args.Query, Limit: args.Limit, Offset: args.Offset}
		if args.ChannelID != "" {
			if ch, _ := st.GetChannel(ctx, args.ChannelID); ch != nil {
				sq.ChannelID = ch.ID
			} else {
				sq.ChannelID = args.ChannelID
			}
		}
		if t, err := localsearch.ParseDate(args.After); err == nil {
			sq.After = t
		}
		if t, err := localsearch.ParseDate(args.Before); err == nil {
			sq.Before = t
		}
		cs := &localsearch.CatalogueSearcher{Store: st}
		results, err := cs.Search(ctx, sq)
		if err != nil {
			return nil, nil, err
		}
		b, _ := json.Marshal(results)
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}, results, nil
	})

	type webArgs struct {
		Query      string `json:"query" jsonschema:"search query"`
		Kind       string `json:"kind,omitempty" jsonschema:"web, news, or video (default web)"`
		Limit      int    `json:"limit,omitempty"`
		Offset     int    `json:"offset,omitempty"`
		Country    string `json:"country,omitempty" jsonschema:"ISO country code e.g. CA"`
		SearchLang string `json:"search_lang,omitempty"`
		SafeSearch string `json:"safesearch,omitempty" jsonschema:"off, moderate, or strict"`
		Freshness  string `json:"freshness,omitempty" jsonschema:"pd, pw, pm, py, or date range"`
		After      string `json:"after,omitempty" jsonschema:"YYYY-MM-DD freshness start"`
		Before     string `json:"before,omitempty" jsonschema:"YYYY-MM-DD freshness end"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "web_search",
		Description: "Brave Search plan API (web/news/video). Requires BRAVE_API_KEY Search subscription token.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args webArgs) (*mcp.CallToolResult, any, error) {
		client := &brave.Client{APIKey: opts.BraveAPIKey}
		k := backends.SearchKind(args.Kind)
		if k == "" {
			k = backends.SearchKindWeb
		}
		sq := backends.SearchQuery{
			Query: args.Query, Limit: args.Limit, Offset: args.Offset,
			Kind: k, Country: args.Country, SearchLang: args.SearchLang,
			SafeSearch: args.SafeSearch, Freshness: args.Freshness,
		}
		if t, err := localsearch.ParseDate(args.After); err == nil {
			sq.After = t
		}
		if t, err := localsearch.ParseDate(args.Before); err == nil {
			sq.Before = t
		}
		results, err := client.Search(ctx, sq)
		if err != nil {
			return nil, nil, err
		}
		b, _ := json.Marshal(results)
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}, results, nil
	})

	type emptyArgs struct{}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_catalogues",
		Description: "List catalogued YouTube channels in the local database",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ emptyArgs) (*mcp.CallToolResult, any, error) {
		chs, err := st.ListChannels(ctx)
		if err != nil {
			return nil, nil, err
		}
		b, _ := json.Marshal(chs)
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}, chs, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "doctor",
		Description: "Report catfu dependency and database health",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ emptyArgs) (*mcp.CallToolResult, any, error) {
		report := map[string]any{"version": version.Version, "db": opts.DBPath}
		yt := &catalogue.YTDLP{Bin: opts.YTDLP}
		if p, err := yt.LookPath(); err != nil {
			report["ytdlp"] = map[string]any{"ok": false, "error": err.Error()}
		} else {
			ver, _ := yt.Version(ctx)
			report["ytdlp"] = map[string]any{"ok": true, "path": p, "version": ver}
		}
		stats, err := st.Stats(ctx)
		if err != nil {
			report["database"] = map[string]any{"ok": false, "error": err.Error()}
		} else {
			report["database"] = map[string]any{"ok": true, "channels": stats.Channels, "videos": stats.Videos}
		}
		report["brave_api_key_set"] = opts.BraveAPIKey != ""
		b, _ := json.Marshal(report)
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}, report, nil
	})

	log.Info("catfu MCP server starting on stdio", "version", version.Version)
	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		return fmt.Errorf("mcp server: %w", err)
	}
	return nil
}
