package model

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"mime/multipart"
	"os"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type CacheNavbar struct{
	Name string
	Email string
	Employee []KeyValue
	Etablishment []KeyValue
}

var navbarCache = make(map[int64]CacheNavbar)

func (u *User) Create() error{
    conn := GetDBPoolConn()
    defer conn.Close()
    cryptPassword, salt := cryptPassword(u.Password)
    userRow := conn.QueryRowContext(context.Background(), "INSERT INTO users(firstname, lastname, email, salt, password) VALUES($1,$2,$3,$4,$5)", u.Firstname, u.Lastname, u.Email, salt, cryptPassword)
    if err := userRow.Err(); err != nil{
        log.Printf("error query: %s", err)
        return errors.New("error creating user")
    }
    return nil
}

func (u *User) UploadPhoto(file multipart.File, contentType string)error{
    _, err := S3Client().PutObject(context.Background(), &s3.PutObjectInput{
        Bucket: aws.String("rdv-da"),
        ContentType: aws.String(contentType),
        Key: aws.String(string(u.Id)),
        Body: file,
    })
    if err != nil{
        log.Printf("error uploading the object: %s", err)
        return errors.New("error uploading the photo")
    }
    conn := GetDBPoolConn()
    defer conn.Close()
    result, err := conn.ExecContext(context.Background(), `UPDATE users SET picture=$1 WHERE id=$2`, fmt.Sprintf("%s/%s", "https://d1fzr3lngay3vb.cloudfront.net", u.Id), u.Id)
    if err != nil{
        log.Printf("error in the query: %s", err)
        return errors.New("error database query")
    }
    affected, err := result.RowsAffected()
    if affected == 0 || err != nil{
        log.Printf("error zero row affected")
        return errors.New("nothings happend")
    }
    u.Picture = fmt.Sprintf("%s/%s", "https://d1fzr3lngay3vb.cloudfront.net", u.Id)
    return nil
}

func (u *User) Profile(conn *sql.Conn)error{
    var town, postal, phone, picture sql.NullString
	var lat, lon sql.NullFloat64
    profileRow := conn.QueryRowContext(context.Background(), `SELECT firstname, lastname, email, town, postal, geolocation[0], geolocation[1], phone, picture, TO_CHAR(created_at, 'DD TMMonth YYYY') 
    FROM users WHERE id=$1`, u.Id)
    if err := profileRow.Scan(&u.Firstname, &u.Lastname, &u.Email, &town, &postal, &lat, &lon, &phone, &picture, &u.Joined); err != nil{
        log.Printf("error scaning the row: %s", err)
        return errors.New("error scanning user profile")
    }
    u.Town = town.String
    u.Postal = postal.String
    u.Phone = phone.String
    u.Lat = lat.Float64
    u.Lon = lon.Float64
    u.Picture = picture.String
    return nil
}

func (u *User) Modify()error{
    conn := GetDBPoolConn()
    defer conn.Close()
    modifyProfile, err := conn.ExecContext(context.Background(), `UPDATE users SET firstname=$1, lastname=$2, town=$3, postal=$4, phone=$5, geolocation=POINT($6,$7) WHERE id=$8`, 
    u.Firstname, u.Lastname, u.Town, u.Postal, u.Phone, u.Lat, u.Lon, u.Id)
    if err != nil{
        log.Printf("error in the query: %s", err)
        return errors.New("error in the query")
    }
    affected, err := modifyProfile.RowsAffected()
    if affected == 0 || err != nil{
        log.Printf("error modifying profile: %s", err)
        return errors.New("update profile error")
    }
    return nil
}

func (u *User) Sign(email string, password string)(error){

    conn := GetDBPoolConn()
    defer conn.Close()
    userRow := conn.QueryRowContext(context.Background(), `SELECT * FROM SignUser($1)`, email)
    if err := userRow.Scan(&u.Id, &u.ShortName, &u.Picture, &u.EmployeeId, &u.EtablishmentId, &u.Salt, &u.Password); err != nil{
        log.Printf("error scanning user: %s", err)
        return errors.New("error selecting user")
    }
	//cryptedPassword := verifyPassword(password, u.Salt)
	//if cryptedPassword != u.Password{
	//	log.Printf("error no the same password: DB: %v \tCrypted password: %v", u.Password, cryptedPassword)
	//	return errors.New("error no the same password")
	//}
    return nil
}

func verifyPassword(password string, salt int) string{
    cypher := sha256.Sum256([]byte(fmt.Sprintf("%v%v%v", password,salt,os.Getenv("PASSWORD_SECRET_KEY"))))
    return fmt.Sprintf("%x", cypher)
}

func cryptPassword(password string)(cryptPassword string, salt int){
    salt = rand.Intn(9999-1000) + 1000
    cypher := sha256.Sum256([]byte(fmt.Sprintf("%v%v%v", password,salt,os.Getenv("PASSWORD_SECRET_KEY"))))
    cryptPassword = fmt.Sprintf("%x", cypher)
    return cryptPassword, salt
}

func GetNavbarFromCache(conn *sql.Conn, u UserClaim)CacheNavbar{

	if reflect.DeepEqual(navbarCache[u.Id], CacheNavbar{}){
		return getNavbarInfo(conn, u)
	}
	return navbarCache[u.Id]
}

func getNavbarInfo(conn *sql.Conn, u UserClaim)CacheNavbar{
	var navigationCache CacheNavbar
	var valueKey KeyValue
	if u.Employee != 0{
		employeeIds, err := conn.QueryContext(context.Background(), `SELECT e.id, et.name FROM employee AS e LEFT JOIN etablishment AS et ON et.id=e.etablishment_id
		WHERE e.user_id=$1`, u.Id)
		if err != nil{
			log.Printf("error getting employee ids: %s", err)
		}
		for employeeIds.Next(){
			if err := employeeIds.Scan(&valueKey.Id, &valueKey.Value); err != nil{
				log.Printf("error scanning the employees: %s", err)
			}
			navigationCache.Employee = append(navigationCache.Employee, valueKey)
		}
	}
	if u.Etablishment != 0{
		etablishmentIds, err := conn.QueryContext(context.Background(), `SELECT et.id, et.name FROM etablishment AS et WHERE et.user_id=$1`, u.Id)
		if err != nil{
			log.Printf("error getting employee ids: %s", err)
		}
		for etablishmentIds.Next(){
			if err := etablishmentIds.Scan(&valueKey.Id, &valueKey.Value); err != nil{
				log.Printf("error scanning the employees: %s", err)
			}
			navigationCache.Etablishment = append(navigationCache.Etablishment, valueKey)
		}

	}
	user := conn.QueryRowContext(context.Background(), `SELECT firstname || ' ' || lastname, email FROM users WHERE id=$1`, u.Id)
	if err := user.Scan(&navigationCache.Name, &navigationCache.Email); err != nil{
		log.Printf("error getting navbar user info: %s", err)
	}
	return navigationCache
}
