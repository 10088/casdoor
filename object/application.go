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

package object

import (
	"github.com/casdoor/casdoor/util"
	"xorm.io/core"
)

type Application struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	DisplayName     string        `xorm:"varchar(100)" json:"displayName"`
	Logo            string        `xorm:"varchar(100)" json:"logo"`
	HomepageUrl     string        `xorm:"varchar(100)" json:"homepageUrl"`
	Description     string        `xorm:"varchar(100)" json:"description"`
	Organization    string        `xorm:"varchar(100)" json:"organization"`
	EnablePassword  bool          `json:"enablePassword"`
	EnableSignUp    bool          `json:"enableSignUp"`
	Providers       []string      `xorm:"varchar(1000)" json:"providers"`
	ProviderObjs    []*Provider   `xorm:"-" json:"providerObjs"`
	OrganizationObj *Organization `xorm:"-" json:"organizationObj"`

	ClientId      string   `xorm:"varchar(100)" json:"clientId"`
	ClientSecret  string   `xorm:"varchar(100)" json:"clientSecret"`
	RedirectUris  []string `xorm:"varchar(1000)" json:"redirectUris"`
	ExpireInHours int      `json:"expireInHours"`
	SignupUrl     string   `xorm:"varchar(100)" json:"signupUrl"`
	ForgetUrl     string   `xorm:"varchar(100)" json:"forgetUrl"`
}

func GetApplications(owner string) []*Application {
	applications := []*Application{}
	err := adapter.Engine.Desc("created_time").Find(&applications, &Application{Owner: owner})
	if err != nil {
		panic(err)
	}

	return applications
}

func extendApplicationWithProviders(application *Application) {
	providers := GetProviders(application.Owner)
	m := map[string]*Provider{}
	for _, provider := range providers {
		provider.ClientSecret = ""
		provider.ProviderUrl = ""
		m[provider.Name] = provider
	}

	application.ProviderObjs = []*Provider{}
	for _, providerName := range application.Providers {
		application.ProviderObjs = append(application.ProviderObjs, m[providerName])
	}
}

func extendApplicationWithOrg(application *Application) {
	organization := getOrganization(application.Owner, application.Organization)
	application.OrganizationObj = organization
}

func getApplication(owner string, name string) *Application {
	application := Application{Owner: owner, Name: name}
	existed, err := adapter.Engine.Get(&application)
	if err != nil {
		panic(err)
	}

	if existed {
		extendApplicationWithProviders(&application)
		extendApplicationWithOrg(&application)
		return &application
	} else {
		return nil
	}
}

func GetDefaultApplication(owner string) *Application {
	name := "app-built-in"
	application := Application{Owner: owner, Name: name}
	existed, err := adapter.Engine.Get(&application)
	if err != nil {
		panic(err)
	}

	if existed {
		extendApplicationWithProviders(&application)
		extendApplicationWithOrg(&application)
		return &application
	} else {
		return nil
	}
}

func getApplicationByClientId(clientId string) *Application {
	application := Application{}
	existed, err := adapter.Engine.Where("client_id=?", clientId).Get(&application)
	if err != nil {
		panic(err)
	}

	if existed {
		extendApplicationWithProviders(&application)
		extendApplicationWithOrg(&application)
		return &application
	} else {
		return nil
	}
}

func GetApplication(id string) *Application {
	owner, name := util.GetOwnerAndNameFromId(id)
	return getApplication(owner, name)
}

func UpdateApplication(id string, application *Application) bool {
	owner, name := util.GetOwnerAndNameFromId(id)
	if getApplication(owner, name) == nil {
		return false
	}

	affected, err := adapter.Engine.ID(core.PK{owner, name}).AllCols().Update(application)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func AddApplication(application *Application) bool {
	application.ClientId = util.GenerateClientId()
	application.ClientSecret = util.GenerateClientSecret()

	affected, err := adapter.Engine.Insert(application)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func DeleteApplication(application *Application) bool {
	affected, err := adapter.Engine.ID(core.PK{application.Owner, application.Name}).Delete(&Application{})
	if err != nil {
		panic(err)
	}

	return affected != 0
}
