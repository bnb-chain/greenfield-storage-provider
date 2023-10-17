package fix_v1_0_0

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/forbole/juno/v4/models"
	"gorm.io/gorm"

	"github.com/bnb-chain/greenfield-storage-provider/pkg/log"
)

type Account struct {
	Addr       string `json:"addr"`
	Owner      string `json:"owner"`
	Refundable bool   `json:"refundable"`
}

type PaymentResult struct {
	PaymentAccount Account `json:"payment_account"`
}

func FixPayment(endpoint string, db *gorm.DB) error {
	log.Infof("job start")

	url := endpoint + "/greenfield/payment/payment_account/%s"
	client := http.Client{}

	var results []*models.PaymentAccount
	if err := db.Table("payment_accounts").Find(&results).Error; err != nil {
		return err
	}
	for _, res := range results {
		var paymentResult PaymentResult
		var httpErr error
		var resp *http.Response
		func() {
			resp, httpErr = client.Get(fmt.Sprintf(url, res.Addr))
			if httpErr != nil {
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				httpErr = fmt.Errorf(resp.Status)
				return
			}
			err := json.NewDecoder(resp.Body).Decode(&paymentResult)
			if err != nil {
				httpErr = err
				return
			}
		}()
		if httpErr != nil {
			log.Errorw("failed to curl chain", "error", httpErr)
			return httpErr
		}
		if paymentResult.PaymentAccount.Refundable == res.Refundable {
			continue
		}
		if err := db.Table("payment_accounts").Where("addr = ?", res.Addr).Update("refundable", paymentResult.PaymentAccount.Refundable).Error; err != nil {
			return err
		}
	}
	return nil
}
