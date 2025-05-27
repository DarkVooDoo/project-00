package model

import (
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UserClaim struct{
    Id int `json:"id,string"`
    ShortName string `json:"short_name"`
    Picture string `json:"picture"`
    Etablishment int
    Employee int
    jwt.RegisteredClaims
}

type User struct{
    Id int
    Town string `json:"town"`
    Postal string `json:"postal"`
    Lat string `json:"lat"`
    Lon string `json:"lon"`
    Phone string `json:"phone"`
    Email string `json:"email"`
    Password string `json:"password"`
    Firstname string `json:"firstname"`
    Lastname string `json:"lastname"`
    Confirmation string `json:"confirmation"`
	Salt int
    Picture string
    Joined string
	ShortName string
    EtablishmentId int
    EmployeeId int
}

const (
    ACCESS_TOKEN_EXPIRE = time.Hour * 24 * 3
)

func CreateAccessToken(id int, shortname string, picture string, etablishment int, employee int, w http.ResponseWriter) error{
    claim := UserClaim{
        id,
        shortname,
        picture,
        etablishment,
        employee,
        jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(ACCESS_TOKEN_EXPIRE)),
            NotBefore: jwt.NewNumericDate(time.Now()),
            IssuedAt: jwt.NewNumericDate(time.Now()),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
    ss, err := token.SignedString([]byte(os.Getenv("ACCESS_TOKEN_KEY")))
    if err != nil{
        log.Printf("error signing the token: %s", err)
        return errors.New("error signing the token")
    }
    cookie := http.Cookie{
        Name: "access-token",
        Value: ss,
        HttpOnly: true,
        Secure: true,
        Expires: claim.ExpiresAt.Time,
        SameSite: http.SameSiteStrictMode,
        Path: "/",
    }
    http.SetCookie(w, &cookie)
    return nil
}

func (u *UserClaim)VerifyAccessToken(token string, w http.ResponseWriter)error{
    tk, err := jwt.ParseWithClaims(token, u, func(t *jwt.Token) (interface{}, error) {
        return []byte(os.Getenv("ACCESS_TOKEN_KEY")), nil
    })
    if err != nil{
        return errors.New("error parsing token")
    }else if _, ok := tk.Claims.(*UserClaim); ok {
        if time.Until(u.ExpiresAt.Time).Minutes() < 60{
            if err := CreateAccessToken(u.Id, u.ShortName, u.Picture, u.Etablishment, u.Employee, w); err != nil{
                log.Print("error recreating the token")
            }
        }
        return nil
    } else {
        return errors.New("unknown claims type, cannot proceed")
    }

}
