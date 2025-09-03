package puzzle

import "15-puzzle/internal/model"

func (c *Controller) ApiResponseHandler(u model.ApiResponse) {
	switch {
	case u.Err != nil:
		c.ApiErrorHandler(*u.Err)
	case u.Info != nil:
		c.ApiInfoHandler(*u.Info)
	case u.Stats != nil:
		c.ApiStatsHandler(*u.Stats)
	case u.Monitoring != nil:
		c.ApiMonitoringHandler(*u.Monitoring)
	}
}

func (c *Controller) ApiInfoHandler(s model.Info) {
	for _, h := range c.screens {
		if i, ok := h.(interface{ ApiInfoHandler(model.Info) }); ok {
			i.ApiInfoHandler(s)
		}
	}
}

func (c *Controller) ApiStatsHandler(s model.Stats) {
	for _, h := range c.screens {
		if i, ok := h.(interface{ ApiStatsHandler(model.Stats) }); ok {
			i.ApiStatsHandler(s)
		}
	}
}

func (c *Controller) ApiMonitoringHandler(m model.Monitoring) {
	for _, h := range c.screens {
		if i, ok := h.(interface{ ApiMonitoringHandler(model.Monitoring) }); ok {
			i.ApiMonitoringHandler(m)
		}
	}
}

func (c *Controller) ApiErrorHandler(e string) {
	c.Debug("api error: %s", e)
}
