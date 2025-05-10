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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)


func (u *User) Create() error{
    var salt int
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
        Key: aws.String(u.Id),
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

func (u *User) Profile()error{
    var town, postal, phone, lat, lon, picture sql.NullString
    conn := GetDBPoolConn()
    defer conn.Close()
    profileRow := conn.QueryRowContext(context.Background(), `SELECT firstname, lastname, email, town, postal, latitude, longitude, phone, picture, TO_CHAR(created_at, 'DD TMMonth YYYY') 
    FROM users WHERE id=$1`, u.Id)
    if err := profileRow.Scan(&u.Firstname, &u.Lastname, &u.Email, &town, &postal, &lat, &lon, &phone, &picture, &u.Joined); err != nil{
        log.Printf("error scaning the row: %s", err)
        return errors.New("error scanning user profile")
    }
    u.Town = town.String
    u.Postal = postal.String
    u.Phone = phone.String
    u.Lat = lat.String
    u.Lon = lon.String
    u.Picture = picture.String
    return nil
}

func (u *User) Modify()error{
    conn := GetDBPoolConn()
    defer conn.Close()
    log.Print(u.Lat, u.Lon)
    modifyProfile, err := conn.ExecContext(context.Background(), `UPDATE users SET firstname=$1, lastname=$2, town=$3, postal=$4, phone=$5, latitude=$6, longitude=$7 WHERE id=$8`, 
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

func (u *User) Get(email string, password string)error{
    conn := GetDBPoolConn()
    defer conn.Close()
    userRow := conn.QueryRowContext(context.Background(), `SELECT firstname || ' ' || lastname FROM users WHERE email=$1`, email)
    if err := userRow.Scan(&u.Firstname, &u.Lastname); err != nil{
        log.Printf("error scanning user: %s", err)
        return errors.New("error selecting user")
    }
    log.Println(u.Firstname)
    return nil
}

func (u *UserClaim) Sign(email string, password string)error{
    var picture sql.NullString
    conn := GetDBPoolConn()
    defer conn.Close()
    userRow := conn.QueryRowContext(context.Background(), `SELECT * FROM SignUser($1)`, email)
    if err := userRow.Scan(&u.Id, &u.ShortName, &picture, &u.Employee, &u.Etablishment); err != nil{
        log.Printf("error scanning user: %s", err)
        return errors.New("error selecting user")
    }
    u.Picture = picture.String
    return nil
}

func cryptPassword(password string)(cryptPassword string, salt int){
    salt = rand.Intn(9999-1000) + 1000
    cypher := sha256.Sum256([]byte(fmt.Sprintf("%v%v%v", password,salt,os.Getenv("PASSWORD_SECRET_KEY"))))
    cryptPassword = fmt.Sprintf("%x", cypher)
    return cryptPassword, salt
}
