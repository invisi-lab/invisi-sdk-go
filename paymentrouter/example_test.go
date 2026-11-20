package paymentrouter_test

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/invisi-labs/invisi-sdk-go/paymentrouter"
)

func ExampleEncodeV3ExactOutputPath() {
	weth := common.HexToAddress("0x4444444444444444444444444444444444444444")
	token := common.HexToAddress("0x2222222222222222222222222222222222222222")
	path, err := paymentrouter.EncodeV3ExactOutputPath([]common.Address{weth, token}, []uint32{3000})
	if err != nil {
		panic(err)
	}
	fmt.Println(hexutil.Encode(path))
	// Output: 0x2222222222222222222222222222222222222222000bb84444444444444444444444444444444444444444
}
