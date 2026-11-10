package session

import (
	"fmt"
	"strconv"

	"github.com/invisi-lab/invisi-sdk-go/rpc/chain"
)

func (s *WalletSession) Scope() WalletScope { return s.scope }

func (s *WalletSession) Protocol() *chain.Protocol { return s.protocol }

func (s *WalletSession) String() string {
	return fmt.Sprintf("WalletSession(%s,%s)", strconv.FormatUint(s.scope.ChainID, 10), s.wallet.OwnerAddress().Hex())
}
