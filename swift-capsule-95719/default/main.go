package AngularBootstrapExample

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"html/template"
	"appengine"
)

func init() {
	
	r := httprouter.New()
	r.ServeFiles("/static/*filepath", http.Dir("static"))
	r.POST("/register", RegisterHandler)
	r.POST("/log_in", LogInHandler)
	r.GET("/", IndexHandler)
	http.Handle("/", r)
}

func IndexHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	
	ctx := appengine.NewContext(r)
	
	var err error
	
	p := IndexTemplateS{}
	
	t := template.New("index.html")
	t = t.Delims("[[", "]]")
	
	t, err = t.ParseFiles("templates/index.html")
	if err != nil {
		ctx.Errorf("%v", err)
		return
	}
	
	err = t.Execute(w, p)
	if err != nil {
		ctx.Errorf("%v", err)
		return
	}
}
