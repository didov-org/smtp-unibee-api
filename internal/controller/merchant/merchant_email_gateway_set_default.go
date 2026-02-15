package merchant

import (
	"context"
	"fmt"
	"unibee/api/merchant/email"
	_interface "unibee/internal/interface/context"
	"unibee/internal/logic/merchant_config"
	"unibee/internal/logic/merchant_config/update"
	"unibee/internal/logic/operation_log"
	"unibee/utility"

	email2 "unibee/internal/logic/email"
)

func (c *ControllerEmail) GatewaySetDefault(ctx context.Context, req *email.GatewaySetDefaultReq) (res *email.GatewaySetDefaultRes, err error) {
	utility.Assert(req.GatewayName == "sendgrid" || req.GatewayName == "smtp", "gatewayName must be 'sendgrid' or 'smtp'")
	merchantId := _interface.GetMerchantId(ctx)
	gwConfig := merchant_config.GetMerchantConfig(ctx, merchantId, req.GatewayName)
	utility.Assert(gwConfig != nil && len(gwConfig.ConfigValue) > 0,
		fmt.Sprintf("email gateway '%s' has no saved configuration, set it up first", req.GatewayName))
	err = update.SetMerchantConfig(ctx, merchantId, email2.KeyMerchantEmailName, req.GatewayName)
	operation_log.AppendOptLog(ctx, &operation_log.OptLogRequest{
		MerchantId:     merchantId,
		Target:         fmt.Sprintf("EmailGateway(%s)", req.GatewayName),
		Content:        "SetDefaultEmailGateway",
		UserId:         0,
		SubscriptionId: "",
		InvoiceId:      "",
		PlanId:         0,
		DiscountCode:   "",
	}, err)
	if err != nil {
		return nil, err
	}
	return &email.GatewaySetDefaultRes{}, nil
}
