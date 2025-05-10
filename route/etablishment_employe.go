package route

import "net/http"

type EtablishmentEmployeRoute struct{
    
}

var EtablishmentEmployeHandler http.Handler = EtablishmentRoute{}

func (ee EtablishmentEmployeRoute) ServeHTTP(w http.ResponseWriter, r *http.Request){

}
