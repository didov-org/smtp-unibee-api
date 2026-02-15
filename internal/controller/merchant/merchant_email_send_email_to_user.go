package merchant

import (
	"context"
	"fmt"
	"strings"
	"time"
	_interface "unibee/internal/interface/context"
	email2 "unibee/internal/logic/email"
	"unibee/internal/logic/invoice/handler"
	"unibee/internal/logic/merchant_config"
	"unibee/internal/query"
	"unibee/utility"

	"unibee/api/merchant/email"
)

func (c *ControllerEmail) SendEmailToUser(ctx context.Context, req *email.SendEmailToUserReq) (res *email.SendEmailToUserRes, err error) {
	user := query.GetUserAccountByEmail(ctx, _interface.GetMerchantId(ctx), req.Email)
	utility.Assert(user != nil, "User not found")
	mailTo := strings.ToLower(req.Email)
	var gatewayName, emailGatewayKey string
	if len(req.GatewayName) > 0 {
		utility.Assert(req.GatewayName == "sendgrid" || req.GatewayName == "smtp",
			"gatewayName must be 'sendgrid' or 'smtp'")
		gatewayName = req.GatewayName
		gwConfig := merchant_config.GetMerchantConfig(ctx, _interface.GetMerchantId(ctx), req.GatewayName)
		utility.Assert(gwConfig != nil && len(gwConfig.ConfigValue) > 0,
			fmt.Sprintf("email gateway '%s' has no saved configuration", req.GatewayName))
		emailGatewayKey = gwConfig.ConfigValue
	} else {
		gatewayName, emailGatewayKey = email2.GetDefaultMerchantEmailConfigWithClusterCloud(ctx, _interface.GetMerchantId(ctx))
	}
	if len(emailGatewayKey) == 0 {
		utility.Assert(false, "Default Email Gateway Need Setup")
	}
	var pdfFileName string
	var attachName string
	if len(req.AttachInvoiceId) == 0 && req.Variables != nil && req.Variables["AttachInvoiceId"] != nil {
		req.AttachInvoiceId = fmt.Sprintf("%s", req.Variables["AttachInvoiceId"])
	}
	if len(req.AttachInvoiceId) > 0 {
		one := query.GetInvoiceByInvoiceId(ctx, req.AttachInvoiceId)
		utility.Assert(one != nil, "invoice not found")
		utility.Assert(one.UserId > 0 && one.UserId == user.Id, "invoice userId not match")
		pdfFileName = handler.GenerateInvoicePdf(ctx, one)
		attachName = fmt.Sprintf("invoice_%s", time.Now().Format("20060102"))
	}
	if len(attachName) > 0 {
		attachName = attachName + ".pdf"
	}
	err = email2.Send(ctx, &email2.EmailSendReq{
		MerchantId:        _interface.GetMerchantId(ctx),
		MailTo:            mailTo,
		Subject:           req.Subject,
		Content:           req.Content,
		LocalFilePath:     pdfFileName,
		AttachName:        attachName,
		APIKey:            emailGatewayKey,
		GatewayName:       gatewayName,
		VariableMap:       req.Variables,
		Language:          user.Language,
		GatewayTemplateId: req.GatewayTemplateId,
	})
	if err != nil {
		return nil, err
	}
	return &email.SendEmailToUserRes{}, nil
}
