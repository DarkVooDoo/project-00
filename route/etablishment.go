package route

import (
	"net/http"
	"planify/model"
)

type EtablishmentRoute struct{
    User model.UserClaim

}

var EtablishmentHandler http.Handler = &EtablishmentRoute{}

func (e EtablishmentRoute) ServeHTTP(w http.ResponseWriter, r *http.Request){

    switch r.Method{
        default:
            e.Get(w, r)
    }
}

func (e EtablishmentRoute) Get(w http.ResponseWriter, r *http.Request){

}
