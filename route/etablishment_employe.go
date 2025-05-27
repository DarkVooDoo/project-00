package route

import (
	"net/http"
)

type EmployeRoute struct{
    
}

var EmployeHandler http.Handler = EmployeRoute{}

func (ee EmployeRoute) ServeHTTP(w http.ResponseWriter, r *http.Request){
}
