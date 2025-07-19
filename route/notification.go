package route

import (
	"log"
	"net/http"
	"planify/model"
)

type ViewNotification struct{
	User model.UserClaim
	Navbar model.CacheNavbar
}

var ViewNotificationHandler http.Handler = ViewNotification{}

func (vn ViewNotification) ServeHTTP(w http.ResponseWriter, r *http.Request){
	switch r.Method{
		default:
			vn.Get(w, r)
	}
}

func (vn ViewNotification) Get(w http.ResponseWriter, r *http.Request){
	if err := VerifyToken(r, w, &vn.User); err != nil{
		w.Header().Add("Location", "/connexion")
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	}
	conn := model.GetDBPoolConn()
	defer conn.Close()
	vn.Navbar = model.GetNavbarFromCache(conn, vn.User)
	if err := CreatePage(vn, w, "view/page.html", "view/notification.tmpl"); err != nil{
		log.Printf("error creating the page notification: %s", err)
		return
	}
}
