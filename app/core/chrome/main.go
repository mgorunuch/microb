package chrome

import (
	"context"
	"time"

	"github.com/chromedp/chromedp"
)

type Controller struct {
	chromeCtx    context.Context
	chromeCancel context.CancelFunc
}

var windowOptions = []chromedp.ExecAllocatorOption{
	chromedp.Flag("headless", false),
	chromedp.Flag("disable-gpu", false),
	chromedp.WindowSize(1200, 800),
}

func NewWindow() *Controller {
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), windowOptions...)
	return &Controller{
		chromeCtx:    ctx,
		chromeCancel: cancel,
	}
}

func (c *Controller) Close() {
	c.chromeCancel()
}

func (c *Controller) NewContext() *Context {
	ctx, cancel := chromedp.NewContext(c.chromeCtx)
	return &Context{
		chromeCtx:    ctx,
		chromeCancel: cancel,
	}
}

type Context struct {
	chromeCtx    context.Context
	chromeCancel context.CancelFunc
}

func (c *Context) Close() {
	c.chromeCancel()
}

func (c *Context) NewContext() *Context {
	ctx, cancel := chromedp.NewContext(c.chromeCtx)
	return &Context{
		chromeCtx:    ctx,
		chromeCancel: cancel,
	}
}

func (c *Context) NewContextWithTimeout(timeout time.Duration) *Context {
	ctx, cancel := chromedp.NewContext(c.chromeCtx)
	ctx, timeoutCancel := context.WithTimeout(ctx, timeout)

	return &Context{
		chromeCtx: ctx,
		chromeCancel: func() {
			timeoutCancel()
			cancel()
		},
	}
}

func (c *Context) Run(actions ...chromedp.Action) error {
	return chromedp.Run(c.chromeCtx, actions...)
}
