package flash

import (
	"fmt"
	"net/http"

	"github.com/goravel/framework/facades"
	"github.com/gorilla/sessions"
)

// Store cookie storage
var Store = sessions.NewCookieStore([]byte("secret-password"))

//var sessionName = "flash-session"

// GetCurrentUserName returns the username of the logged in user
func GetCurrentUserName(r *http.Request) string {
	session, err := Store.Get(r, "session")
	if err == nil {
		return session.Values["username"].(string)
	}
	return ""
}

// SetFlashMessage set flash msg
func SetFlashMessage(w http.ResponseWriter, r *http.Request, name string, value string) {
	session, err := Store.Get(r, flashName)
	if err != nil {
		facades.Log().Warningf("[session] set flash message err: %v", err)
	}
	session.AddFlash(value, name)
	err = session.Save(r, w)
	if err != nil {
		facades.Log().Warningf("[session] session save err: %v", err)
	}
}

// GetFlashMessage get flash msg from session
func GetFlashMessage(w http.ResponseWriter, r *http.Request, name string) string {
	session, err := Store.Get(r, flashName)
	if err != nil {
		facades.Log().Warningf("[session] get flash message err: %v", err)
		return ""
	}

	fm := session.Flashes(name)
	if fm == nil {
		fmt.Fprint(w, "No flash messages")
		return ""
	}
	_ = session.Save(r, w)
	_, _ = fmt.Fprintf(w, "%v", fm[0])

	return fm[0].(string)
}
