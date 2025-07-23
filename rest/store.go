package rest

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"sort"
	"strconv"
)

var ID = func() string { return rand.Text() }
var HashPasswd = func(passwd, salt string) string {
	sum := sha256.Sum256([]byte(salt + passwd))
	return base32.StdEncoding.EncodeToString(sum[:])
}

type Store struct {
	Dir       string
	Schemas   map[string]Schema
	Resources map[string]DB
}

func NewStore(dir string) (*Store, error) {
	s := &Store{Dir: dir, Schemas: map[string]Schema{}, Resources: map[string]DB{}}
	schemaDB, err := NewCSVDB(s.Dir + "/_schemas.csv")
	if err != nil {
		return nil, err
	}

	for rec, err := range schemaDB.Iter() {
		if err != nil {
			return nil, err
		}
		if len(rec) != 8 {
			return nil, fmt.Errorf("invalid schema record: %v", rec)
		}
		schema := FieldSchema{
			Resource: rec[2],
			Field:    rec[3],
			Type:     FieldType(rec[4]),
			Regex:    rec[7],
		}
		schema.Min, _ = strconv.ParseFloat(rec[5], 64)
		schema.Max, _ = strconv.ParseFloat(rec[6], 64)
		s.Schemas[schema.Resource] = append(s.Schemas[schema.Resource], schema)
		if _, ok := s.Resources[schema.Resource]; !ok {
			db, err := NewCSVDB(s.Dir + "/" + schema.Resource + ".csv")
			if err != nil {
				return nil, err
			}
			s.Resources[schema.Resource] = db
		}
	}
	return s, nil
}

func (s *Store) Create(resource string, r Resource) (string, error) {
	db, ok := s.Resources[resource]
	if !ok {
		return "", fmt.Errorf("resource %s not found", resource)
	}
	newID := ID()
	r["_id"] = newID
	r["_v"] = 1.0
	rec, err := s.Schemas[resource].Record(r)
	if err != nil {
		return "", err
	}
	if err := db.Create(rec); err != nil {
		return "", err
	}
	return newID, nil
}

func (s *Store) Update(resource string, r Resource) error {
	db, ok := s.Resources[resource]
	if !ok {
		return fmt.Errorf("resource %s not found", resource)
	}
	orig, err := s.Get(resource, r["_id"].(string))
	if err != nil {
		return fmt.Errorf("record not found: %w", err)
	}
	for _, field := range s.Schemas[resource] {
		if _, ok := r[field.Field]; !ok {
			r[field.Field] = orig[field.Field]
		}
	}
	r["_v"] = orig["_v"].(float64) + 1
	rec, err := s.Schemas[resource].Record(r)
	if err != nil {
		return err
	}
	return db.Update(rec)
}

func (s *Store) Delete(resource, id string) error {
	db, ok := s.Resources[resource]
	if !ok {
		return fmt.Errorf("resource %s not found", resource)
	}
	return db.Delete(id)
}

func (s *Store) Get(resource, id string) (Resource, error) {
	db, ok := s.Resources[resource]
	if !ok {
		return nil, fmt.Errorf("resource %s not found", resource)
	}
	rec, err := db.Get(id)
	if err != nil {
		return nil, err
	}
	if len(rec) < 2 {
		return nil, nil // record not found
	}
	return s.Schemas[resource].Resource(rec)
}

func (s *Store) List(resource, sortBy string) ([]Resource, error) {
	db, ok := s.Resources[resource]
	if !ok {
		return nil, fmt.Errorf("resource %s not found", resource)
	}
	res := []Resource{}
	for rec, err := range db.Iter() {
		if err != nil {
			return nil, err
		}
		if len(rec) < 2 {
			continue
		}
		r, err := s.Schemas[resource].Resource(rec)
		if err != nil {
			return res, err
		}
		res = append(res, r)
	}
	if sortBy != "" {
		sort.Slice(res, func(i, j int) bool {
			if res[i][sortBy] == nil {
				return false
			}
			if res[j][sortBy] == nil {
				return true
			}
			switch res[i][sortBy].(type) {
			case string:
				return res[i][sortBy].(string) < res[j][sortBy].(string)
			case float64:
				return res[i][sortBy].(float64) < res[j][sortBy].(float64)
			default:
				return false
			}
		})
	}
	return res, nil
}

func (s *Store) Close() error {
	for _, db := range s.Resources {
		if err := db.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) Authenticate(r *http.Request) (Resource, error) {
	if cookie, err := r.Cookie("session"); err == nil {
		if username, ok := VerifySession(cookie.Value); ok {
			u, err := s.Get("_users", username)
			if err != nil {
				return nil, fmt.Errorf("users error: %w", err)
			}
			return u, nil
		}
	}
	if username, password, ok := r.BasicAuth(); ok {
		return s.AuthenticateBasic(username, password)
	}
	return nil, errors.New("unauthenticated")
}

func (s *Store) AuthenticateBasic(username, password string) (Resource, error) {
	u, err := s.Get("_users", username)
	if err != nil {
		return nil, fmt.Errorf("users error: %w", err)
	}
	if u["password"] != HashPasswd(password, u["salt"].(string)) {
		return nil, errors.New("unauthenticated")
	}
	return u, nil
}

func (s *Store) Authorize(resource, id, action string, user Resource) error {
	permissions, err := s.List("_permissions", "")
	if err != nil {
		return fmt.Errorf("permissions error: %w", err)
	}
	for _, p := range permissions {
		if p["resource"] != resource || (p["action"] != "*" && p["action"] != action) {
			continue
		}
		if p["field"] == "" && p["role"] == "" { // public
			return nil
		}
		if user == nil {
			return errors.New("unauthenticated")
		}
		// Any role? Or user has the role?
		if p["role"] == "*" || slices.Contains(user["roles"].([]string), p["role"].(string)) {
			return nil
		}
		if id != "" {
			res, err := s.Get(resource, id)
			if err != nil {
				return err
			}
			username := user["_id"].(string)
			if user, ok := res[p["field"].(string)]; ok && user == username {
				return nil // user name matches requested resource field (string)
			} else if users, ok := res[p["field"].(string)].([]string); ok && slices.Contains(users, username) {
				return nil // user name is in the requested resource field (list)
			}
		}
	}
	return errors.New("unauthorized")
}
