package dkron

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"

	"github.com/distribworks/dkron/v3/dkron/assets_ui"
	"github.com/gin-gonic/gin"
)

const uiPathPrefix = "ui/"

// UI registers UI specific routes on the gin RouterGroup.
func (h *HTTPTransport) UI(r *gin.RouterGroup) {
	// If we are visiting from a browser redirect to the dashboard
	r.GET("/", func(c *gin.Context) {
		switch c.NegotiateFormat(gin.MIMEHTML) {
		case gin.MIMEHTML:
			c.Redirect(http.StatusSeeOther, "/ui/")
		default:
			c.AbortWithStatus(http.StatusNotFound)
		}
	})

	ui := r.Group("/" + uiPathPrefix)

	a, err := assets_ui.Assets.Open("index.html")
	if err != nil {
		log.Fatal(err)
	}
	b, err := ioutil.ReadAll(a)
	if err != nil {
		log.Fatal(err)
	}
	t, err := template.New("index.html").Parse(string(b))
	if err != nil {
		log.Fatal(err)
	}
	h.Engine.SetHTMLTemplate(t)

	ui.GET("/*filepath", func(ctx *gin.Context) {
		p := ctx.Param("filepath")
		_, err := assets_ui.Assets.Open(p)
		if err == nil && p != "/" && p != "/index.html" {
			ctx.FileFromFS(p, assets_ui.Assets)
		} else {
			jobs, err := h.agent.Store.GetJobs(nil)
			if err != nil {
				log.Error(err)
			}
			var (
				totalJobs = len(jobs)
				successfulJobs, failedJobs, untriggeredJobs int
			)
			for _, j := range jobs {
				if j.Status == "success" {
					successfulJobs++
				} else if j.Status == "failed" {
					failedJobs++
				} else if j.Status == "" {
					untriggeredJobs++
				}
			}
			l, err := h.agent.leaderMember()
			ln := "no leader"
			if err != nil {
				log.Error(err)
			} else {
				ln = l.Name
			}
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"DKRON_API_URL":         	fmt.Sprintf("/%s", apiPathPrefix),
				"DKRON_LEADER":          	ln,
				"DKRON_TOTAL_JOBS":      	totalJobs,
				"DKRON_FAILED_JOBS":     	failedJobs,
				"DKRON_UNTRIGGERED_JOBS":	untriggeredJobs,
				"DKRON_SUCCESSFUL_JOBS": 	successfulJobs,
			})
		}
	})
}