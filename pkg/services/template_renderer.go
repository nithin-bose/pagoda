package services

import (
	"context"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/mikestefanello/pagoda/config"
	"github.com/mikestefanello/pagoda/pkg/page"
)

type TemplCtxKey string

const TemplCtxKeyPage TemplCtxKey = "page"

// TemplateRenderer provides a flexible and easy to use method of rendering Templ templates while
// also providing caching and/or hot-reloading depending on your current environment
type TemplateRenderer struct {
	// For page caching
	*pageCache

	// config stores application configuration
	config *config.Config
}

// NewTemplateRenderer creates a new TemplRenderer
func NewTemplateRenderer(cfg *config.Config, cache *CacheClient) *TemplateRenderer {
	return &TemplateRenderer{
		pageCache: NewPageCache(cfg, cache),
		config:    cfg,
	}
}

// RenderPage renders a Page as an HTTP response
func (t *TemplateRenderer) RenderPage(ctx echo.Context, page *page.Page) error {
	var err error

	// Check if this is an HTMX non-boosted request which indicates that only partial
	// content should be rendered
	// if page.HTMX.Request.Enabled && !page.HTMX.Request.Boosted {
	// 	// Switch the layout which will only render the page content
	// 	page.Layout = templates.LayoutHTMX

	// 	// Alter the template group so this is cached separately
	// 	templateGroup = "page:htmx"
	// }

	// Set the status code
	ctx.Response().Status = page.StatusCode

	// Set any headers
	for k, v := range page.Headers {
		ctx.Response().Header().Set(k, v)
	}

	// Apply the HTMX response, if one
	if page.HTMX.Response != nil {
		page.HTMX.Response.Apply(ctx)
	}

	buf := templ.GetBuffer()
	defer templ.ReleaseBuffer(buf)

	templCtx := context.WithValue(ctx.Request().Context(), TemplCtxKeyPage, page)
	err = page.TemplComponent.Render(templCtx, buf)
	if err != nil {
		return err
	}

	// Cache this page, if caching was enabled
	t.cachePage(ctx, page, buf)

	return ctx.HTMLBlob(ctx.Response().Status, buf.Bytes())
}
