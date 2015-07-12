package AngularBootstrapExample

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"appengine"
	"appengine/datastore"
	"appengine/memcache"
	"fmt"
	"encoding/json"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"strings"
)

type RegisterReqS struct {
	Username 		string	`json:"username"`
	Password 		string	`json:"password"`
	ConfirmPassword string	`json:"confirm_password"`
	Email 			string	`json:"email"`
}

type UserS struct {
	Username 		string	`json:"username"`
	Passhash 		string	`json:"passhash"`
	Email 			string	`json:"email"`
	Salt			string	`json:"salt"`
	Secret			string	`json:"secret"`
}

func RegisterHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	
	ctx := appengine.NewContext(r)
	
	d := json.NewDecoder(r.Body)
	var req_body_struct RegisterReqS
	err := d.Decode(&req_body_struct)
	if err != nil {
		err_msg := "400 Bad Request: Failed decoding register request: " + err.Error()
		http.Error(w, err_msg, 400)
		ctx.Errorf("%v: %v", err_msg, err)
		return
	}
	
	if req_body_struct.Username == "" 		 || 
	   req_body_struct.Password == "" 		 || 
	   req_body_struct.ConfirmPassword == "" || 
	   req_body_struct.Email == "" {
		
		err_msg := "400 Bad Request: All fields must be filled in."
		http.Error(w, err_msg, 400)
		ctx.Errorf("%v", err_msg)
		return
	}
	
	if req_body_struct.Password != req_body_struct.ConfirmPassword {
		err_msg := "400 Bad Request: Passwords don't match."
		http.Error(w, err_msg, 400)
		ctx.Errorf("%v", err_msg)
		return
	}
	
	// Check if username already exists in memcache
	_, err = memcache.Get(ctx, strings.ToLower(req_body_struct.Username))
	if err == memcache.ErrCacheMiss {
	
		// Check if username already exists in datastore
		q := datastore.NewQuery("user").Filter("Username =", strings.ToLower(req_body_struct.Username))
		for t := q.Run(ctx); ; {
			
			var user UserS
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
			http.Error(w, "400 Bad Request: Username already registered (found in datastore). Please choose another.", 400)
			return
		}
		
	} else if err != nil {
		err_msg := "500 Internal Server Error: Failed checking memcache for user: " + err.Error()
		http.Error(w, err_msg, 500)
		ctx.Errorf("%v: %v", err_msg, err)
		return
	
	} else if err == nil {
		
		// User found
		http.Error(w, "400 Bad Request: Username already registered (found in memcache). Please choose another.", 400)
		return
	}
	
	ik := datastore.NewIncompleteKey(ctx, "user", nil)
	
	salt_len := 512
	salt_bytes := make([]byte, salt_len)
	_, err = rand.Read(salt_bytes)
	if err != nil {
		err_msg := "500 Internal Server Error: Failed creating salt for new user: " + err.Error()
		http.Error(w, err_msg, 500)
		ctx.Errorf("%v: %v", err_msg, err)
		return
	}
	
	h := sha512.New()
	h.Write(salt_bytes)
	salt := base64.URLEncoding.EncodeToString(h.Sum(nil))
	
	h = sha512.New()
	h.Write([]byte(req_body_struct.Password + salt))
	passhash := base64.URLEncoding.EncodeToString(h.Sum(nil))
	
	secret_len := 512
	secret_bytes := make([]byte, secret_len)
	_, err = rand.Read(secret_bytes)
	if err != nil {
		err_msg := "500 Internal Server Error: Failed creating secret for new user: " + err.Error()
		http.Error(w, err_msg, 500)
		ctx.Errorf("%v: %v", err_msg, err)
		return
	}
	
	h = sha512.New()
	h.Write(secret_bytes)
	secret := base64.URLEncoding.EncodeToString(h.Sum(nil))
	
	e := &UserS{
		Username: 	strings.ToLower(req_body_struct.Username),
		Passhash:	passhash,
		Email: 		req_body_struct.Email,
		Salt:		salt,
		Secret:		secret,
	}
	
	//ctx.Infof("%+v", e)
	
	// Put new user into memcache
	e_json, err := json.Marshal(e)
	if err != nil {
		err_msg := "500 Internal Server Error: Failed marshaling new user to json: " + err.Error()
		http.Error(w, err_msg, 500)
		ctx.Errorf("%v: %v", err_msg, err)
		return
	}
	
	item := &memcache.Item{
		Key: e.Username,
		Value: e_json,
	}
	
	err = memcache.Set(ctx, item)
	if err != nil {
		err_msg := "500 Internal Server Error: Failed putting new user into memcache: " + err.Error()
		http.Error(w, err_msg, 500)
		ctx.Errorf("%v: %v", err_msg, err)
		return
	}
	
	// Put new user into datastore
	k, err := datastore.Put(ctx, ik, e)
	if err != nil {
		err_msg := "500 Internal Server Error: Failed adding new user to datastore: " + err.Error()
		http.Error(w, err_msg, 500)
		ctx.Errorf("%v: %v", err_msg, err)
		return
	}
	_ = k
	
	fmt.Fprintf(w, "success")
}