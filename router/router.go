package router

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/drone/drone/controller"
	"github.com/drone/drone/router/middleware/cache"
	"github.com/drone/drone/router/middleware/header"
	"github.com/drone/drone/router/middleware/location"
	"github.com/drone/drone/router/middleware/session"
	"github.com/drone/drone/router/middleware/token"
	"github.com/drone/drone/shared/envconfig"
	"github.com/drone/drone/static"
	"github.com/drone/drone/template"
)

func Load(env envconfig.Env, middleware ...gin.HandlerFunc) http.Handler {
	root := env.String("SERVER_ROOT", "")

	if strings.HasSuffix(root, "/") {
		root = strings.TrimSuffix(root, "/")
	}

	e := gin.Default()
	e.SetHTMLTemplate(template.Load())
	e.StaticFS(root+"/static", static.FileSystem())

	e.Use(location.Resolve)
	e.Use(header.NoCache)
	e.Use(header.Options)
	e.Use(header.Secure)
	e.Use(middleware...)
	e.Use(session.SetUser())
	e.Use(cache.Perms)
	e.Use(token.Refresh)

	e.GET(root+"/", cache.Repos, controller.ShowIndex)
	e.GET(root+"/login", controller.ShowLogin)
	e.GET(root+"/login/form", controller.ShowLoginForm)
	e.GET(root+"/logout", controller.GetLogout)

	settings := e.Group(root + "/settings")
	{
		settings.Use(session.MustUser())
		settings.GET("/profile", controller.ShowUser)
		settings.GET("/people", session.MustAdmin(), controller.ShowUsers)
		settings.GET("/nodes", session.MustAdmin(), controller.ShowNodes)
	}
	repo := e.Group(root + "/repos/:owner/:name")
	{
		repo.Use(session.SetRepo())
		repo.Use(session.SetPerm())
		repo.Use(session.MustPull)

		repo.GET("", controller.ShowRepo)
		repo.GET("/builds/:number", controller.ShowBuild)
		repo.GET("/builds/:number/:job", controller.ShowBuild)
		repo_settings := repo.Group("/settings")
		{
			repo_settings.GET("", session.MustPush, controller.ShowRepoConf)
			repo_settings.GET("/encrypt", session.MustPush, controller.ShowRepoEncrypt)
			repo_settings.GET("/badges", controller.ShowRepoBadges)
		}
	}

	user := e.Group(root + "/api/user")
	{
		user.Use(session.MustUser())
		user.GET("", controller.GetSelf)
		user.GET("/feed", controller.GetFeed)
		user.GET("/repos", cache.Repos, controller.GetRepos)
		user.GET("/repos/remote", cache.Repos, controller.GetRemoteRepos)
		user.POST("/token", controller.PostToken)
	}

	users := e.Group(root + "/api/users")
	{
		users.Use(session.MustAdmin())
		users.GET("", controller.GetUsers)
		users.POST("", controller.PostUser)
		users.GET("/:login", controller.GetUser)
		users.PATCH("/:login", controller.PatchUser)
		users.DELETE("/:login", controller.DeleteUser)
	}

	nodes := e.Group(root + "/api/nodes")
	{
		nodes.Use(session.MustAdmin())
		nodes.GET("", controller.GetNodes)
		nodes.POST("", controller.PostNode)
		nodes.DELETE("/:node", controller.DeleteNode)
	}

	repos := e.Group(root + "/api/repos/:owner/:name")
	{
		repos.POST("", controller.PostRepo)

		repo := repos.Group("")
		{
			repo.Use(session.SetRepo())
			repo.Use(session.SetPerm())
			repo.Use(session.MustPull)

			repo.GET("", controller.GetRepo)
			repo.GET("/key", controller.GetRepoKey)
			repo.GET("/builds", controller.GetBuilds)
			repo.GET("/builds/:number", controller.GetBuild)
			repo.GET("/logs/:number/:job", controller.GetBuildLogs)

			// requires authenticated user
			repo.POST("/encrypt", session.MustUser(), controller.PostSecure)

			// requires push permissions
			repo.PATCH("", session.MustPush, controller.PatchRepo)
			repo.DELETE("", session.MustPush, controller.DeleteRepo)

			repo.POST("/builds/:number", session.MustPush, controller.PostBuild)
			repo.DELETE("/builds/:number/:job", session.MustPush, controller.DeleteBuild)
		}
	}

	badges := e.Group(root + "/api/badges/:owner/:name")
	{
		badges.GET("/status.svg", controller.GetBadge)
		badges.GET("/cc.xml", controller.GetCC)
	}

	e.POST(root+"/hook", controller.PostHook)
	e.POST(root+"/api/hook", controller.PostHook)

	stream := e.Group(root + "/api/stream")
	{
		stream.Use(session.SetRepo())
		stream.Use(session.SetPerm())
		stream.Use(session.MustPull)
		stream.GET("/:owner/:name", controller.GetRepoEvents)
		stream.GET("/:owner/:name/:build/:number", controller.GetStream)
	}

	auth := e.Group(root + "/authorize")
	{
		auth.GET("", controller.GetLogin)
		auth.POST("", controller.GetLogin)
		auth.POST("/token", controller.GetLoginToken)
	}

	gitlab := e.Group(root + "/gitlab/:owner/:name")
	{
		gitlab.Use(session.SetRepo())
		gitlab.GET("/commits/:sha", controller.GetCommit)
		gitlab.GET("/pulls/:number", controller.GetPullRequest)

		redirects := gitlab.Group("/redirect")
		{
			redirects.GET("/commits/:sha", controller.RedirectSha)
			redirects.GET("/pulls/:number", controller.RedirectPullRequest)
		}
	}

	return normalize(root, e)
}

// normalize is a helper function to work around the following
// issue with gin. https://github.com/gin-gonic/gin/issues/388
func normalize(root string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, root), "/")[1:]

		if len(parts) > 0 {
			switch parts[0] {
			case "settings", "api", "login", "logout", "", "authorize", "hook", "static", "gitlab":
				// no-op
			default:
				if len(parts) > 2 && parts[2] != "settings" {
					parts = append(parts[:2], append([]string{"builds"}, parts[2:]...)...)
				}

				// prefix the URL with /repo so that it
				// can be effectively routed.
				parts = append([]string{root, "repos"}, parts...)

				// reconstruct the path
				r.URL.Path = strings.Join(parts, "/")
			}
		}

		h.ServeHTTP(w, r)
	})
}
