package merchant

import (
	"context"
	"strings"
	"unibee/api/merchant/email"
	_interface "unibee/internal/interface/context"
	email2 "unibee/internal/logic/email"
	"unibee/internal/logic/email/gateway"
	"unibee/utility"
)

func (c *ControllerEmail) GatewaySetupV2(ctx context.Context, req *email.GatewaySetupV2Req) (res *email.GatewaySetupV2Res, err error) {
	utility.Assert(req.GatewayName == "sendgrid" || req.GatewayName == "smtp", "gatewayName must be 'sendgrid' or 'smtp'")
	utility.Assert(req.ApiCredential != nil, "apiCredential is required")

	var data string
	if req.GatewayName == "sendgrid" {
		utility.Assert(len(req.ApiCredential.ApiKey) > 0, "apiKey is required for sendgrid")
		data = req.ApiCredential.ApiKey
	} else {
		smtpHost := strings.TrimSpace(req.ApiCredential.SmtpHost)
		utility.Assert(len(smtpHost) > 0, "smtpHost is required for smtp")
		utility.Assert(utility.ValidateExternalHost(smtpHost) == nil, "smtpHost must be a valid external SMTP server")
		utility.Assert(req.ApiCredential.SmtpPort > 0 && req.ApiCredential.SmtpPort <= 65535, "smtpPort must be between 1 and 65535")
		authType := req.ApiCredential.AuthType
		if authType == "" {
			authType = "plain"
		}
		switch authType {
		case "plain", "cram-md5", "login":
			utility.Assert(len(req.ApiCredential.Username) > 0, "username is required for smtp with "+authType+" auth")
			utility.Assert(len(req.ApiCredential.Password) > 0, "password is required for smtp with "+authType+" auth")
		case "xoauth2":
			utility.Assert(len(req.ApiCredential.Username) > 0, "username is required for smtp with xoauth2 auth")
			utility.Assert(len(req.ApiCredential.OAuthToken) > 0, "oauthToken is required for smtp with xoauth2 auth")
		case "none":
			utility.Assert(false, "authType 'none' is not supported; SMTP requires authentication")
		default:
			utility.Assert(false, "unsupported authType: "+authType)
		}
		smtpCfg := gateway.SmtpConfig{
			SmtpHost:      smtpHost,
			SmtpPort:      req.ApiCredential.SmtpPort,
			Username:      req.ApiCredential.Username,
			Password:      req.ApiCredential.Password,
			UseTLS:        req.ApiCredential.UseTLS,
			SkipTLSVerify: false,
			AuthType:      authType,
			OAuthToken:    req.ApiCredential.OAuthToken,
		}
		data = utility.MarshalToJsonString(smtpCfg)
	}

	err = email2.SetupMerchantEmailConfig(ctx, _interface.GetMerchantId(ctx), req.GatewayName, data, req.IsDefault)
	if err != nil {
		return nil, err
	}
	return &email.GatewaySetupV2Res{Data: utility.HideStar(data)}, nil
}
