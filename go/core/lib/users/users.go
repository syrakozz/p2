package users

import (
	"context"
	"errors"
	"log/slog"

	"cloud.google.com/go/firestore"
	"golang.org/x/crypto/bcrypt"

	"disruptive/config"
	"disruptive/lib/common"
)

type _user struct {
	Username     string   `firestore:"username"`
	Name         string   `firestore:"name"`
	Passhash     string   `firestore:"passhash"`
	Permissions  []string `firestore:"permissions"`
	LoginSession string   `firestore:"login_session"`
}

// User contains Firestore user request document.
type User struct {
	Username     string   `json:"username"`
	Name         string   `json:"name"`
	Password     string   `json:"password,omitempty"`
	Permissions  []string `json:"permissions"`
	LoginSession string   `json:"login_session"`
}

func renderFromUser(user User) _user {
	return _user{
		Username:    user.Username,
		Name:        user.Name,
		Permissions: user.Permissions,
	}
}

func renderToUser(user _user) User {
	return User{
		Username:    user.Username,
		Name:        user.Name,
		Permissions: user.Permissions,
	}
}

// Add creates a new user in Firestore
func Add(ctx context.Context, logCtx *slog.Logger, user User) error {
	logCtx = logCtx.With("fid", "users.Add")

	passhash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		logCtx.Error("unable to generate password hash", "error", err)
		return err
	}

	client, err := firestore.NewClient(ctx, config.VARS.FirebaseProject)
	if err != nil {
		logCtx.Error("unable to create firestore client", "error", err)
		return err
	}
	defer client.Close()

	u := renderFromUser(user)
	u.Passhash = string(passhash)

	if _, err := client.Collection("users").Doc(u.Username).Create(ctx, u); err != nil {
		err = common.ConvertGRPCError(err)
		if errors.Is(err, common.ErrAlreadyExists{}) {
			logCtx.Error("user already exists")
			return err
		}
		logCtx.Error("unable to create user document", "error", err)
		return err
	}

	return nil
}

// GetUsersByPermission returns all users that contain a permission.
func GetUsersByPermission(ctx context.Context, logCtx *slog.Logger, permission string) ([]User, error) {
	logCtx = logCtx.With("fid", "users.GetUsersByService")

	client, err := firestore.NewClient(ctx, config.VARS.FirebaseProject)
	if err != nil {
		logCtx.Error("unable to create firestore client", "error", err)
		return nil, err
	}
	defer client.Close()

	docs, err := client.CollectionGroup("users").
		Where("permissions", "array-contains", permission).
		Documents(ctx).
		GetAll()
	if err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to query documents", "error", err)
		return nil, err
	}

	users := make([]User, len(docs))

	for i, doc := range docs {
		u := _user{}
		if err := doc.DataTo(&u); err != nil {
			err = common.ConvertGRPCError(err)
			logCtx.Error("unable to read user data", "error", err)
			return nil, err
		}
		users[i] = renderToUser(u)
	}

	return users, nil
}

func get(ctx context.Context, logCtx *slog.Logger, username string) (_user, error) {
	logCtx = logCtx.With("fid", "users.get")

	client, err := firestore.NewClient(ctx, config.VARS.FirebaseProject)
	if err != nil {
		logCtx.Error("unable to create firestore client", "error", err)
		return _user{}, err
	}
	defer client.Close()

	doc, err := client.Collection("users").Doc(username).Get(ctx)
	if err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to get user document", "error", err)
		return _user{}, err
	}

	u := _user{}
	if err := doc.DataTo(&u); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to read user data", "error", err)
		return _user{}, err
	}

	return u, nil
}

// Get retrieves a user from Firestore
func Get(ctx context.Context, logCtx *slog.Logger, username string) (User, error) {
	logCtx = logCtx.With("fid", "users.Get")
	u, err := get(ctx, logCtx, username)
	return renderToUser(u), err
}

// Modify modifies a user in Firestore.
func Modify(ctx context.Context, logCtx *slog.Logger, user User) error {
	logCtx = logCtx.With("fid", "users.Modify")

	u, err := get(ctx, logCtx, user.Username)
	if err != nil {
		return err
	}

	update := []firestore.Update{}

	if user.Name != "" {
		update = append(update, firestore.Update{Path: "name", Value: user.Name})
	}

	if user.Password != "" {
		passhash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			logCtx.Error("unable to generate password hash", "error", err)
			return err
		}
		update = append(update, firestore.Update{Path: "passhash", Value: string(passhash)})
	}

	// permissions
	if user.Permissions != nil {
		permsMap := map[string]bool{}
		permsList := []string{}

		for _, p := range u.Permissions {
			permsMap[p] = true
		}

		for _, p := range user.Permissions {
			if len(p) < 2 {
				continue
			}

			if p[0] == '-' {
				delete(permsMap, p[1:])
				continue
			}

			permsMap[p] = true
		}

		for p := range permsMap {
			permsList = append(permsList, p)
		}

		update = append(update, firestore.Update{Path: "permissions", Value: permsList})
	}

	client, err := firestore.NewClient(ctx, config.VARS.FirebaseProject)
	if err != nil {
		logCtx.Error("unable to create firestore client", "error", err)
		return err
	}
	defer client.Close()

	if _, err = client.Collection("users").Doc(user.Username).Update(ctx, update); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to update user document", "error", err)
		return err
	}

	return nil
}

// Delete removes a user in Firestore.
func Delete(ctx context.Context, logCtx *slog.Logger, username string) error {
	logCtx = logCtx.With("fid", "users.Delete")

	client, err := firestore.NewClient(ctx, config.VARS.FirebaseProject)
	if err != nil {
		return err
	}
	defer client.Close()

	if _, err := client.Collection("users").Doc(username).Delete(ctx); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to delete firestore client", "error", err)
		return err
	}

	return nil
}

// VerifyPassword validates a user's password in Firestore.
func VerifyPassword(ctx context.Context, logCtx *slog.Logger, username, password string) (bool, error) {
	u, err := get(ctx, logCtx, username)
	if err != nil {
		logCtx.Error("unable to get user", "error", err, "username", username)
		return false, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Passhash), []byte(password)); err != nil {
		return false, err
	}

	return true, nil
}

// GetAndVerifyPassword validates a user's password in Firestore and return the user.
func GetAndVerifyPassword(ctx context.Context, logCtx *slog.Logger, username, password string) (User, error) {
	u, err := get(ctx, logCtx, username)
	if err != nil {
		logCtx.Error("unable to get user", "error", err, "username", username)
		return User{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Passhash), []byte(password)); err != nil {
		return User{}, err
	}

	return renderToUser(u), nil
}

// SetLoginSession sets the user's login session value.
// To logout, call this function with loginSession set to empty string.
func SetLoginSession(ctx context.Context, logCtx *slog.Logger, username, loginSession string) error {
	logCtx = logCtx.With("fid", "users.SetLoginSession")

	client, err := firestore.NewClient(ctx, config.VARS.FirebaseProject)
	if err != nil {
		logCtx.Error("unable to create firestore client", "error", err)
		return err
	}
	defer client.Close()

	update := []firestore.Update{{Path: "login_session", Value: loginSession}}

	if _, err = client.Collection("users").Doc(username).Update(ctx, update); err != nil {
		err = common.ConvertGRPCError(err)
		logCtx.Error("unable to update user document", "error", err)
		return err
	}

	return nil
}

func verifyLoginSession(ctx context.Context, logCtx *slog.Logger, username, loginSession string) bool {
	u, err := get(ctx, logCtx, username)
	if err != nil {
		return false
	}

	return u.LoginSession == loginSession
}
