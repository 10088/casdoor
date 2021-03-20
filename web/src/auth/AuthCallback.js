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

import React from "react";
import {Spin} from "antd";
import {withRouter} from "react-router-dom";
import * as AuthBackend from "./AuthBackend";
import * as Util from "./Util";
import {authConfig} from "./Auth";

class AuthCallback extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      applicationName: props.match.params.applicationName,
      providerName: props.match.params.providerName,
      method: props.match.params.method,
    };
  }

  getResponseType() {
    // "http://localhost:8000"
    const authServerUrl = authConfig.serverUrl;

    // For example, for Casbin-OA, realRedirectUri = "http://localhost:9000/login"
    // realRedirectUrl = "http://localhost:9000"
    const params = new URLSearchParams(this.props.location.search);
    const realRedirectUri = params.get("redirect_uri");
    const realRedirectUrl = new URL(realRedirectUri).origin;

    // For Casdoor itself, we use "login" directly
    if (authServerUrl === realRedirectUrl) {
      return "login";
    } else {
      return "code";
    }
  }

  componentWillMount() {
    const params = new URLSearchParams(this.props.location.search);
    let redirectUri = `${window.location.origin}/callback/${this.state.applicationName}/${this.state.providerName}/${this.state.method}`;
    const body = {
      type: this.getResponseType(),
      application: this.state.applicationName,
      provider: this.state.providerName,
      code: params.get("code"),
      state: params.get("state"),
      redirectUri: redirectUri,
      method: this.state.method,
    };
    const oAuthParams = Util.getOAuthGetParameters();
    AuthBackend.login(body, oAuthParams)
      .then((res) => {
        if (res.status === 'ok') {
          const responseType = this.getResponseType();
          if (responseType === "login") {
            // this.props.onLoggedIn();
            Util.showMessage("success", `Logged in successfully`);
            Util.goToLink("/");
          } else if (responseType === "code") {
            const code = res.data;
            Util.goToLink(`${oAuthParams.redirectUri}?code=${code}&state=${oAuthParams.state}`);
            // Util.showMessage("success", `Authorization code: ${res.data}`);
          }
        } else {
          Util.showMessage("error", `Log in failed：${res.msg}`);
        }
      });
  }

  render() {
    return (
      <div style={{textAlign: "center"}}>
        <Spin size="large" tip="Signing in..." style={{paddingTop: "10%"}} />
      </div>
    )
  }
}

export default withRouter(AuthCallback);
