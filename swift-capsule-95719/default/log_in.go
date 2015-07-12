package AngularBootstrapExample

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"appengine"
	"appengine/datastore"
	"appengine/memcache"
	"fmt"
	"encoding/json"
	"crypto/sha512"
	"crypto/rand"
	"encoding/base64"
	"strings"
	"github.com/dgrijalva/jwt-go"
	"time"
	"math/big"
	"strconv"
)

type LoginReqS struct {
	Username	string	`json:"username"`
	Password	string	`json:"password"`
}

func LogInHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	
	ctx := appengine.NewContext(r)
	
	d := json.NewDecoder(r.Body)
	var req_body_struct LoginReqS
	err := d.Decode(&req_body_struct)
	if err != nil {
		err_msg := "400 Bad Request: Failed decoding login request: " + err.Error()
		http.Error(w, err_msg, 400)
		ctx.Errorf("%v: %v", err_msg, err)
		return
	}
	
	if req_body_struct.Username == "" || 
	   req_body_struct.Password == "" {
		
		err_msg := "400 Bad Request: All fields must be filled in."
		http.Error(w, err_msg, 400)
		ctx.Errorf("%v", err_msg)
		return
	}
	
	username_found := false
	var user UserS
	username := strings.ToLower(req_body_struct.Username)
	
	// Lookup username in memcache
	user_mc, err := memcache.Get(ctx, username)
	if err == memcache.ErrCacheMiss {
	
		ctx.Infof("User %v not found in memcache, checking datastore.", username)
	
		// Lookup username in datastore
		q := datastore.NewQuery("user").Filter("Username =", strings.ToLower(req_body_struct.Username))
		for t := q.Run(ctx); ; {
			
			key, err := t.Next(&user)
			_ = key
			if err == datastore.Done {
				break
			}
			if err != nil {
				err_msg := "500 Internal Server Error: Failed looking for user in datastore: " + err.Error()
				http.Error(w, err_msg, 500)
				ctx.Errorf("%v: %v", err_msg, err)
				return
			}
			
			// User found
			username_found = true
			
			user_json, err := json.Marshal(user)
			if err != nil {
				err_msg := "500 Internal Server Error: Failed marshaling user to json: " + err.Error()
				http.Error(w, err_msg, 500)
				ctx.Errorf("%v: %v", err_msg, err)
				return
			}
			
			err = memcache.Set(ctx, &memcache.Item{
				Key: strings.ToLower(user.Username),
				Value: user_json,
			})
			if err != nil {
				err_msg := "500 Internal Server Error: Failed putting user into memcache: " + err.Error()
				http.Error(w, err_msg, 500)
				ctx.Errorf("%v: %v", err_msg, err)
				return
			}
		}
		
	} else if err != nil {
		
		err_msg := "500 Internal Server Error: Failed checking memcache for user: " + err.Error()
		http.Error(w, err_msg, 500)
		ctx.Errorf("%v: %v", err_msg, err)
		return
	
	} else if err == nil {
		
		ctx.Infof("User %v found in memcache.", username)
		
		err = json.Unmarshal(user_mc.Value, &user)
		if err != nil {
			err_msg := "500 Internal Server Error: Failed unmarshaling user from json: " + err.Error()
			http.Error(w, err_msg, 500)
			ctx.Errorf("%v: %v", err_msg, err)
			return
		}
		
		username_found = true
	}
	
	if !username_found {
		err_msg := "401 Unauthorized: Username not found."
		http.Error(w, err_msg, 401)
		ctx.Errorf("%v", err_msg)
		return
	}
	
	// Check password
	h := sha512.New()
	h.Write([]byte(req_body_struct.Password + user.Salt))
	passhash := base64.URLEncoding.EncodeToString(h.Sum(nil)) 
	
	if user.Passhash != passhash {
		err_msg := "401 Unauthorized: Incorrect password."
		http.Error(w, err_msg, 401)
		ctx.Errorf("%v", err_msg)
		return
	}
	
	// Make jwt
	
	max := *big.NewInt(99999999999)
	jti_randnum, err := rand.Int(rand.Reader, &max)
	if err != nil {
		err_msg := "500 Internal Server Error: Failed getting random number: " + err.Error()
		http.Error(w, err_msg, 500)
		ctx.Errorf("%v: %v", err_msg, err)
		return
	}
	
	h = sha512.New()
	h.Write([]byte(strconv.FormatInt(time.Now().UnixNano() + jti_randnum.Int64(), 10)))
	jti := base64.URLEncoding.EncodeToString(h.Sum(nil)) 
	
	current_time := time.Now().Unix()
	token_valid_for := int64(30 * 60)	// 30 minutes
	auth_jwt := jwt.New(jwt.SigningMethodHS512)
	
	auth_jwt.Claims["iss"] = r.Host
	auth_jwt.Claims["sub"] = r.Host + "/jwt/user/" + user.Username
	auth_jwt.Claims["aud"] = []string{r.Host,}
	auth_jwt.Claims["exp"] = current_time + token_valid_for
	auth_jwt.Claims["nbf"] = current_time
	auth_jwt.Claims["iat"] = current_time
	auth_jwt.Claims["jti"] = jti
	auth_jwt.Claims[r.Host + "/jwt/claim/admin"] = false	// TODO: load this from datastore
	
	auth_jwt_string, err := auth_jwt.SignedString([]byte(user.Secret))
	if err != nil {
		err_msg := "500 Internal Server Error: Failed signing jwt: " + err.Error()
		http.Error(w, err_msg, 500)
		ctx.Errorf("%v: %v", err_msg, err)
		return
	}
	
	fmt.Fprintf(w, auth_jwt_string)
}
