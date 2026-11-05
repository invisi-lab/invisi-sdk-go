package session

import (
	"fmt"
	"strconv"

	"github.com/invisi-lab/invisi-sdk-go/protocol"
)

func (s *WalletSession) Scope() WalletScope { return s.scope }

func (s *WalletSession) Protocol() *protocol.Protocol { return s.protocol }

func (s *WalletSession) String() string {
	return fmt.Sprintf("WalletSession(%s,%s)", strconv.FormatUint(s.scope.ChainID, 10), s.wallet.OwnerAddress().Hex())
}
