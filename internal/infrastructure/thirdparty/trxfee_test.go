package thirdparty

import (
	"testing"
)

func TestActivationAddress(t *testing.T) {
	trxfee := NewTrxfeeClient("https://trxfee.io/", "CC4F20ACDB45AFA10A22D6BDA2AE9F3F", "99144B2AC7ED7F73ECFF59144D46E321F1DC83B373DE2FF6A367423F4CF61FB5")
	trxfee.Activation("TBCG8qr7TSLZqYLYsf8UB3uoKSMWJ9qo94")
}
