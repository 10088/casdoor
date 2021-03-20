// Copyright 2021 The casbin Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controllers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/astaxie/beego"
	"github.com/casdoor/casdoor/idp"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
	"golang.org/x/oauth2"
)

func (c *ApiController) HandleLoggedIn(userId string, form *RequestForm) *Response {
	resp := &Response{}
	if form.Type == ResponseTypeLogin {
		c.SetSessionUser(userId)
		util.LogInfo(c.Ctx, "API: [%s] signed in", userId)
		resp = nil
	} else if form.Type == ResponseTypeCode {
		clientId := c.Input().Get("clientId")
		responseType := c.Input().Get("responseType")
		redirectUri := c.Input().Get("redirectUri")
		scope := c.Input().Get("scope")
		state := c.Input().Get("state")

		code := object.GetOAuthCode(userId, clientId, responseType, redirectUri, scope, state)
		resp = codeToResponse(code)
	} else {
		resp = &Response{Status: "error", Msg: fmt.Sprintf("unknown response type: %s", form.Type)}
	}
	return resp
}

func (c *ApiController) GetApplicationLogin() {
	var resp Response

	clientId := c.Input().Get("clientId")
	responseType := c.Input().Get("responseType")
	redirectUri := c.Input().Get("redirectUri")
	scope := c.Input().Get("scope")
	state := c.Input().Get("state")

	msg, application := object.CheckOAuthLogin(clientId, responseType, redirectUri, scope, state)
	if msg != "" {
		resp = Response{Status: "error", Msg: msg, Data: application}
	} else {
		resp = Response{Status: "ok", Msg: "", Data: application}
	}
	c.Data["json"] = resp
	c.ServeJSON()
}

func (c *ApiController) Login() {
	var resp Response
	var form RequestForm
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &form)
	if err != nil {
		panic(err)
	}

	if form.Username != "" {
		if c.GetSessionUser() != "" {
			resp = Response{Status: "error", Msg: "please log out first before signing in", Data: c.GetSessionUser()}
			c.Data["json"] = resp
			c.ServeJSON()
			return
		}

		userId := fmt.Sprintf("%s/%s", form.Organization, form.Username)
		password := form.Password
		msg := object.CheckUserLogin(userId, password)

		if msg != "" {
			resp = Response{Status: "error", Msg: msg, Data: ""}
		} else {
			c.HandleLoggedIn(userId, &form)
			resp = Response{Status: "ok", Msg: "", Data: userId}
		}
	} else if form.Provider != "" {
		application := object.GetApplication(fmt.Sprintf("admin/%s", form.Application))
		provider := object.GetProvider(fmt.Sprintf("admin/%s", form.Provider))

		idProvider := idp.GetIdProvider(provider.Type)
		oauthConfig := idProvider.GetConfig()
		oauthConfig.ClientID = provider.ClientId
		oauthConfig.ClientSecret = provider.ClientSecret
		oauthConfig.RedirectURL = form.RedirectUri

		var res authResponse

		if form.State != beego.AppConfig.String("AuthState") {
			resp = Response{Status: "error", Msg: "unauthorized", Data: res}
			c.ServeJSON()
			return
		}

		// https://github.com/golang/oauth2/issues/123#issuecomment-103715338
		ctx := context.WithValue(oauth2.NoContext, oauth2.HTTPClient, httpClient)
		token, err := oauthConfig.Exchange(ctx, form.Code)
		if err != nil {
			panic(err)
		}

		if !token.Valid() {
			resp = Response{Status: "error", Msg: "unauthorized", Data: res}
			c.Data["json"] = resp
			c.ServeJSON()
			return
		}

		res.Email, res.Method, res.Avatar, err = idProvider.GetUserInfo(httpClient, token)
		if err != nil {
			resp = Response{Status: "error", Msg: "Login failed, please try again."}
			c.Data["json"] = resp
			c.ServeJSON()
			return
		}

		if form.Method == "signup" {
			userId := ""
			if provider.Type == "github" {
				userId = object.GetUserIdByField(application, "github", res.Method)
			} else if provider.Type == "google" {
				userId = object.GetUserIdByField(application, "google", res.Email)
			}

			if userId != "" {
				//if object.IsForbidden(userId) {
				//	c.forbiddenAccountResp(userId)
				//	return
				//}

				//if len(object.GetMemberAvatar(userId)) == 0 {
				//	avatar := UploadAvatarToOSS(res.Avatar, userId)
				//	object.LinkMemberAccount(userId, "avatar", avatar)
				//}

				c.HandleLoggedIn(userId, &form)
			} else {
				//if object.IsForbidden(userId) {
				//	c.forbiddenAccountResp(userId)
				//	return
				//}

				if userId := object.GetUserIdByField(application, "email", res.Email); userId != "" {
					c.HandleLoggedIn(userId, &form)

					if provider.Type == "github" {
						_ = object.LinkUserAccount(userId, "github", res.Method)
					} else if provider.Type == "google" {
						_ = object.LinkUserAccount(userId, "google", res.Email)
					}
				}
			}
			resp = Response{Status: "ok", Msg: "success", Data: res}
		} else {
			userId := c.GetSessionUser()
			if userId == "" {
				resp = Response{Status: "error", Msg: "user doesn't exist", Data: res}
				c.Data["json"] = resp
				c.ServeJSON()
				return
			}

			linkRes := false
			if provider.Type == "github" {
				linkRes = object.LinkUserAccount(userId, "github", res.Method)
			} else if provider.Type == "google" {
				linkRes = object.LinkUserAccount(userId, "google", res.Email)
			}
			if linkRes {
				resp = Response{Status: "ok", Msg: "success", Data: linkRes}
			} else {
				resp = Response{Status: "error", Msg: "link account failed", Data: linkRes}
			}
			//if len(object.GetMemberAvatar(userId)) == 0 {
			//	avatar := UploadAvatarToOSS(tempUserAccount.AvatarUrl, userId)
			//	object.LinkUserAccount(userId, "avatar", avatar)
			//}
		}
	} else {
		panic("unknown authentication type (not password or provider)")
	}

	c.Data["json"] = resp
	c.ServeJSON()
}
