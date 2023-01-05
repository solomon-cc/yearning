// Copyright 2019 HenryYee.
//
// Licensed under the AGPL, Version 3.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    https://www.gnu.org/licenses/agpl-3.0.en.html
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package login

import (
	"Yearning-go/src/handler/commom"
	"Yearning-go/src/lib"
	"Yearning-go/src/model"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cookieY/yee"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"io/ioutil"
	"net/http"
)

type loginForm struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Config struct {
	State string `json:"state"`
	Code  string `json:"code"`
}

type GitlabUser struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	State    string `json:"state"`
	Email    string `json:"email"`
}

const (
	oauthStateString = "state"
	unauth           = ""
)

func UserLdapLogin(c yee.Context) (err error) {
	u := new(loginForm)
	if err = c.Bind(u); err != nil {
		return c.JSON(http.StatusOK, commom.ERR_REQ_BIND)
	}
	isOk, err := lib.LdapContent(&model.GloLdap, u.Username, u.Password, false)
	if err != nil {
		return c.JSON(http.StatusOK, commom.ERR_COMMON_MESSAGE(err))
	}
	if isOk {
		var account model.CoreAccount
		if model.DB().Where("username = ?", u.Username).First(&account).RecordNotFound() {
			model.DB().Create(&model.CoreAccount{
				Username:   u.Username,
				RealName:   "请重置你的真实姓名",
				Password:   lib.DjangoEncrypt(lib.GenWorkid(), string(lib.GetRandom())),
				Rule:       "guest",
				Department: "all",
				Email:      "",
			})
			ix, _ := json.Marshal([]string{})
			model.DB().Create(&model.CoreGrained{Username: u.Username, Group: ix})
		}
		token, tokenErr := lib.JwtAuth(u.Username, account.Rule)
		if tokenErr != nil {
			c.Logger().Error(tokenErr.Error())
			return
		}
		dataStore := map[string]string{
			"token":       token,
			"permissions": account.Rule,
			"real_name":   account.RealName,
		}
		return c.JSON(http.StatusOK, commom.SuccessPayload(dataStore))
	}
	return c.JSON(http.StatusOK, commom.ERR_LOGIN)
}

func UserGitlabLogin(c yee.Context) (err error) {
	conf := Setup()

	token := c.GetHeader("x-token")

	if token == unauth {
		// Redirect user to consent page to ask for permission
		// for the scopes specified above.
		url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)

		return c.JSON(http.StatusOK, commom.UserUnAuth(url))
	}

	return c.JSON(http.StatusMovedPermanently, commom.UserAuthed(model.GloBaseDomin))

}

func Setup() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     model.C.Gitlab.ClientID,
		ClientSecret: model.C.Gitlab.ClientSecret,
		RedirectURL:  model.C.Gitlab.RedirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  model.C.Gitlab.AuthURL,
			TokenURL: model.C.Gitlab.TokenURL,
		},
	}
}

func HandleCallback(c yee.Context) (err error) {
	c.Logger().SetLevel(4)
	conf := Setup()
	ctx := context.Background()
	g := new(Config)
	if err = c.Bind(g); err != nil {
		return c.JSON(http.StatusOK, commom.ERR_REQ_BIND)
	}

	if g.State != oauthStateString {
		c.Logger().Error("invalid oauth state")
		return c.JSON(http.StatusInternalServerError, commom.ERR_COMMON_MESSAGE(errors.New("invalid oauth state")))
	}

	// Use the authorization code that is pushed to the redirect
	// URL. Exchange will do the handshake to retrieve the
	// initial access token. The HTTP Client returned by
	// conf.Client will refresh the token as necessary.
	tok, err := conf.Exchange(ctx, g.Code)

	if err != nil {
		c.Logger().Error(err.Error())
		return c.JSON(http.StatusInternalServerError, commom.ERR_COMMON_MESSAGE(err))
	}

	gu, _ := getGitlabUserInfo(c, tok.AccessToken)
	cu, err := gitlabUserRegister(gu)

	token, tokenErr := lib.JwtAuth(cu.Username, cu.Rule)
	if tokenErr != nil {
		c.Logger().Error(tokenErr.Error())
		return c.JSON(http.StatusInternalServerError, commom.ERR_COMMON_MESSAGE(err))
	}

	dataStore := map[string]interface{}{
		"token":       token,
		"real_name":   gu.Name,
		"user":        gu.Username,
		"permissions": cu.Rule,
	}

	// 若用户存在，rule设置为当前rule
	// 否则默认为 guest
	if errors.Is(err, lib.ErrExist) {
		dataStore["rule"] = cu.Rule
		c.Logger().Info(fmt.Sprintf("%s 用户已存在请重新注册!", cu.Username))
	} else {
		dataStore["rule"] = "guest"
	}

	return c.JSON(http.StatusOK, commom.SuccessPayload(dataStore))
}

func UserGeneralLogin(c yee.Context) (err error) {
	u := new(loginForm)
	if err = c.Bind(u); err != nil {
		c.Logger().Error(err.Error())
		return c.JSON(http.StatusOK, commom.ERR_REQ_BIND)
	}
	var account model.CoreAccount
	if !model.DB().Where("username = ?", u.Username).First(&account).RecordNotFound() {
		if account.Username != u.Username {
			return c.JSON(http.StatusOK, commom.ERR_LOGIN)
		}
		if e := lib.DjangoCheckPassword(&account, u.Password); e {
			token, tokenErr := lib.JwtAuth(u.Username, account.Rule)
			if tokenErr != nil {
				c.Logger().Error(tokenErr.Error())
				return
			}
			dataStore := map[string]string{
				"token":       token,
				"permissions": account.Rule,
				"real_name":   account.RealName,
			}
			return c.JSON(http.StatusOK, commom.SuccessPayload(dataStore))
		}

	}
	return c.JSON(http.StatusOK, commom.ERR_LOGIN)

}

func UserRegister(c yee.Context) (err error) {

	if model.GloOther.Register {
		u := new(model.CoreAccount)
		if err = c.Bind(u); err != nil {
			c.Logger().Error(err.Error())
			return c.JSON(http.StatusOK, commom.ERR_REQ_BIND)
		}
		var unique model.CoreAccount
		ix, _ := json.Marshal([]string{})
		model.DB().Where("username = ?", u.Username).Select("username").First(&unique)
		if unique.Username != "" {
			return c.JSON(http.StatusOK, commom.ERR_COMMON_MESSAGE(errors.New("用户已存在请重新注册！")))
		}
		model.DB().Create(&model.CoreAccount{
			Username:   u.Username,
			RealName:   u.RealName,
			Password:   lib.DjangoEncrypt(u.Password, string(lib.GetRandom())),
			Rule:       "guest",
			Department: u.Department,
			Email:      u.Email,
		})
		model.DB().Create(&model.CoreGrained{Username: u.Username, Group: ix})
		return c.JSON(http.StatusOK, commom.SuccessPayLoadToMessage("注册成功！"))
	}
	return c.JSON(http.StatusOK, commom.ERR_REGISTER)

}

func gitlabUserRegister(user *GitlabUser) (cu *model.CoreAccount, err error) {
	u := new(model.CoreAccount)

	u.Username = user.Username
	u.RealName = user.Name
	u.Email = user.Email

	var unique model.CoreAccount
	ix, _ := json.Marshal([]string{})
	if !model.DB().Where("username = ?", u.Username).First(&unique).RecordNotFound() {
		return &unique, lib.ErrExist
	}
	model.DB().Create(&model.CoreAccount{
		Username: u.Username,
		RealName: u.RealName,
		Rule:     "guest",
		Email:    u.Email,
	})
	model.DB().Create(&model.CoreGrained{Username: u.Username, Group: ix})
	return u, nil

}

func getGitlabUserInfo(c yee.Context, t string) (user *GitlabUser, err error) {
	url := fmt.Sprintf("%s/api/v4/user?access_token=%s", model.GloGitlab, t)
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		c.Logger().Error(err.Error())
		return
	}

	res, err := client.Do(req)
	if err != nil {
		c.Logger().Error(err.Error())
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		c.Logger().Error(err.Error())
		return
	}

	var u *GitlabUser

	if err = json.Unmarshal(body, &u); err != nil {
		c.Logger().Error("Unmarshal Gitlab userinfo failed!")
		return nil, err
	}

	return u, nil

}
